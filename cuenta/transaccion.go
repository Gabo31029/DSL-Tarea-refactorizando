/*
source file: transaccion.go
US-04: integridad de la transaccion + log en JSON
en la version web se agrega el mutex porque llegan varias peticiones a la vez
*/
package cuenta

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	Exitosa   = "EXITOSA"
	Rechazada = "RECHAZADA"
	Error     = "ERROR"
)

type RegistroCuenta struct {
	ID     string  `json:"id"`
	Nombre string  `json:"nombre"`
	Saldo  float64 `json:"saldo"`
	Activa bool    `json:"activa"`
}

// cada transaccion se guarda como una linea del log
type Transaccion struct {
	Fecha        string  `json:"fecha"`
	CuentaID     string  `json:"cuenta_id"`
	Tipo         string  `json:"tipo"`
	Monto        float64 `json:"monto"`
	Estado       string  `json:"estado"`
	SaldoAntes   float64 `json:"saldo_antes"`
	SaldoDespues float64 `json:"saldo_despues"`
	Mensaje      string  `json:"mensaje"`
}

type Banco struct {
	mu        sync.Mutex
	cuentas   map[string]*RegistroCuenta
	orden     []string
	archDatos string
	archLog   string
}

func CargarBanco(archDatos, archLog string) (*Banco, error) {
	data, err := os.ReadFile(archDatos)
	if err != nil {
		return nil, err
	}
	var lista []RegistroCuenta
	if err := json.Unmarshal(data, &lista); err != nil {
		return nil, fmt.Errorf("%s: %v", archDatos, err)
	}
	b := &Banco{cuentas: map[string]*RegistroCuenta{}, archDatos: archDatos, archLog: archLog}
	for i := range lista {
		id := lista[i].ID
		if _, existe := b.cuentas[id]; existe {
			return nil, fmt.Errorf("%s: la cuenta %s esta repetida", archDatos, id)
		}
		b.cuentas[id] = &lista[i]
		b.orden = append(b.orden, id)
	}
	return b, nil
}

func (b *Banco) Cuenta(id string) (RegistroCuenta, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	c, ok := b.cuentas[id]
	if !ok {
		return RegistroCuenta{}, false
	}
	return *c, true
}

func (b *Banco) Cuentas() []RegistroCuenta {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.listar()
}

func (b *Banco) listar() []RegistroCuenta {
	lista := make([]RegistroCuenta, 0, len(b.orden))
	for _, id := range b.orden {
		lista = append(lista, *b.cuentas[id])
	}
	return lista
}

// Procesar hace un RETIRO o ABONO completo o no hace nada (RN-03).
// La operacion se prueba en una copia de la cuenta y el saldo real solo
// cambia si se pudo guardar el archivo. Todo intento queda en el log.
func (b *Banco) Procesar(id, tipo string, monto float64) Transaccion {
	b.mu.Lock()
	defer b.mu.Unlock()
	tx := Transaccion{Fecha: time.Now().Format(time.RFC3339), CuentaID: id, Tipo: tipo, Monto: monto}
	if math.IsNaN(monto) || math.IsInf(monto, 0) {
		tx.Monto = 0 // JSON no acepta NaN ni Inf
	}

	reg, ok := b.cuentas[id]
	switch {
	case !ok:
		tx.Estado, tx.Mensaje = Error, "cuenta no existe"
	case !reg.Activa:
		tx.Estado, tx.Mensaje = Error, "cuenta inactiva"
		tx.SaldoAntes, tx.SaldoDespues = reg.Saldo, reg.Saldo
	default:
		tx.SaldoAntes = reg.Saldo
		copia := NewCuenta(reg.Nombre, reg.Saldo)

		var exito bool
		var err error
		switch tipo {
		case "RETIRO":
			exito, err = copia.Retirar(monto)
		case "ABONO":
			exito, err = copia.Abonar(monto)
		default:
			err = errors.New("tipo de transaccion no valido")
		}

		switch {
		case err != nil:
			tx.Estado, tx.Mensaje = Error, err.Error()
		case !exito:
			tx.Estado, tx.Mensaje = Rechazada, "Saldo insuficiente. Intente con un monto menor."
		default:
			anterior := reg.Saldo
			reg.Saldo = copia.ObtenerSaldo()
			if err := b.guardar(); err != nil {
				reg.Saldo = anterior // no se guardo, se deja como estaba
				tx.Estado, tx.Mensaje = Error, "no se pudo guardar: "+err.Error()
			} else {
				tx.Estado, tx.Mensaje = Exitosa, tipo+" procesado"
			}
		}
		tx.SaldoDespues = reg.Saldo
	}

	if err := b.registrar(tx); err != nil {
		fmt.Fprintln(os.Stderr, "aviso: no se pudo escribir el log:", err)
	}
	return tx
}

// se escribe primero a un .tmp y luego se renombra, asi el archivo
// nunca queda a medio escribir
func (b *Banco) guardar() error {
	data, err := json.MarshalIndent(b.listar(), "", "  ")
	if err != nil {
		return err
	}
	tmp := b.archDatos + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, b.archDatos)
}

func (b *Banco) registrar(tx Transaccion) error {
	if b.archLog == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(b.archLog), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(b.archLog, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	linea, err := json.Marshal(tx)
	if err != nil {
		return err
	}
	_, err = f.Write(append(linea, '\n'))
	return err
}

// Historial lee el log para mostrarlo en la pagina
func (b *Banco) Historial() []Transaccion {
	b.mu.Lock()
	defer b.mu.Unlock()
	lista := []Transaccion{}
	f, err := os.Open(b.archLog)
	if err != nil {
		return lista
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		var tx Transaccion
		if json.Unmarshal(sc.Bytes(), &tx) == nil {
			lista = append(lista, tx)
		}
	}
	return lista
}
