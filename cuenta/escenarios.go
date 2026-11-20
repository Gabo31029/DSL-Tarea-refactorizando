/*
source file: escenarios.go
escenarios gherkin (given / when / then) que se corren desde el CLI
*/
package cuenta

import (
	"fmt"
	"math"
)

func imprimirSaldo(c *Cuenta) { fmt.Printf("Saldo de %.2f\n", c.ObtenerSaldo()) }

func iguales(a, b float64) bool { return math.Abs(a-b) < 0.001 }

// TestEscenario1Exito: Successful withdrawal within balance
func TestEscenario1Exito(nombre string, saldoInicial, cantidad float64) {
	fmt.Printf("--- Escenario 1: Successful withdrawal within balance (%s) ---\n", nombre)
	cuenta := NewCuenta(nombre, saldoInicial)
	fmt.Printf("Given: %s tiene %.2f en su cuenta\n", nombre, saldoInicial)
	fmt.Printf("When: %s intenta retirar %.2f\n", nombre, cantidad)
	exito, err := cuenta.Retirar(cantidad)
	switch {
	case err != nil:
		fmt.Printf("Resultado: FALLO. Error inesperado: %v\n\n", err)
	case exito && iguales(cuenta.ObtenerSaldo(), saldoInicial-cantidad):
		fmt.Println("Then: Resultado: ÉXITO. El retiro fue procesado.")
		imprimirSaldo(cuenta)
		fmt.Println()
	default:
		fmt.Println("Resultado: FALLO. El retiro fue rechazado (incorrecto).")
		fmt.Println()
	}
}

// TestEscenario2Fallo: Declined withdrawal in excess of balance
func TestEscenario2Fallo(nombre string, saldoInicial, cantidad float64) {
	fmt.Printf("--- Escenario 2: Declined withdrawal in excess of balance (%s) ---\n", nombre)
	cuenta := NewCuenta(nombre, saldoInicial)
	fmt.Printf("Given: %s tiene %.2f en su cuenta\n", nombre, saldoInicial)
	fmt.Printf("When: %s intenta retirar %.2f\n", nombre, cantidad)
	exito, err := cuenta.Retirar(cantidad)
	switch {
	case err != nil:
		fmt.Printf("Resultado: FALLO. Error inesperado: %v\n\n", err)
	case !exito && iguales(cuenta.ObtenerSaldo(), saldoInicial):
		fmt.Println("Then: Resultado: ÉXITO. El retiro fue rechazado correctamente.")
		fmt.Println("Mensaje: Saldo insuficiente. Intente con un monto menor.")
		imprimirSaldo(cuenta)
		fmt.Println()
	default:
		fmt.Println("Resultado: FALLO. El retiro fue procesado (debió rechazarse).")
		fmt.Println()
	}
}

// TestEscenario3Error: monto inválido (negativo o cero)
func TestEscenario3Error(nombre string, saldoInicial, cantidad float64) {
	fmt.Printf("--- Escenario 3: Validation Error (%s) ---\n", nombre)
	cuenta := NewCuenta(nombre, saldoInicial)
	_, err := cuenta.Retirar(cantidad)
	if err != nil {
		fmt.Printf("Then: Resultado: ÉXITO. El sistema atrapó el error: %v\n", err)
		imprimirSaldo(cuenta)
	} else {
		fmt.Println("Resultado: FALLO. No se detectó el error de validación.")
	}
	fmt.Println()
}

// TestEscenario4Abono: Successful cash deposit
func TestEscenario4Abono(nombre string, saldoInicial, cantidad float64) {
	fmt.Printf("--- Escenario 4: Successful cash deposit (%s) ---\n", nombre)
	cuenta := NewCuenta(nombre, saldoInicial)
	fmt.Printf("Given: %s tiene %.2f en su cuenta\n", nombre, saldoInicial)
	fmt.Printf("When: %s abona %.2f\n", nombre, cantidad)
	exito, err := cuenta.Abonar(cantidad)
	switch {
	case err != nil:
		fmt.Printf("Resultado: FALLO. Error inesperado: %v\n\n", err)
	case exito && iguales(cuenta.ObtenerSaldo(), saldoInicial+cantidad):
		fmt.Println("Then: Resultado: ÉXITO. El abono fue procesado.")
		imprimirSaldo(cuenta)
		fmt.Println()
	default:
		fmt.Println("Resultado: FALLO. El saldo no coincide.")
		fmt.Println()
	}
}

// TestEscenario5AbonoInvalido: Rejected deposit (cero o negativo)
func TestEscenario5AbonoInvalido(nombre string, saldoInicial, cantidad float64) {
	fmt.Printf("--- Escenario 5: Rejected deposit (%s) ---\n", nombre)
	cuenta := NewCuenta(nombre, saldoInicial)
	fmt.Printf("Given: %s tiene %.2f en su cuenta\n", nombre, saldoInicial)
	fmt.Printf("When: %s abona %.2f\n", nombre, cantidad)
	_, err := cuenta.Abonar(cantidad)
	if err != nil && iguales(cuenta.ObtenerSaldo(), saldoInicial) {
		fmt.Printf("Then: Resultado: ÉXITO. El sistema atrapó el error: %v\n", err)
		imprimirSaldo(cuenta)
	} else {
		fmt.Println("Resultado: FALLO. Se aceptó un abono inválido.")
	}
	fmt.Println()
}
