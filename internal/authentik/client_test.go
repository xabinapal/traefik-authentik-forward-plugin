package authentik_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/xabinapal/traefik-authentik-forward-plugin/internal/authentik"
)

func TestCheck(t *testing.T) {
	t.Run("with unauthenticated response", func(t *testing.T) {
		akCalled := false
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			akCalled = true

			// check that the path is set correctly
			expectedPath := "/outpost.goauthentik.io/auth/nginx"
			if r.URL.Path != expectedPath {
				t.Fatalf("expected path %s, got %s", expectedPath, r.URL.Path)
			}

			// set response cookies
			http.SetCookie(w, &http.Cookie{
				Name:  "authentik_proxy_session",
				Value: "test-session",
			})

			// send response
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer server.Close()

		config := &authentik.Config{Address: server.URL}
		client := authentik.NewClient(context.Background(), server.Client(), config)

		reqMeta := &authentik.RequestMeta{
			URL: &url.URL{
				Scheme: "https",
				Host:   "example.com",
				Path:   "/protected",
			},
			Cookies: []*http.Cookie{},
		}

		resMeta, err := client.Check(reqMeta)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// check that the authentik server was called
		if !akCalled {
			t.Fatalf("expected authentik server to be called")
		}

		// check that the request was unauthenticated
		if resMeta.Session.IsAuthenticated {
			t.Error("expected request to be unauthenticated")
		}

		// check that the request url is set correctly
		if resMeta.URL.String() != reqMeta.URL.String() {
			t.Errorf("expected request url to be %v, got %v", reqMeta.URL.String(), resMeta.URL.String())
		}

		// check that the received headers are set correctly
		if len(resMeta.Session.Headers) != 0 {
			t.Errorf("expected 0 received headers, got %d", len(resMeta.Session.Headers))
		}

		// check that the received cookies are set correctly
		if len(resMeta.Session.Cookies) != 1 {
			t.Fatalf("expected 1 received cookie, got %d", len(resMeta.Session.Cookies))
		}

		expectedCookieName := "authentik_proxy_session"
		expectedCookieValue := "test-session"
		for _, c := range resMeta.Session.Cookies {
			if c.Name != expectedCookieName || c.Value != expectedCookieValue {
				t.Errorf("expected received cookie %s=%s, got %s=%s", expectedCookieName, expectedCookieValue, c.Name, c.Value)
			}
		}
	})

	t.Run("with authenticated response", func(t *testing.T) {
		akCalled := false
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			akCalled = true

			// check that the path is set correctly
			expectedPath := "/outpost.goauthentik.io/auth/nginx"
			if r.URL.Path != expectedPath {
				t.Fatalf("expected path %s, got %s", expectedPath, r.URL.Path)
			}

			// set response headers
			w.Header().Set("X-Authentik-User", "testuser")

			// set response cookies
			http.SetCookie(w, &http.Cookie{
				Name:  "authentik_proxy_session",
				Value: "test-session",
			})

			// send response
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		config := &authentik.Config{Address: server.URL}
		client := authentik.NewClient(context.Background(), server.Client(), config)

		reqMeta := &authentik.RequestMeta{
			URL: &url.URL{
				Scheme: "https",
				Host:   "example.com",
				Path:   "/protected",
			},
			Cookies: []*http.Cookie{},
		}

		resMeta, err := client.Check(reqMeta)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// check that the authentik server was called
		if !akCalled {
			t.Fatalf("expected authentik server to be called")
		}

		// check that the request was authenticated
		if !resMeta.Session.IsAuthenticated {
			t.Error("expected request to be authenticated")
		}

		// check that the request url is set correctly
		if resMeta.URL.String() != reqMeta.URL.String() {
			t.Errorf("expected request url to be %v, got %v", reqMeta.URL.String(), resMeta.URL.String())
		}

		// check that the received headers are set correctly
		if len(resMeta.Session.Headers) != 1 {
			t.Errorf("expected 1 received header, got %d", len(resMeta.Session.Headers))
		}

		expectedHeaderName := "X-Authentik-User"
		expectedHeaderValue := "testuser"
		for k, v := range resMeta.Session.Headers {
			if k != expectedHeaderName || v[0] != expectedHeaderValue {
				t.Errorf("expected received header %s=%s, got %s=%s", expectedHeaderName, expectedHeaderValue, k, v[0])
			}
		}

		// check that the received cookies are set correctly
		if len(resMeta.Session.Cookies) != 1 {
			t.Fatalf("expected 1 received cookie, got %d", len(resMeta.Session.Cookies))
		}

		expectedCookieName := "authentik_proxy_session"
		expectedCookieValue := "test-session"
		for _, c := range resMeta.Session.Cookies {
			if c.Name != expectedCookieName || c.Value != expectedCookieValue {
				t.Errorf("expected received cookie %s=%s, got %s=%s", expectedCookieName, expectedCookieValue, c.Name, c.Value)
			}
		}
	})

	t.Run("with server error", func(t *testing.T) {
		akCalled := false
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			akCalled = true

			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		config := &authentik.Config{Address: server.URL}
		client := authentik.NewClient(context.Background(), server.Client(), config)

		meta := &authentik.RequestMeta{
			URL: &url.URL{
				Scheme: "https",
				Host:   "example.com",
				Path:   "/protected",
			},
			Cookies: []*http.Cookie{},
		}

		resMeta, err := client.Check(meta)

		// check that the authentik server was called
		if !akCalled {
			t.Fatalf("expected authentik server to be called")
		}

		// check that the error is not nil
		if err == nil {
			t.Fatalf("expected error, got none")
		}

		// check that the response meta is nil
		if resMeta != nil {
			t.Errorf("expected resMeta to be nil, got %v", resMeta)
		}
	})
}

func TestRequest(t *testing.T) {
	t.Run("with authenticated request", func(t *testing.T) {
		akCalled := false
		akServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			akCalled = true

			// check that the path is set correctly
			expectedPath := "/outpost.goauthentik.io/callback"
			if r.URL.Path != expectedPath {
				t.Errorf("expected path %s, got %s", expectedPath, r.URL.Path)
			}

			// check that the query is set correctly
			expectedQuery := "additional=param"
			if r.URL.RawQuery != expectedQuery {
				t.Errorf("expected query %s, got %s", expectedQuery, r.URL.RawQuery)
			}

			// check that the forwarded host header is set
			expectedForwardedHost := "example.com"
			if r.Header.Get("X-Forwarded-Host") != expectedForwardedHost {
				t.Errorf("expected X-Forwarded-Host to be %s, got %s", expectedForwardedHost, r.Header.Get("X-Forwarded-Host"))
			}

			// check that the original uri header is set
			expectedURI := "https://example.com/protected?query=value"
			if r.Header.Get("X-Original-Uri") != expectedURI {
				t.Errorf("expected X-Original-Uri to be %s, got %s", expectedURI, r.Header.Get("X-Original-Uri"))
			}

			// check that the sent cookies are set correctly
			cookies := r.Cookies()
			if len(cookies) != 1 {
				t.Fatalf("expected 1 received cookie, got %d", len(cookies))
			}

			expectedCookieName := "authentik_proxy_session"
			expectedCookieValue := "test-session"
			for _, c := range cookies {
				if c.Name != expectedCookieName || c.Value != expectedCookieValue {
					t.Errorf("expected received cookie %s=%s, got %s=%s", expectedCookieName, expectedCookieValue, c.Name, c.Value)
				}
			}

			w.WriteHeader(http.StatusOK)
		}))
		defer akServer.Close()

		config := &authentik.Config{Address: akServer.URL}
		client := authentik.NewClient(context.Background(), akServer.Client(), config)

		meta := &authentik.RequestMeta{
			URL: &url.URL{
				Scheme:   "https",
				Host:     "example.com",
				Path:     "/protected",
				RawQuery: "query=value",
			},
			Cookies: []*http.Cookie{
				{Name: "authentik_proxy_session", Value: "test-session"},
			},
		}

		response, err := client.Request(meta, "/outpost.goauthentik.io/callback", "additional=param")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer func() { _ = response.Body.Close() }()

		// check that the authentik server was called
		if !akCalled {
			t.Fatalf("expected authentik server to be called")
		}

		// check that the status code is correct
		expectedStatusCode := http.StatusOK
		if response.StatusCode != expectedStatusCode {
			t.Errorf("expected status %d, got %d", expectedStatusCode, response.StatusCode)
		}
	})

	t.Run("with mangled location", func(t *testing.T) {
		akServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Location", "https://example.com/mangled")
			w.WriteHeader(http.StatusFound)
		}))
		akServer.Client().CheckRedirect = func(req *http.Request, via []*http.Request) error {
			// don't follow redirects
			return http.ErrUseLastResponse
		}
		defer akServer.Close()

		config := &authentik.Config{Address: akServer.URL}
		client := authentik.NewClient(context.Background(), akServer.Client(), config)

		meta := &authentik.RequestMeta{
			URL: &url.URL{
				Scheme: "https",
				Host:   "example.com",
				Path:   "/protected",
			},
			Cookies: []*http.Cookie{},
		}

		response, err := client.Request(meta, "/outpost.goauthentik.io/callback", "additional=param")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer func() { _ = response.Body.Close() }()

		// check that the status code is correct
		expectedStatusCode := http.StatusFound
		if response.StatusCode != expectedStatusCode {
			t.Errorf("expected status %d, got %d", expectedStatusCode, response.StatusCode)
		}

		// check that the location is set correctly
		expectedLocation := "https://example.com/mangled"
		if response.Header.Get("Location") != expectedLocation {
			t.Errorf("expected location %s, got %s", expectedLocation, response.Header.Get("Location"))
		}
	})

	t.Run("with mangled cookies", func(t *testing.T) {
		akServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Set-Cookie", "authentik_proxy_session=mangled-session")
			w.WriteHeader(http.StatusOK)
		}))
		defer akServer.Close()

		config := &authentik.Config{Address: akServer.URL}
		client := authentik.NewClient(context.Background(), akServer.Client(), config)

		meta := &authentik.RequestMeta{
			URL: &url.URL{
				Scheme: "https",
				Host:   "example.com",
				Path:   "/protected",
			},
			Cookies: []*http.Cookie{},
		}
		response, err := client.Request(meta, "/outpost.goauthentik.io/callback", "additional=param")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer func() { _ = response.Body.Close() }()

		// check that the status code is correct
		if response.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", response.StatusCode)
		}

		// check that the cookie is set correctly
		if len(response.Cookies()) != 1 {
			t.Errorf("expected 1 cookie, got %d", len(response.Cookies()))
		}

		expectedName := "authentik_proxy_session"
		expectedValue := "mangled-session"
		if response.Cookies()[0].Name != expectedName || response.Cookies()[0].Value != expectedValue {
			t.Errorf("expected cookie %s=%s, got %s=%s", expectedName, expectedValue, response.Cookies()[0].Name, response.Cookies()[0].Value)
		}
	})
}
