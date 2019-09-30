package mux

import (
	"fmt"
	"net/http"
	"strings"
)

func SetAllCORSHeaders(w http.ResponseWriter, ac AccessControl) {
	headers := ac.Headers()
	for k, v := range headers {
		w.Header().Set(k, v)
	}
}

func CORS(ac AccessControl) Decorator {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Does the request have an Origin header?
			origin := r.Header.Get(HeaderOrigin)
			if origin == "" {
				// invalid CORS Request
				next.ServeHTTP(w, r)
				return
			}
			// Is the HTTP method an OPTIONS request?
			if r.Method == http.MethodOptions {
				// Is there an Access-Control-Request-Method header?
				acrm := r.Header.Get(HeaderAccessControlRequestMethod)
				if acrm != "" {
					// Preflight Request
					// TODO: check validity of acrm -> invalid error
					// acrh := r.Header.Get(HeaderAccessControlRequestHeaders)
					// TODO: check validity of acrh -> invalid error
					SetAllCORSHeaders(w, ac)
					w.WriteHeader(http.StatusNoContent)
					return
				}
			}
			// Actual Request
			if len(ac.ExposeHeaders) > 0 {
				expose := strings.Join(ac.ExposeHeaders, ", ")
				w.Header().Set(HeaderAccessControlExposeHeaders, expose)
			}
			if ac.AllowOrigin != "" {
				w.Header().Set(HeaderAccessControlAllowOrigin, ac.AllowOrigin)
			}
			if ac.AllowCredentials {
				w.Header().Set(HeaderAccessControlAllowCredentials, fmt.Sprintf("%t", ac.AllowCredentials))
			}
			next.ServeHTTP(w, r)
		})
	}
}

var (
	AccessControlDefaults  AccessControl
	AccessControlPreflight AccessControl
)

func init() {
	AccessControlDefaults = AccessControl{
		AllowOrigin:  "*",
		MaxAge:       600,
		AllowMethods: []string{"GET", "POST", "DELETE"},
		AllowHeaders: []string{"Accept", "Accept-Language", "Content-Type", "Authorization", "If-None-Match"},
		// Cache-Control, Content-Language, Content-Type, Expires, Last-Modified and Pragma are allowed by cors defaults
		ExposeHeaders: []string{"E-Tag", "Location", "Link"},
	}
}

const (
	HeaderOrigin                        = "Origin"
	HeaderAccessControlAllowOrigin      = "Access-Control-Allow-Origin"
	HeaderAccessControlExposeHeaders    = "Access-Control-Expose-Headers"
	HeaderAccessControlMaxAge           = "Access-Control-Max-Age"
	HeaderAccessControlAllowCredentials = "Access-Control-Allow-Credentials"
	HeaderAccessControlAllowMethods     = "Access-Control-Allow-Methods"
	HeaderAccessControlAllowHeaders     = "Access-Control-Allow-Headers"
)

const (
	HeaderAccessControlRequestMethod  = "Access-Control-Request-Method"
	HeaderAccessControlRequestHeaders = "Access-Control-Request-Headers"
)

type AccessControl struct {
	AllowOrigin      string
	ExposeHeaders    []string
	MaxAge           uint64
	AllowCredentials bool
	AllowMethods     []string
	AllowHeaders     []string
}

func (ac AccessControl) Headers() map[string]string {
	headers := map[string]string{}
	if ac.AllowOrigin != "" {
		headers[HeaderAccessControlAllowOrigin] = ac.AllowOrigin
	}
	if len(ac.ExposeHeaders) > 0 {
		expose := strings.Join(ac.ExposeHeaders, ", ")
		headers[HeaderAccessControlExposeHeaders] = expose
	}
	if ac.MaxAge > 0 {
		headers[HeaderAccessControlMaxAge] = fmt.Sprintf("%d", ac.MaxAge)
	}
	if ac.AllowCredentials {
		headers[HeaderAccessControlAllowCredentials] = fmt.Sprintf("%t", ac.AllowCredentials)
	}
	if len(ac.AllowMethods) > 0 {
		methods := strings.Join(ac.AllowMethods, ", ")
		headers[HeaderAccessControlAllowMethods] = methods
	}
	if len(ac.AllowHeaders) > 0 {
		hs := strings.Join(ac.AllowHeaders, ", ")
		headers[HeaderAccessControlAllowHeaders] = hs
	}
	return headers
}
