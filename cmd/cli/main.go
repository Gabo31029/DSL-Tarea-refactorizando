/*
source file: cmd/cli/main.go [CLI v2.0.0]
entrega 3: misma CLI, ahora usa el paquete cuenta
nota: ejecutar desde la carpeta del proyecto (usa data/ y logs/)
*/
package main

import (
	"bufio"
	"flag"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"prj_withdrawal/cuenta"
)

const (
	archDatos = "data/cuentas.json"
	archLog   = "logs/transacciones.json"
)

func main() {
	nombre := flag.String("nombre", "", "Nombre del titular de la cuenta")
	saldo := flag.Float64("saldo", 0.0, "Saldo inicial de la cuenta")
	monto := flag.Float64("monto", 0.0, "Monto a retirar o abonar")
	retiro := flag.Float64("retiro", 0.0, "Monto a retirar (igual que en la v1.2.0)")
	op := flag.String("op", "retiro", "Operacion: retiro o abono")
	escenario := flag.String("test", "", "Forzar test: exito, fallo, error")
	menu := flag.Bool("menu", false, "Modo interactivo (usa data/cuentas.json)")
	cuentaID := flag.String("cuenta", "", "ID de la cuenta en data/cuentas.json, ej. C001")
	flag.Parse()

	if len(os.Args) == 1 {
		fmt.Println("=== EJECUTANDO PRUEBAS POR DEFECTO (Gherkin) ===")
		cuenta.TestEscenario1Exito("Alice", 234.56, 200.00)
		cuenta.TestEscenario2Fallo("Hamza", 198.76, 200.00)
		cuenta.TestEscenario3Error("Pedro", 300.00, -20.00)
		cuenta.TestEscenario4Abono("Alice", 100.00, 50.50)
		cuenta.TestEscenario5AbonoInvalido("Hamza", 80.00, 0)
		return
	}

	if *menu {
		modoInteractivo()
		return
	}

	m := *monto
	if *retiro != 0 {
		m = *retiro
	}

	// transaccion real sobre data/cuentas.json, queda en el log
	if *cuentaID != "" {
		banco := cargarBanco()
		mostrar(banco.Procesar(strings.ToUpper(*cuentaID), strings.ToUpper(*op), m))
		return
	}

	if *nombre == "" {
		fmt.Println("Error: falta el -nombre. Usa -h para ver la ayuda.")
		os.Exit(1)
	}
	// montos menores a un centimo (o NaN) se tratan como error
	montoValido := m >= 0.01

	switch *op {
	case "abono":
		if montoValido {
			cuenta.TestEscenario4Abono(*nombre, *saldo, m)
		} else {
			cuenta.TestEscenario5AbonoInvalido(*nombre, *saldo, m)
		}
		return
	case "retiro":
	default:
		fmt.Printf("Error: operacion '%s' no valida (usa retiro o abono)\n", *op)
		os.Exit(1)
	}

	switch *escenario {
	case "exito":
		cuenta.TestEscenario1Exito(*nombre, *saldo, m)
	case "fallo":
		cuenta.TestEscenario2Fallo(*nombre, *saldo, m)
	case "error":
		cuenta.TestEscenario3Error(*nombre, *saldo, m)
	default:
		fmt.Println("=== EJECUTANDO TEST DINÁMICO AUTOMÁTICO ===")
		if !montoValido {
			cuenta.TestEscenario3Error(*nombre, *saldo, m)
		} else if m > *saldo {
			cuenta.TestEscenario2Fallo(*nombre, *saldo, m)
		} else {
			cuenta.TestEscenario1Exito(*nombre, *saldo, m)
		}
	}
}

var entrada = bufio.NewScanner(os.Stdin)

// leer devuelve false cuando ya no hay entrada (Ctrl+D / Ctrl+Z)
func leer(msg string) (string, bool) {
	fmt.Print(msg)
	if !entrada.Scan() {
		fmt.Println()
		return "", false
	}
	return strings.TrimSpace(entrada.Text()), true
}

// acepta 10.50 o 10,50
func leerMonto(msg string) (float64, bool) {
	s, ok := leer(msg)
	if !ok {
		return 0, false
	}
	v, err := strconv.ParseFloat(strings.Replace(s, ",", ".", 1), 64)
	if err != nil {
		return math.NaN(), true // Retirar/Abonar lo rechazan como monto no valido
	}
	return v, true
}

func cargarBanco() *cuenta.Banco {
	banco, err := cuenta.CargarBanco(archDatos, archLog)
	if err != nil {
		fmt.Println("Error al leer los datos:", err)
		os.Exit(1)
	}
	return banco
}

func mostrar(tx cuenta.Transaccion) {
	fmt.Printf("[%s] %s %.2f en %s: %s (saldo: %.2f)\n",
		tx.Estado, tx.Tipo, tx.Monto, tx.CuentaID, tx.Mensaje, tx.SaldoDespues)
}

func modoInteractivo() {
	banco := cargarBanco()

	fmt.Println("Cuentas:")
	for _, c := range banco.Cuentas() {
		estado := "activa"
		if !c.Activa {
			estado = "inactiva"
		}
		fmt.Printf("  %s  %-8s %10.2f  %s\n", c.ID, c.Nombre, c.Saldo, estado)
	}

	var id string
	for {
		s, ok := leer("ID de cuenta: ")
		if !ok {
			return
		}
		id = strings.ToUpper(s)
		if _, existe := banco.Cuenta(id); existe {
			break
		}
		fmt.Println("Esa cuenta no existe.")
	}

	for {
		fmt.Println("\n1) Retirar  2) Abonar  3) Ver saldo  0) Salir")
		opcion, ok := leer("Opcion: ")
		if !ok {
			return
		}
		switch opcion {
		case "1", "2":
			m, ok := leerMonto("Monto: ")
			if !ok {
				return
			}
			tipo := "RETIRO"
			if opcion == "2" {
				tipo = "ABONO"
			}
			mostrar(banco.Procesar(id, tipo, m))
		case "3":
			c, _ := banco.Cuenta(id)
			fmt.Printf("Saldo actual: %.2f\n", c.Saldo)
		case "0":
			fmt.Println("Hasta luego.")
			return
		default:
			fmt.Println("Opcion no valida.")
		}
	}
}
