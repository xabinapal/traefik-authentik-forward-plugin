package httputil

import (
	"net/http"
)

type ResponseMangler struct {
	http.ResponseWriter
	MangleFunc func(rw http.ResponseWriter)
}

func (rcm *ResponseMangler) WriteHeader(code int) {
	rcm.MangleFunc(rcm.ResponseWriter)
	rcm.ResponseWriter.WriteHeader(code)
}
