/*
source file: cmd/web/main.go [App Web v2.0.0]
entrega 3: retiro, abono y transacciones desde el navegador
ejecutar desde la carpeta del proyecto: go run ./cmd/web
*/
package main

import (
	"embed"
	"encoding/json"
	"flag"
	"io/fs"
	"log"
	"net/http"
	"strings"

	"prj_withdrawal/cuenta"
)

//go:embed static/index.html
var static embed.FS

func main() {
	puerto := flag.String("puerto", "8080", "Puerto HTTP")
	datos := flag.String("datos", "data/cuentas.json", "Archivo de cuentas")
	logTx := flag.String("log", "logs/transacciones.json", "Archivo de log")
	flag.Parse()

	banco, err := cuenta.CargarBanco(*datos, *logTx)
	if err != nil {
		log.Fatal("Error al leer los datos: ", err)
	}

	log.Printf("Cajero web en http://localhost:%s", *puerto)
	log.Fatal(http.ListenAndServe(":"+*puerto, nuevoMux(banco)))
}

// las rutas van aparte para poder probarlas con httptest
func nuevoMux(banco *cuenta.Banco) *http.ServeMux {
	pagina, _ := fs.Sub(static, "static")

	mux := http.NewServeMux()
	mux.Handle("GET /", http.FileServerFS(pagina))

	mux.HandleFunc("GET /api/cuentas", func(w http.ResponseWriter, r *http.Request) {
		responder(w, http.StatusOK, banco.Cuentas())
	})

	mux.HandleFunc("GET /api/transacciones", func(w http.ResponseWriter, r *http.Request) {
		responder(w, http.StatusOK, banco.Historial())
	})

	mux.HandleFunc("POST /api/transaccion", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			CuentaID string  `json:"cuenta_id"`
			Tipo     string  `json:"tipo"`
			Monto    float64 `json:"monto"`
		}
		r.Body = http.MaxBytesReader(w, r.Body, 1<<16)
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			responder(w, http.StatusBadRequest, map[string]string{"error": "datos invalidos"})
			return
		}
		tx := banco.Procesar(strings.ToUpper(req.CuentaID), strings.ToUpper(req.Tipo), req.Monto)

		codigo := http.StatusOK
		switch tx.Estado {
		case cuenta.Rechazada:
			codigo = http.StatusConflict
		case cuenta.Error:
			codigo = http.StatusUnprocessableEntity
		}
		responder(w, codigo, tx)
	})

	return mux
}

func responder(w http.ResponseWriter, codigo int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(codigo)
	json.NewEncoder(w).Encode(v)
}
