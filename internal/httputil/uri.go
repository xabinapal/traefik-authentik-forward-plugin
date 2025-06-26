package httputil

import (
	"net/http"
	"net/url"
)

func GetRequestURI(req *http.Request) (*url.URL, error) {
	uri, err := url.Parse(req.RequestURI)
	if err != nil {
		return nil, err
	}

	if req.TLS == nil {
		uri.Scheme = "http"
	} else {
		uri.Scheme = "https"
	}

	uri.Host = req.Host

	return uri, nil
}
