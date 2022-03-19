package router

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/nillga/jwt-server/errors"
)

type vanillaRouter struct{}

var (
	vanillaDispatcher = http.NewServeMux()
)

func NewVanillaRouter() Router {
	return &vanillaRouter{}
}

func (v *vanillaRouter) GET(uri string, f func(w http.ResponseWriter, r *http.Request)) {
	vanillaDispatcher.HandleFunc(uri, func(w http.ResponseWriter, r *http.Request) {
		if invalidMethod(w, r, "GET") {
			return
		}
		f(enableCORS(w), r)
	})
}

func (v *vanillaRouter) POST(uri string, f func(w http.ResponseWriter, r *http.Request)) {
	vanillaDispatcher.HandleFunc(uri, func(w http.ResponseWriter, r *http.Request) {
		if invalidMethod(w, r, "POST") {
			return
		}
		f(enableCORS(w), r)
	})
}

func (v *vanillaRouter) DELETE(uri string, f func(w http.ResponseWriter, r *http.Request)) {
	vanillaDispatcher.HandleFunc(uri, func(w http.ResponseWriter, r *http.Request) {
		if invalidMethod(w, r, "DELETE") {
			return
		}
		f(enableCORS(w), r)
	})
}

func (v *vanillaRouter) PUT(uri string, f func(w http.ResponseWriter, r *http.Request)) {
	vanillaDispatcher.HandleFunc(uri, func(w http.ResponseWriter, r *http.Request) {
		if invalidMethod(w, r, "PUT") {
			return
		}
		f(enableCORS(w), r)
	})
}

func (v *vanillaRouter) SERVE(port string) {
	log.Println("Vanilla Server running on port " + port)
	log.Fatalln(http.ListenAndServe(":"+port, vanillaDispatcher))
}

func invalidMethod(w http.ResponseWriter, r *http.Request, method string) bool {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != method {
		w.Header().Add("Allow", method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(errors.ProceduralError{Message: "Invalid method " + r.Method})
		return true
	}
	return false
}

func enableCORS(w http.ResponseWriter) http.ResponseWriter {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	return w
}
