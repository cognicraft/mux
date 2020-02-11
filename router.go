package mux

import (
	"bytes"
	"net/http"
	"strings"

	"context"
)

func New() *Router {
	return &Router{}
}

type Router struct {
	root *Route
}

func (r *Router) Route(path string) *Route {
	if r.root == nil {
		r.root = &Route{
			path:     "/",
			handlers: map[string]http.Handler{},
		}
	}
	if strings.HasPrefix(path, "/") {
		path = path[1:]
	}
	return r.root.Route(path)
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if r.root == nil {
		http.NotFound(w, req)
		return
	}
	route, vars := r.root.Match(req.URL.Path[1:])
	if route == nil {
		http.NotFound(w, req)
		return
	}
	h, ok := route.Handler(req.Method)
	if !ok {
		// options
		if req.Method == "OPTIONS" {
			ac := AccessControlDefaults
			ac.AllowMethods = route.Methods()
			ac.AllowMethods = append(ac.AllowMethods, "OPTIONS")
			SetAllCORSHeaders(w, req, ac)
			w.WriteHeader(http.StatusNoContent)
			return
		}
		http.NotFound(w, req)
		return
	}
	for k, v := range vars {
		req = req.WithContext(context.WithValue(req.Context(), k, v))
	}
	h.ServeHTTP(w, req)
}

func (r *Router) String() string {
	var buf bytes.Buffer
	buf.WriteString(strings.Repeat("-", 75))
	buf.WriteString("\n")
	buf.WriteString(Tree(r.root))
	buf.WriteString(strings.Repeat("-", 75))
	buf.WriteString("\n")
	return buf.String()
}
