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
	url, err := httputil.GetRequestURI(req)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	meta := &authentik.RequestMeta{
		URL:     url,
		Cookies: authentik.GetCookies(req),
	}

	authentik.RequestMangle(req)

	if strings.HasPrefix(url.Path, authentik.BasePath) {
		p.handleAuthentik(meta, req, rw)
	} else {
		p.handleUpstream(meta, req, rw)
	}
}

func (p *Plugin) handleAuthentik(meta *authentik.RequestMeta, req *http.Request, rw http.ResponseWriter) {
	if req.Method != http.MethodGet {
		rw.WriteHeader(http.StatusMethodNotAllowed)
		_, _ = rw.Write([]byte(http.StatusText(http.StatusMethodNotAllowed)))
		return
	}

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

func (p *Plugin) handleUpstream(meta *authentik.RequestMeta, req *http.Request, rw http.ResponseWriter) {
	if p.config.Authentik.IsSkippedPath(meta.URL.Path) {
		p.serveUpstream(nil, req, rw)
		return
	}

	sc := p.config.Authentik.GetUnauthorizedStatusCode(meta.URL.Path)

	resMeta, err := p.client.CheckRequest(meta)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	if !resMeta.IsAuthenticated && sc != http.StatusOK {
		p.serveUnauthorized(resMeta, rw, sc)
	} else {
		p.serveUpstream(resMeta, req, rw)
	}
}

func (p *Plugin) serveUpstream(meta *authentik.ResponseMeta, req *http.Request, rw http.ResponseWriter) {
	var cookies []*http.Cookie

	if meta != nil {
		for k, vs := range meta.Headers {
			for _, v := range vs {
				req.Header.Add(k, v)
			}
		}

		cookies = meta.Cookies
	} else {
		cookies = []*http.Cookie{}
	}

	rcm := &httputil.ResponseMangler{
		ResponseWriter: rw,
		MangleFunc:     authentik.GetResponseMangler(cookies),
	}

	p.next.ServeHTTP(rcm, req)
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
