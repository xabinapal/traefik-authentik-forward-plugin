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
	"github.com/xabinapal/traefik-authentik-forward-plugin/internal/httputil"
)

func CreateConfig() *config.RawConfig {
	return &config.RawConfig{
		Address:                "",
		KeepPrefix:             "",
		UnauthorizedStatusCode: http.StatusUnauthorized,
	}
}

type Plugin struct {
	next   http.Handler
	config *config.Config
	name   string
}

func New(ctx context.Context, next http.Handler, config *config.RawConfig, name string) (http.Handler, error) {
	pc, err := config.Parse()
	if err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &Plugin{
		next:   next,
		config: pc,
		name:   name,
	}, nil
}

func (a *Plugin) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if strings.HasPrefix(req.URL.Path, a.config.KeepPrefix+authentik.BasePath) {
		akPath := strings.TrimPrefix(req.URL.Path, a.config.KeepPrefix)

		if !authentik.IsPathAllowed(akPath) {
			rw.WriteHeader(http.StatusNotFound)
			return
		}

		akRes, err := a.requestAuthentik(req, akPath)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		defer akRes.Body.Close()

		a.serveAuthentik(rw, akRes)
	} else if a.config.KeepPrefix == "" || strings.HasPrefix(req.URL.Path, a.config.KeepPrefix) {
		akRes, err := a.requestAuthentik(req, authentik.AuthPath)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		defer akRes.Body.Close()

		if akRes.StatusCode == 200 {
			a.serveUpstream(rw, req, akRes)
		} else {
			a.serveUnauthorized(rw, req)
		}
	} else {
		a.serveUpstream(rw, req, nil)
	}
}

func (a *Plugin) requestAuthentik(req *http.Request, reqPath string) (*http.Response, error) {
	akReq, err := http.NewRequest(req.Method, a.config.Address+reqPath, nil)
	if err != nil {
		return nil, err
	}

	akReq.Header.Set("X-Forwarded-Host", req.Host)
	akReq.Header.Set("X-Original-URI", req.URL.String())

	for _, c := range authentik.GetCookies(req) {
		akReq.AddCookie(c)
	}

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // Don't follow redirects
		},
	}
	akRes, err := client.Do(akReq)
	if err != nil {
		return nil, err
	}

	return akRes, nil
}

func (a *Plugin) serveAuthentik(rw http.ResponseWriter, akRes *http.Response) {
	for k, vs := range akRes.Header {
		rw.Header().Del(k)
		for _, v := range vs {
			rw.Header().Add(k, v)
		}
	}

	location := akRes.Header.Get("Location")
	if location != "" {
		locUrl, err := url.Parse(location)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		if locUrl.IsAbs() && strings.HasPrefix(location, a.config.Address+authentik.BasePath) {
			location = strings.TrimPrefix(location, a.config.Address)
			location = a.config.KeepPrefix + location
		} else if !locUrl.IsAbs() && strings.HasPrefix(location, authentik.BasePath) {
			location = a.config.KeepPrefix + location
		}

		rw.Header().Set("Location", location)
	}

	rw.WriteHeader(akRes.StatusCode)

	if _, err := io.Copy(rw, akRes.Body); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (a *Plugin) serveUpstream(rw http.ResponseWriter, req *http.Request, akRes *http.Response) {
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

func (a *Plugin) serveUnauthorized(rw http.ResponseWriter, req *http.Request) {
	statusCode := a.config.GetUnauthorizedStatusCode(req.URL.Path)

	if statusCode >= 300 && statusCode < 400 {
		loc := url.URL{
			Scheme: req.URL.Scheme,
			Host:   req.Host,
			Path:   a.config.KeepPrefix + authentik.BasePath + "/start",
			RawQuery: url.Values{
				"rd": {req.URL.String()},
			}.Encode(),
		}

		rw.Header().Set("Location", loc.String())
	}

	rw.WriteHeader(statusCode)
}
