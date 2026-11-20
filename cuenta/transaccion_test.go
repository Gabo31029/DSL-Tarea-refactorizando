// pruebas de integridad-transaccion.feature (US-04)
package cuenta

import (
	"bufio"
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"testing"
)

// mismos datos iniciales que data/cuentas.json
const cuentasPrueba = `[
  {"id":"C001","nombre":"Alice","saldo":234.56,"activa":true},
  {"id":"C002","nombre":"Hamza","saldo":198.76,"activa":true},
  {"id":"C003","nombre":"Pedro","saldo":300.00,"activa":true},
  {"id":"C004","nombre":"Maria","saldo":50.00,"activa":false}
]`

// cada prueba trabaja en una carpeta temporal para no tocar data/cuentas.json
func bancoDePrueba(t *testing.T) (*Banco, string) {
	dir := t.TempDir()
	datos := filepath.Join(dir, "cuentas.json")
	os.WriteFile(datos, []byte(cuentasPrueba), 0o644)
	log := filepath.Join(dir, "transacciones.json")
	b, err := CargarBanco(datos, log)
	if err != nil {
		t.Fatal(err)
	}
	return b, log
}

func ultimoLog(t *testing.T, path string) Transaccion {
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var tx Transaccion
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		json.Unmarshal(sc.Bytes(), &tx)
	}
	return tx
}

func TestIntegridadTransaccion(t *testing.T) {
	casos := []struct {
		escenario, id, tipo string
		monto               float64
		estado              string
		saldoFinal          float64
	}{
		{"Successful transaction is logged", "C001", "RETIRO", 200.00, Exitosa, 34.56},
		{"Declined transaction is logged", "C002", "RETIRO", 200.00, Rechazada, 198.76},
		{"Error transaction is logged", "C003", "ABONO", -10.00, Error, 300.00},
		{"Inactive account", "C004", "RETIRO", 10.00, Error, 50.00},
		{"Successful deposit is logged", "C003", "ABONO", 25.50, Exitosa, 325.50},
		{"Tipo no valido", "C001", "TRANSFERENCIA", 10.00, Error, 234.56},
		{"Monto NaN", "C001", "RETIRO", math.NaN(), Error, 234.56},
	}
	for _, c := range casos {
		t.Run(c.escenario, func(t *testing.T) {
			b, log := bancoDePrueba(t)
			tx := b.Procesar(c.id, c.tipo, c.monto)
			if tx.Estado != c.estado {
				t.Errorf("estado = %s, esperado %s", tx.Estado, c.estado)
			}
			cta, _ := b.Cuenta(c.id)
			if !iguales(cta.Saldo, c.saldoFinal) {
				t.Errorf("saldo = %.2f, esperado %.2f", cta.Saldo, c.saldoFinal)
			}
			if ultimoLog(t, log).Estado != c.estado {
				t.Errorf("el log no registro el estado %s", c.estado)
			}
		})
	}
}

// si no se puede guardar el archivo, el saldo no debe cambiar
func TestAtomicidadRollback(t *testing.T) {
	b, _ := bancoDePrueba(t)
	b.archDatos = filepath.Join(t.TempDir(), "no-existe", "cuentas.json")
	tx := b.Procesar("C001", "RETIRO", 100)
	cta, _ := b.Cuenta("C001")
	if tx.Estado != Error || !iguales(cta.Saldo, 234.56) {
		t.Errorf("no hubo rollback: estado=%s saldo=%.2f", tx.Estado, cta.Saldo)
	}
}

func TestCuentaInexistente(t *testing.T) {
	b, _ := bancoDePrueba(t)
	if tx := b.Procesar("X999", "RETIRO", 10); tx.Estado != Error {
		t.Errorf("estado = %s, esperado ERROR", tx.Estado)
	}
}

func TestCuentaRepetida(t *testing.T) {
	datos := filepath.Join(t.TempDir(), "cuentas.json")
	os.WriteFile(datos, []byte(`[{"id":"C001","saldo":10,"activa":true},{"id":"C001","saldo":20,"activa":true}]`), 0o644)
	if _, err := CargarBanco(datos, ""); err == nil {
		t.Error("deberia fallar con una cuenta repetida")
	}
}

func TestGuardaEnArchivo(t *testing.T) {
	b, _ := bancoDePrueba(t)
	b.Procesar("C002", "ABONO", 1.24)
	otro, err := CargarBanco(b.archDatos, "")
	if err != nil {
		t.Fatal(err)
	}
	if c, _ := otro.Cuenta("C002"); !iguales(c.Saldo, 200.00) {
		t.Errorf("saldo guardado = %.2f, esperado 200.00", c.Saldo)
	}
}
