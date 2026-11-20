// pruebas de la API con los mismos escenarios de los .feature
package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"prj_withdrawal/cuenta"
)

func servidorPrueba(t *testing.T) *http.ServeMux {
	dir := t.TempDir()
	datos := filepath.Join(dir, "cuentas.json")
	os.WriteFile(datos, []byte(`[{"id":"C001","nombre":"Alice","saldo":234.56,"activa":true},
		{"id":"C002","nombre":"Hamza","saldo":198.76,"activa":true}]`), 0o644)
	b, err := cuenta.CargarBanco(datos, filepath.Join(dir, "log.json"))
	if err != nil {
		t.Fatal(err)
	}
	return nuevoMux(b)
}

func TestAPITransaccion(t *testing.T) {
	casos := []struct {
		escenario, body string
		codigo          int
		estado          string
	}{
		{"Successful withdrawal", `{"cuenta_id":"C001","tipo":"RETIRO","monto":200}`, 200, "EXITOSA"},
		{"Declined withdrawal", `{"cuenta_id":"C002","tipo":"RETIRO","monto":200}`, 409, "RECHAZADA"},
		{"Successful deposit", `{"cuenta_id":"C002","tipo":"ABONO","monto":50}`, 200, "EXITOSA"},
		{"Invalid deposit", `{"cuenta_id":"C002","tipo":"ABONO","monto":-5}`, 422, "ERROR"},
	}
	for _, c := range casos {
		t.Run(c.escenario, func(t *testing.T) {
			mux := servidorPrueba(t)
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, httptest.NewRequest("POST", "/api/transaccion", strings.NewReader(c.body)))
			if rec.Code != c.codigo || !strings.Contains(rec.Body.String(), c.estado) {
				t.Errorf("código=%d body=%s; esperado %d %s", rec.Code, rec.Body.String(), c.codigo, c.estado)
			}
		})
	}
}

func TestPaginaPrincipal(t *testing.T) {
	rec := httptest.NewRecorder()
	servidorPrueba(t).ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), "Cajero") {
		t.Errorf("la página principal no cargó (código %d)", rec.Code)
	}
}

func TestJSONMalo(t *testing.T) {
	rec := httptest.NewRecorder()
	servidorPrueba(t).ServeHTTP(rec, httptest.NewRequest("POST", "/api/transaccion", strings.NewReader("{monto:")))
	if rec.Code != http.StatusBadRequest {
		t.Errorf("codigo = %d, esperado 400", rec.Code)
	}
}

// varias peticiones al mismo tiempo: el saldo final tiene que cuadrar
func TestConcurrencia(t *testing.T) {
	mux := servidorPrueba(t)
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			body := `{"cuenta_id":"C001","tipo":"RETIRO","monto":1}`
			mux.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("POST", "/api/transaccion", strings.NewReader(body)))
		}()
	}
	wg.Wait()

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest("GET", "/api/cuentas", nil))
	if !strings.Contains(rec.Body.String(), `"saldo":184.56`) {
		t.Errorf("saldo final incorrecto: %s", rec.Body.String())
	}
}
