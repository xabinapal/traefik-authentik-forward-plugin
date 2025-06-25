package traefik_authentik_forward_plugin

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"text/template"

	"github.com/xabinapal/traefik-authentik-forward-plugin/internal"
)

func CreateConfig() *internal.Config {
	return &internal.Config{
		Headers: make(map[string]string),
	}
}

type Plugin struct {
	next     http.Handler
	headers  map[string]string
	name     string
	template *template.Template
}

func New(ctx context.Context, next http.Handler, config *internal.Config, name string) (http.Handler, error) {
	if len(config.Headers) == 0 {
		return nil, fmt.Errorf("headers cannot be empty")
	}

	return &Plugin{
		headers:  config.Headers,
		next:     next,
		name:     name,
		template: template.New("demo").Delims("[[", "]]"),
	}, nil
}

func (a *Plugin) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	for key, value := range a.headers {
		tmpl, err := a.template.Parse(value)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		writer := &bytes.Buffer{}

		err = tmpl.Execute(writer, req)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		req.Header.Set(key, writer.String())
	}

	a.next.ServeHTTP(rw, req)
}
