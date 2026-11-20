// pruebas de los escenarios de retiro-cash.feature y abono-cash.feature
package cuenta

import (
	"math"
	"testing"
)

func TestRetiro(t *testing.T) {
	casos := []struct {
		escenario     string
		saldo, monto  float64
		exitoEsperado bool
		errorEsperado bool
		saldoEsperado float64
	}{
		{"Successful withdrawal within balance", 234.56, 200.00, true, false, 34.56},
		{"Declined withdrawal in excess of balance", 198.76, 200.00, false, false, 198.76},
		{"Invalid withdrawal amount (negativo)", 300.00, -20.00, false, true, 300.00},
		{"Invalid withdrawal amount (cero)", 300.00, 0, false, true, 300.00},
		{"Monto menor a un centimo", 300.00, 0.004, false, true, 300.00},
		{"Monto NaN", 300.00, math.NaN(), false, true, 300.00},
	}
	for _, c := range casos {
		t.Run(c.escenario, func(t *testing.T) {
			cuenta := NewCuenta("Test", c.saldo)
			exito, err := cuenta.Retirar(c.monto)
			if exito != c.exitoEsperado {
				t.Errorf("exito = %v, esperado %v", exito, c.exitoEsperado)
			}
			if (err != nil) != c.errorEsperado {
				t.Errorf("error = %v, esperado error=%v", err, c.errorEsperado)
			}
			if !iguales(cuenta.ObtenerSaldo(), c.saldoEsperado) {
				t.Errorf("saldo = %.2f, esperado %.2f", cuenta.ObtenerSaldo(), c.saldoEsperado)
			}
		})
	}
}

func TestAbono(t *testing.T) {
	casos := []struct {
		escenario     string
		saldo, monto  float64
		exitoEsperado bool
		saldoEsperado float64
	}{
		{"Successful cash deposit", 100.00, 50.50, true, 150.50},
		{"Rejected deposit with zero amount", 80.00, 0, false, 80.00},
		{"Rejected deposit with negative amount", 80.00, -10.00, false, 80.00},
		{"Abono con mas de 2 decimales", 80.00, 10.556, true, 90.56},
		{"Abono infinito", 80.00, math.Inf(1), false, 80.00},
	}
	for _, c := range casos {
		t.Run(c.escenario, func(t *testing.T) {
			cuenta := NewCuenta("Test", c.saldo)
			exito, err := cuenta.Abonar(c.monto)
			if exito != c.exitoEsperado {
				t.Errorf("exito = %v, esperado %v", exito, c.exitoEsperado)
			}
			if !c.exitoEsperado && err == nil {
				t.Errorf("se esperaba error de validación")
			}
			if !iguales(cuenta.ObtenerSaldo(), c.saldoEsperado) {
				t.Errorf("saldo = %.2f, esperado %.2f", cuenta.ObtenerSaldo(), c.saldoEsperado)
			}
		})
	}
}
