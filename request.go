package mux

import (
	"fmt"
	"net/http"
	"net/http/httputil"
)

func DumpRequest(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		bs, _ := httputil.DumpRequest(r, true)
		fmt.Printf("%s\n", string(bs))
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
