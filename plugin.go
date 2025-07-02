package traefik_authentik_forward_plugin

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/xabinapal/traefik-authentik-forward-plugin/internal/authentik"
	"github.com/xabinapal/traefik-authentik-forward-plugin/internal/config"
	"github.com/xabinapal/traefik-authentik-forward-plugin/internal/httpclient"
	"github.com/xabinapal/traefik-authentik-forward-plugin/internal/httputil"
)

func CreateConfig() *config.RawConfig {
	return &config.RawConfig{
		Address: "",
		Timeout: config.DefaultTimeout,
		TLS: &config.RawTLSConfig{
			CA:                 "",
			Cert:               "",
			Key:                "",
			MinVersion:         config.DefaultTLSMinVersion,
			MaxVersion:         config.DefaultTLSMaxVersion,
			InsecureSkipVerify: config.DefaultTLSInsecureSkipVerify,
		},
		UnauthorizedStatusCode: config.DefaultUnauthorizedStatusCode,
		RedirectStatusCode:     config.DefaultRedirectStatusCode,
		UnauthorizedPaths:      config.DefaultUnauthorizedPaths,
		RedirectPaths:          config.DefaultRedirectPaths,
	}
}

type Plugin struct {
	next   http.Handler
	config *config.Config
	name   string
	client *http.Client
}

func New(ctx context.Context, next http.Handler, config *config.RawConfig, name string) (http.Handler, error) {
	pc, err := config.Parse()
	if err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	client, err := httpclient.CreateHTTPClient(pc)
	if err != nil {
		return nil, fmt.Errorf("failed to create http client: %w", err)
	}

	return &Plugin{
		next:   next,
		config: pc,
		name:   name,
		client: client,
	}, nil
}

func (a *Plugin) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	reqURL, err := httputil.GetRequestURI(req)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	if strings.HasPrefix(reqURL.Path, authentik.BasePath) {
		a.handleAuthentik(rw, req, reqURL)
	} else {
		a.handleUpstream(rw, req, reqURL)
	}
}

func (a *Plugin) handleAuthentik(rw http.ResponseWriter, req *http.Request, reqURL *url.URL) {
	if !authentik.IsPathAllowedDownstream(reqURL.Path) {
		rw.WriteHeader(http.StatusNotFound)
		_, _ = rw.Write([]byte(http.StatusText(http.StatusNotFound)))
		return
	}

	akRes, err := a.requestAuthentik(req, reqURL, reqURL.Path)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	defer func() { _ = akRes.Body.Close() }()

	a.serveAuthentik(rw, reqURL, akRes)
}

func (a *Plugin) handleUpstream(rw http.ResponseWriter, req *http.Request, reqURL *url.URL) {
	akRes, err := a.requestAuthentik(req, reqURL, authentik.AuthPath)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	defer func() { _ = akRes.Body.Close() }()

	if sc := a.config.GetUnauthorizedStatusCode(reqURL.Path); sc == http.StatusOK || akRes.StatusCode == http.StatusOK {
		a.serveUpstream(rw, req, akRes)
	} else {
		a.serveUnauthorized(rw, reqURL, sc)
	}
}

func (a *Plugin) requestAuthentik(req *http.Request, reqURL *url.URL, akPath string) (*http.Response, error) {
	akReq, err := http.NewRequest(req.Method, a.config.Address+akPath, nil)
	if err != nil {
		return nil, err
	}

	akReq.URL.RawQuery = reqURL.RawQuery

	akReq.Header.Set("X-Forwarded-Host", reqURL.Host)
	akReq.Header.Set("X-Original-Uri", reqURL.String())

	for _, c := range authentik.GetCookies(req) {
		akReq.AddCookie(c)
	}

	akRes, err := a.client.Do(akReq)
	if err != nil {
		return nil, err
	}

	return akRes, nil
}

func (a *Plugin) serveAuthentik(rw http.ResponseWriter, reqURL *url.URL, akRes *http.Response) {
	for k, vs := range akRes.Header {
		rw.Header().Del(k)
		for _, v := range vs {
			rw.Header().Add(k, v)
		}
	}

	location := akRes.Header.Get("Location")
	if location != "" {
		if strings.HasPrefix(location, a.config.Address+authentik.BasePath) {
			location = strings.TrimPrefix(location, a.config.Address)

			locURL, err := url.Parse(location)
			if err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
				return
			}

			locURL.Scheme = reqURL.Scheme
			locURL.Host = reqURL.Host

			location = locURL.String()
		}

		rw.Header().Set("Location", location)
		rw.WriteHeader(akRes.StatusCode)
		return
	}

	rw.WriteHeader(akRes.StatusCode)

	if _, err := io.Copy(rw, akRes.Body); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (a *Plugin) serveUpstream(rw http.ResponseWriter, req *http.Request, akRes *http.Response) {
	akUserHeaders := []string{}
	for k := range req.Header {
		if strings.HasPrefix(k, "X-Authentik-") {
			akUserHeaders = append(akUserHeaders, k)
		}
	}

	for _, h := range akUserHeaders {
		req.Header.Del(h)
	}

	if akRes == nil {
		a.next.ServeHTTP(rw, req)
		return
	}

	headers := authentik.GetHeaders(akRes)
	for k, vs := range headers {
		for _, v := range vs {
			req.Header.Add(k, v)
		}
	}

	cookies := authentik.GetCookies(akRes)
	for _, c := range cookies {
		req.AddCookie(c)
	}

	rcm := &httputil.ResponseCookieModifier{
		ResponseWriter: rw,

		Cookies:       cookies,
		CookiesPrefix: authentik.CookiePrefix,
	}

	a.next.ServeHTTP(rcm, req)
}

func (a *Plugin) serveUnauthorized(rw http.ResponseWriter, reqURL *url.URL, sc int) {
	if sc >= 300 && sc < 400 {
		loc := url.URL{
			Scheme: reqURL.Scheme,
			Host:   reqURL.Host,
			Path:   authentik.BasePath + "/start",
			RawQuery: url.Values{
				"rd": {reqURL.String()},
			}.Encode(),
		}

		rw.Header().Set("Location", loc.String())
	}

	rw.WriteHeader(sc)
	_, _ = rw.Write([]byte(http.StatusText(sc)))
}
