/*
source file: cuenta.go
entidad Cuenta (retiro y abono)
*/
package cuenta

import (
	"errors"
	"fmt"
	"math"
)

type Cuenta struct {
	saldo  float64
	nombre string
}

func NewCuenta(nombre string, saldoInicial float64) *Cuenta {
	return &Cuenta{nombre: nombre, saldo: redondear(saldoInicial)}
}

func (c *Cuenta) ObtenerSaldo() float64 { return c.saldo }

func (c *Cuenta) Nombre() string { return c.nombre }

// RN-01: no se puede retirar mas de lo que hay en la cuenta
func (c *Cuenta) Retirar(cantidad float64) (bool, error) {
	monto, err := validarMonto(cantidad, "retirar")
	if err != nil {
		return false, err
	}
	if monto > c.saldo {
		return false, nil
	}
	c.saldo = redondear(c.saldo - monto)
	return true, nil
}

// US-01b: abono en efectivo
func (c *Cuenta) Abonar(cantidad float64) (bool, error) {
	monto, err := validarMonto(cantidad, "abonar")
	if err != nil {
		return false, err
	}
	nuevo := redondear(c.saldo + monto)
	if math.IsInf(nuevo, 0) {
		return false, errors.New("monto fuera de rango")
	}
	c.saldo = nuevo
	return true, nil
}

func (c *Cuenta) String() string {
	return fmt.Sprintf("%s: %.2f", c.nombre, c.saldo)
}

// el monto se trabaja con 2 decimales, asi 0.001 no cuenta como un abono valido
func validarMonto(m float64, op string) (float64, error) {
	if math.IsNaN(m) || math.IsInf(m, 0) {
		return 0, errors.New("monto no valido")
	}
	m = redondear(m)
	if m <= 0 || math.IsInf(m, 0) {
		return 0, fmt.Errorf("el monto a %s debe ser mayor a 0", op)
	}
	return m, nil
}

func redondear(v float64) float64 { return math.Round(v*100) / 100 }
