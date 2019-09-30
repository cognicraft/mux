package mux

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

const (
	HeaderAcceptEncoding  = "Accept-Encoding"
	HeaderContentType     = "Content-Type"
	HeaderContentEncoding = "Content-Encoding"
	HeaderVary            = "Vary"
)

const (
	ContentEncodingGZIP = "gzip"
)

func GZIP(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip compression if the client doesn't accept gzip encoding.
		if !strings.Contains(r.Header.Get(HeaderAcceptEncoding), ContentEncodingGZIP) {
			h.ServeHTTP(w, r)
			return
		}
		// Skip compression if already compressed
		if w.Header().Get(HeaderContentEncoding) == ContentEncodingGZIP {
			h.ServeHTTP(w, r)
			return
		}

		w.Header().Set(HeaderContentEncoding, ContentEncodingGZIP)
		w.Header().Set(HeaderVary, HeaderAcceptEncoding)

		gz := gzip.NewWriter(w)
		gzr := &gzipResponseWriter{Writer: gz, ResponseWriter: w}
		h.ServeHTTP(gzr, r)
		gz.Close()
	})
}

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	if w.Header().Get(HeaderContentType) == "" {
		// If no content type, apply sniffing algorithm to un-gzipped body.
		w.Header().Set(HeaderContentType, http.DetectContentType(b))
	}
	return w.Writer.Write(b)
}
