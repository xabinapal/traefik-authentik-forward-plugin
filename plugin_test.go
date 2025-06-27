package traefik_authentik_forward_plugin_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	plugin "github.com/xabinapal/traefik-authentik-forward-plugin"
	"github.com/xabinapal/traefik-authentik-forward-plugin/internal/config"
)

func TestServeHTTP_UpstreamPaths(t *testing.T) {
	t.Run("unauthenticated request without redirect", func(t *testing.T) {
		akCalled := true
		akServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			akCalled = true

			// check that the forwarded host header is set
			expectedHost := "example.com"
			actualHost := req.Header.Get("X-Forwarded-Host")
			if actualHost != expectedHost {
				t.Errorf("expected X-Forwarded-Host header to be %s, got %s", expectedHost, actualHost)
			}

			// check that the original uri header is set
			expectedURI := "http://example.com/users"
			actualURI := req.Header.Get("X-Original-URI")
			if actualURI != expectedURI {
				t.Errorf("expected X-Original-URI header to be %s, got %s", expectedURI, actualURI)
			}

			rw.WriteHeader(http.StatusUnauthorized)
		}))
		defer akServer.Close()

		next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			// check that the next handler was not called
			t.Fatalf("expected next handler not to be called")
		})

		config := &config.RawConfig{
			Address:                akServer.URL,
			UnauthorizedStatusCode: http.StatusForbidden,
		}
		handler, _ := plugin.New(context.Background(), next, config, "test")

		req := httptest.NewRequest("GET", "http://example.com/users", nil)

		rw := httptest.NewRecorder()
		handler.ServeHTTP(rw, req)

		// check that the authentik server was called
		if !akCalled {
			t.Fatalf("expected authentik server to be called")
		}

		// check that the response status code is the one configured
		expectedCode := http.StatusForbidden
		if rw.Code != expectedCode {
			t.Errorf("expected status %d, got %d", expectedCode, rw.Code)
		}

		// check that the location header is not present
		if rw.Header().Get("Location") != "" {
			t.Errorf("expected location header to be empty, got %s", rw.Header().Get("Location"))
		}
	})

	t.Run("unauthenticated request with redirect", func(t *testing.T) {
		akCalled := true
		akServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			akCalled = true

			// check that the forwarded host header is set
			expectedHost := "example.com"
			actualHost := req.Header.Get("X-Forwarded-Host")
			if actualHost != expectedHost {
				t.Errorf("expected X-Forwarded-Host header to be %s, got %s", expectedHost, actualHost)
			}

			// check that the original uri header is set
			expectedURI := "http://example.com/users"
			actualURI := req.Header.Get("X-Original-URI")
			if actualURI != expectedURI {
				t.Errorf("expected X-Original-URI header to be %s, got %s", expectedURI, actualURI)
			}

			rw.WriteHeader(http.StatusUnauthorized)
		}))
		defer akServer.Close()

		next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			// check that the next handler was not called
			t.Fatalf("expected next handler not to be called")
		})

		config := &config.RawConfig{
			Address:                akServer.URL,
			UnauthorizedStatusCode: http.StatusMovedPermanently,
		}
		handler, _ := plugin.New(context.Background(), next, config, "test")

		req := httptest.NewRequest("GET", "http://example.com/users", nil)

		rw := httptest.NewRecorder()
		handler.ServeHTTP(rw, req)

		// check that the authentik server was called
		if !akCalled {
			t.Fatalf("expected authentik server to be called")
		}

		// check that the response status code is the one configured
		expectedCode := http.StatusMovedPermanently
		actualCode := rw.Code
		if actualCode != expectedCode {
			t.Errorf("expected status %d, got %d", expectedCode, actualCode)
		}

		// check that the location header starts authorization flow
		expectedLocation := "http://example.com/outpost.goauthentik.io/start?rd=http%3A%2F%2Fexample.com%2Fusers"
		actualLocation := rw.Header().Get("Location")
		if actualLocation != expectedLocation {
			t.Errorf("expected location header to be %s, got %s", expectedLocation, actualLocation)
		}
	})

	t.Run("authenticated request", func(t *testing.T) {
		akCalled := true
		akServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			akCalled = true

			// check that the forwarded host header is set
			expectedHost := "example.com"
			actualHost := req.Header.Get("X-Forwarded-Host")
			if actualHost != expectedHost {
				t.Errorf("expected X-Forwarded-Host header to be %s, got %s", expectedHost, actualHost)
			}

			// check that the original uri header is set
			expectedURI := "http://example.com/users"
			actualURI := req.Header.Get("X-Original-URI")
			if actualURI != expectedURI {
				t.Errorf("expected X-Original-URI header to be %s, got %s", expectedURI, actualURI)
			}

			// check that the authentication cookie is set
			expectedCookie := "authentik_proxy_user=testuser"
			actualCookie := req.Header.Get("Cookie")
			if actualCookie != expectedCookie {
				t.Errorf("expected Cookie header to be %s, got %s", expectedCookie, actualCookie)
			}

			rw.Header().Set("X-Authentik-User", "testuser")
			rw.Header().Set("Set-Cookie", "authentik_proxy_user=testuser; Path=/")

			rw.WriteHeader(http.StatusOK)
		}))
		defer akServer.Close()

		nextCalled := false
		next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			nextCalled = true

			// check that authentik headers were added to the request
			expectedUser := "testuser"
			actualUser := req.Header.Get("X-Authentik-User")
			if actualUser != expectedUser {
				t.Errorf("expected X-Authentik-User header to be %s, got %s", expectedUser, actualUser)
			}

			// check that authentik cookies were added to the request
			if _, err := req.Cookie("authentik_proxy_user"); err != nil {
				t.Error("expected authentik_proxy_user cookie to be added to request")
			}

			rw.WriteHeader(http.StatusAccepted)
		})

		config := &config.RawConfig{Address: akServer.URL}
		handler, _ := plugin.New(context.Background(), next, config, "test")

		req := httptest.NewRequest("GET", "http://example.com/users", nil)
		req.AddCookie(&http.Cookie{Name: "authentik_proxy_user", Value: "testuser"})

		rw := httptest.NewRecorder()
		handler.ServeHTTP(rw, req)

		// check that the authentik server was called
		if !akCalled {
			t.Fatalf("expected authentik server to be called")
		}

		// check that the next handler was called
		if !nextCalled {
			t.Fatalf("expected next handler to be called")
		}

		// check that the response status code comes from the upstream
		expectedCode := http.StatusAccepted
		actualCode := rw.Code
		if actualCode != expectedCode {
			t.Errorf("expected status %d, got %d", expectedCode, actualCode)
		}

		// check that the authentik headers were not added to the response
		if rw.Header().Get("X-Authentik-User") != "" {
			t.Errorf("expected X-Authentik-User header to not be added to response")
		}

		// check that authentik cookies were added to the response
		cookies := rw.Result().Cookies()
		if len(cookies) != 1 {
			t.Errorf("expected 1 cookie, got %d", len(cookies))
		}

		if cookies[0].Name != "authentik_proxy_user" {
			t.Errorf("expected authentik_proxy_user cookie, got %s", cookies[0].Name)
		}

		if cookies[0].Value != "testuser" {
			t.Errorf("expected authentik_proxy_user cookie value to be testuser, got %s", cookies[0].Value)
		}
	})
}

func TestServeHTTP_UpstreamPaths_WithPathStatusCodes(t *testing.T) {
	t.Run("unauthenticated request without path status code", func(t *testing.T) {
		akServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			rw.WriteHeader(http.StatusUnauthorized)
		}))
		defer akServer.Close()

		config := &config.RawConfig{
			Address:                akServer.URL,
			UnauthorizedStatusCode: http.StatusUnauthorized,
			UnauthorizedPathStatusCodes: map[string]uint{
				"/users/wrong": http.StatusNotFound,
			},
		}
		handler, _ := plugin.New(context.Background(), nil, config, "test")

		req := httptest.NewRequest("GET", "http://example.com/users", nil)

		rw := httptest.NewRecorder()
		handler.ServeHTTP(rw, req)

		// check that the response status code is the one configured
		expectedCode := http.StatusUnauthorized
		actualCode := rw.Code
		if actualCode != expectedCode {
			t.Errorf("expected status %d, got %d", expectedCode, actualCode)
		}
	})

	t.Run("unauthenticated request with path status code", func(t *testing.T) {
		akServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			rw.WriteHeader(http.StatusUnauthorized)
		}))
		defer akServer.Close()

		config := &config.RawConfig{
			Address:                akServer.URL,
			UnauthorizedStatusCode: http.StatusUnauthorized,
			UnauthorizedPathStatusCodes: map[string]uint{
				"/users": http.StatusNotFound,
			},
		}
		handler, _ := plugin.New(context.Background(), nil, config, "test")

		req := httptest.NewRequest("GET", "http://example.com/users", nil)

		rw := httptest.NewRecorder()
		handler.ServeHTTP(rw, req)

		// check that the response status code is the one configured
		expectedCode := http.StatusNotFound
		actualCode := rw.Code
		if actualCode != expectedCode {
			t.Errorf("expected status %d, got %d", expectedCode, actualCode)
		}
	})

	t.Run("unauthenticated request with multiple path status codes", func(t *testing.T) {
		akServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			rw.WriteHeader(http.StatusUnauthorized)
		}))
		defer akServer.Close()

		config := &config.RawConfig{
			Address:                akServer.URL,
			UnauthorizedStatusCode: http.StatusUnauthorized,
			UnauthorizedPathStatusCodes: map[string]uint{
				"/users":   http.StatusBadRequest,
				"/users/?": http.StatusNotFound,
			},
		}
		handler, _ := plugin.New(context.Background(), nil, config, "test")

		req := httptest.NewRequest("GET", "http://example.com/users", nil)

		rw := httptest.NewRecorder()
		handler.ServeHTTP(rw, req)

		// check that the response status code is the one configured
		expectedCode := http.StatusNotFound
		actualCode := rw.Code
		if actualCode != expectedCode {
			t.Errorf("expected status %d, got %d", expectedCode, actualCode)
		}
	})
}

func TestServeHTTP_AuthentikPaths(t *testing.T) {
	t.Run("explicitly allowed authentik path", func(t *testing.T) {
		akCalled := true
		akServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			akCalled = true

			rw.WriteHeader(http.StatusTeapot)
			rw.Write([]byte("i'm a teapot"))
		}))
		defer akServer.Close()

		config := &config.RawConfig{Address: akServer.URL}
		handler, _ := plugin.New(context.Background(), nil, config, "test")

		req := httptest.NewRequest("GET", "http://example.com/outpost.goauthentik.io/auth/start", nil)

		rw := httptest.NewRecorder()

		handler.ServeHTTP(rw, req)

		// check that the authentik server was called
		if !akCalled {
			t.Fatalf("expected authentik server to be called")
		}

		// check that the response status code comes from the authentik server
		expectedCode := http.StatusTeapot
		actualCode := rw.Code
		if actualCode != expectedCode {
			t.Errorf("expected status %d, got %d", expectedCode, actualCode)
		}

		// check that the response body comes from the authentik server
		expectedBody := "i'm a teapot"
		actualBody := rw.Body.String()
		if actualBody != expectedBody {
			t.Errorf("expected content to be %s, got %s", expectedBody, actualBody)
		}
	})

	t.Run("explicitly denied authentik path", func(t *testing.T) {
		akServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			// check that the authentik server was not called
			t.Fatalf("expected authentik server not to be called")
		}))
		defer akServer.Close()

		config := &config.RawConfig{Address: akServer.URL}
		handler, _ := plugin.New(context.Background(), nil, config, "test")

		req := httptest.NewRequest("GET", "http://example.com/outpost.goauthentik.io/auth/nginx", nil)

		rw := httptest.NewRecorder()
		handler.ServeHTTP(rw, req)

		// check that the response status code is 404
		expectedCode := http.StatusNotFound
		actualCode := rw.Code
		if actualCode != expectedCode {
			t.Errorf("expected status %d, got %d", expectedCode, actualCode)
		}
	})

	t.Run("allowed by default authentik path", func(t *testing.T) {
		akCalled := true
		akServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			akCalled = true

			rw.WriteHeader(http.StatusOK)
			rw.Write([]byte("i'm a teapot"))
		}))
		defer akServer.Close()

		config := &config.RawConfig{Address: akServer.URL}
		handler, _ := plugin.New(context.Background(), nil, config, "test")

		req := httptest.NewRequest("GET", "http://example.com/outpost.goauthentik.io/static/styles.css", nil)

		rw := httptest.NewRecorder()
		handler.ServeHTTP(rw, req)

		// check that the authentik server was called
		if !akCalled {
			t.Fatalf("expected authentik server to be called")
		}

		// check that the response status code comes from the authentik server
		expectedCode := http.StatusOK
		actualCode := rw.Code
		if actualCode != expectedCode {
			t.Errorf("expected status %d, got %d", expectedCode, actualCode)
		}

		// check that the response body comes from the authentik server
		expectedBody := "i'm a teapot"
		actualBody := rw.Body.String()
		if actualBody != expectedBody {
			t.Errorf("expected content to be %s, got %s", expectedBody, actualBody)
		}
	})
}
