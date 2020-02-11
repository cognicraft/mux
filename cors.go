package mux

import (
	"fmt"
	"net/http"
	"strings"
)

func SetAllCORSHeaders(w http.ResponseWriter, r *http.Request, ac AccessControl) {
	origin := r.Header.Get(HeaderOrigin)
	if origin == "" {
		return
	}
	if ac.AllowOrigin != "" {
		if ac.AllowOrigin == "*" {
			if ac.AllowCredentials {
				w.Header().Set(HeaderAccessControlAllowOrigin, origin)
				w.Header().Add(HeaderVary, HeaderOrigin)
			} else {
				w.Header().Set(HeaderAccessControlAllowOrigin, ac.AllowOrigin)
			}
		} else {
			w.Header().Set(HeaderAccessControlAllowOrigin, ac.AllowOrigin)
			w.Header().Add(HeaderVary, HeaderOrigin)
		}
	}
	if len(ac.ExposeHeaders) > 0 {
		expose := strings.Join(ac.ExposeHeaders, ", ")
		w.Header().Set(HeaderAccessControlExposeHeaders, expose)
	}
	if ac.MaxAge > 0 {
		w.Header().Set(HeaderAccessControlMaxAge, fmt.Sprintf("%d", ac.MaxAge))
	}
	if ac.AllowCredentials {
		w.Header().Set(HeaderAccessControlAllowCredentials, fmt.Sprintf("%t", ac.AllowCredentials))
	}
	if len(ac.AllowMethods) > 0 {
		methods := strings.Join(ac.AllowMethods, ", ")
		w.Header().Set(HeaderAccessControlAllowMethods, methods)
	}
	if len(ac.AllowHeaders) > 0 {
		hs := strings.Join(ac.AllowHeaders, ", ")
		w.Header().Set(HeaderAccessControlAllowHeaders, hs)
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
					SetAllCORSHeaders(w, r, ac)
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
				if ac.AllowOrigin == "*" {
					if ac.AllowCredentials {
						w.Header().Set(HeaderAccessControlAllowOrigin, origin)
						w.Header().Add(HeaderVary, HeaderOrigin)
					} else {
						w.Header().Set(HeaderAccessControlAllowOrigin, ac.AllowOrigin)
					}
				} else {
					w.Header().Set(HeaderAccessControlAllowOrigin, ac.AllowOrigin)
					w.Header().Add(HeaderVary, HeaderOrigin)
				}
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
		AllowOrigin:      "*",
		AllowCredentials: true,
		MaxAge:           600,
		AllowMethods:     []string{"GET", "POST", "DELETE"},
		AllowHeaders:     []string{"Accept", "Accept-Language", "Content-Type", "Authorization", "If-None-Match"},
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
