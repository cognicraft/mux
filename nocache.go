package mux

import (
	"net/http"
	"strings"
)

// IE-Cache-Buster
func NoCache(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ua := r.UserAgent(); strings.Contains(ua, "MSIE") || strings.Contains(ua, "Trident") {
			// The above alone did not solve the problem in IE9, so here are some more!
			// Set to expire far in the past.
			w.Header().Set("Expires", "Mon, 23 Aug 1982 12:00:00 GMT")

			// If webfont is requested, some cache headers are not allowed
			if strings.HasSuffix(r.URL.Path, ".eot") {
				// Set standard HTTP/1.1 no-cache headers.
				w.Header().Set("Cache-Control", "must-revalidate")
			} else {
				// Set standard HTTP/1.1 no-cache headers.
				w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
				// Set standard HTTP/1.0 no-cache header.
				w.Header().Set("Pragma", "no-cache")
			}

			//w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
			w.Header().Set("Cache-Control", "must-revalidate")
			// Set IE extended HTTP/1.1 no-cache headers (use add).
			w.Header().Add("Cache-Control", "post-check=0, pre-check=0")

		} else {
			w.Header().Add("Cache-Control", "private, max-age=0")
		}
		h.ServeHTTP(w, r)
	})
}
