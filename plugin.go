package traefik_authentik_forward_plugin

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/xabinapal/traefik-authentik-forward-plugin/internal/authentik"
	"github.com/xabinapal/traefik-authentik-forward-plugin/internal/config"
	"github.com/xabinapal/traefik-authentik-forward-plugin/internal/httpclient"
	"github.com/xabinapal/traefik-authentik-forward-plugin/internal/httputil"
)

func CreateConfig() *config.Config {
	return &config.Config{
		// authentik settings
		Address:                "",
		UnauthorizedStatusCode: config.DefaultUnauthorizedStatusCode,
		RedirectStatusCode:     config.DefaultRedirectStatusCode,
		UnauthorizedPaths:      config.DefaultUnauthorizedPaths,
		RedirectPaths:          config.DefaultRedirectPaths,

		// http settings
		Timeout: config.DefaultTimeout,
		TLS: config.TLSConfig{
			CA:                 "",
			Cert:               "",
			Key:                "",
			MinVersion:         config.DefaultTLSMinVersion,
			MaxVersion:         config.DefaultTLSMaxVersion,
			InsecureSkipVerify: config.DefaultTLSInsecureSkipVerify,
		},
	}
}

type Plugin struct {
	next   http.Handler
	config *config.PluginConfig
	name   string
	client *authentik.Client
}

func New(ctx context.Context, next http.Handler, config *config.Config, name string) (http.Handler, error) {
	pc, err := config.Parse()
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	httpClient, err := httpclient.New(pc.HTTPClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create http client: %w", err)
	}

	client := authentik.NewClient(httpClient, pc.Authentik)

	return &Plugin{
		next:   next,
		config: pc,
		name:   name,
		client: client,
	}, nil
}

func (p *Plugin) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	reqURL, err := httputil.GetRequestURI(req)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	meta := &authentik.RequestMeta{
		URL:     reqURL,
		Cookies: authentik.GetCookies(req),
	}

	authentik.RequestMangle(req)

	if strings.HasPrefix(reqURL.Path, authentik.BasePath) {
		if req.Method == http.MethodGet {
			p.handleAuthentik(meta, rw)
		} else {
			rw.WriteHeader(http.StatusMethodNotAllowed)
			_, _ = rw.Write([]byte(http.StatusText(http.StatusMethodNotAllowed)))
			return
		}
	} else if state, err := p.client.CheckRequest(meta); err == nil {
		p.handleUpstream(state, req, rw)
	} else {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
}

func (p *Plugin) handleAuthentik(meta *authentik.RequestMeta, rw http.ResponseWriter) {
	if !authentik.IsAuthentikPathAllowed(meta.URL.Path) {
		rw.WriteHeader(http.StatusNotFound)
		_, _ = rw.Write([]byte(http.StatusText(http.StatusNotFound)))
		return
	}

	// send request to authentik
	res, err := p.client.Request(meta, meta.URL.Path, meta.URL.RawQuery)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	defer func() { _ = res.Body.Close() }()

	// copy headers from response to request
	for k, vs := range res.Header {
		if strings.HasPrefix(k, authentik.HeaderPrefix) {
			continue
		}

		for _, v := range vs {
			rw.Header().Add(k, v)
		}
	}

	// send response
	rw.WriteHeader(res.StatusCode)

	// send response body
	if _, err := io.Copy(rw, res.Body); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (p *Plugin) handleUpstream(meta *authentik.ResponseMeta, req *http.Request, rw http.ResponseWriter) {
	sc := p.config.Authentik.GetUnauthorizedStatusCode(meta.URL.Path)

	if !meta.IsAuthenticated && sc != http.StatusOK {
		p.serveUnauthorized(meta, rw, sc)
	} else {
		p.serveUpstream(meta, req, rw)
	}
}

func (p *Plugin) serveUnauthorized(meta *authentik.ResponseMeta, rw http.ResponseWriter, sc int) {
	if sc >= 300 && sc < 400 {
		loc := authentik.GetAuthentikStartPath(meta.URL)
		rw.Header().Set("Location", loc)
	}

	for _, c := range meta.Cookies {
		rw.Header().Add("Set-Cookie", c.String())
	}

	rw.WriteHeader(sc)
	_, _ = rw.Write([]byte(http.StatusText(sc)))
}

func (p *Plugin) serveUpstream(meta *authentik.ResponseMeta, req *http.Request, rw http.ResponseWriter) {
	for k, vs := range meta.Headers {
		for _, v := range vs {
			req.Header.Add(k, v)
		}
	}

	rcm := &httputil.ResponseMangler{
		ResponseWriter: rw,
		MangleFunc:     authentik.GetResponseMangler(meta.Cookies),
	}

	p.next.ServeHTTP(rcm, req)
}
