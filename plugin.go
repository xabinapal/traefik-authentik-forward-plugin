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
		UnauthorizedStatusCode: http.StatusUnauthorized,
		RedirectStatusCode:     http.StatusFound,
		UnauthorizedPaths:      []string{"^/.*$"},
		RedirectPaths:          []string{},
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
	reqUrl, err := httputil.GetRequestURI(req)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	if strings.HasPrefix(reqUrl.Path, authentik.BasePath) {
		if !authentik.IsPathAllowedDownstream(reqUrl.Path) {
			rw.WriteHeader(http.StatusNotFound)
			rw.Write([]byte(http.StatusText(http.StatusNotFound)))
			return
		}

		akRes, err := a.requestAuthentik(req, reqUrl, reqUrl.Path)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		defer akRes.Body.Close()

		a.serveAuthentik(rw, reqUrl, akRes)
	} else {
		akRes, err := a.requestAuthentik(req, reqUrl, authentik.AuthPath)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		defer akRes.Body.Close()

		if sc := a.config.GetUnauthorizedStatusCode(reqUrl.Path); sc == http.StatusOK || akRes.StatusCode == 200 {
			a.serveUpstream(rw, req, akRes)
		} else {
			a.serveUnauthorized(rw, reqUrl, sc)
		}
	}
}

func (a *Plugin) requestAuthentik(req *http.Request, reqUrl *url.URL, akPath string) (*http.Response, error) {
	akReq, err := http.NewRequest(req.Method, a.config.Address+akPath, nil)
	if err != nil {
		return nil, err
	}

	akReq.URL.RawQuery = reqUrl.RawQuery

	akReq.Header.Set("X-Forwarded-Host", reqUrl.Host)
	akReq.Header.Set("X-Original-URI", reqUrl.String())

	for _, c := range authentik.GetCookies(req) {
		akReq.AddCookie(c)
	}

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // don't follow redirects
		},
	}
	akRes, err := client.Do(akReq)
	if err != nil {
		return nil, err
	}

	return akRes, nil
}

func (a *Plugin) serveAuthentik(rw http.ResponseWriter, reqUrl *url.URL, akRes *http.Response) {
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

			locUrl, err := url.Parse(location)
			if err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
				return
			}

			locUrl.Scheme = reqUrl.Scheme
			locUrl.Host = reqUrl.Host

			location = locUrl.String()
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

func (a *Plugin) serveUnauthorized(rw http.ResponseWriter, reqUrl *url.URL, sc int) {
	if sc >= 300 && sc < 400 {
		loc := url.URL{
			Scheme: reqUrl.Scheme,
			Host:   reqUrl.Host,
			Path:   authentik.BasePath + "/start",
			RawQuery: url.Values{
				"rd": {reqUrl.String()},
			}.Encode(),
		}

		rw.Header().Set("Location", loc.String())
	}

	rw.WriteHeader(sc)
	rw.Write([]byte(http.StatusText(sc)))
}
