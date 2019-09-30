package mux

import (
	"log"
	"net/http"
)

func LogRequests(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
		log.Printf("%s requested %s %s", r.RemoteAddr, r.Method, r.URL)
	})
}
