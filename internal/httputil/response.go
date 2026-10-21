package httputil

import (
	"net/http"
	"strings"
)

type ResponseMangler struct {
	http.ResponseWriter
	MangleFunc func(rw http.ResponseWriter)
}

func (rcm *ResponseMangler) WriteHeader(code int) {
	rcm.MangleFunc(rcm.ResponseWriter)
	rcm.ResponseWriter.WriteHeader(code)
}

func (rcm *ResponseMangler) Write(data []byte) (int, error) {
	n, err := rcm.ResponseWriter.Write(data)
	// Yaegi can hide http.Flusher when this wrapper crosses a ResponseWriter interface
	// boundary (traefik/yaegi#1600). Flush SSE writes through the original writer too.
	mediaType, _, _ := strings.Cut(rcm.Header().Get("Content-Type"), ";")
	if mediaType == "text/event-stream" {
		rcm.Flush()
	}
	return n, err
}

// Flush preserves streaming responses through the response wrapper.
func (rcm *ResponseMangler) Flush() {
	if flusher, ok := rcm.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}
