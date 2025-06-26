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
	t.Run("unauthenticated request", func(t *testing.T) {
		akCalled := true
		akServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			akCalled = true

			// check that the forwarded host header is set
			if req.Header.Get("X-Forwarded-Host") != "example.com" {
				t.Errorf("expected X-Forwarded-Host header")
			}

			// check that the original uri header is set
			if req.Header.Get("X-Original-URI") != "http://example.com/api/users" {
				t.Errorf("expected X-Original-URI header")
			}

			rw.Header().Set("Location", "/outpost.goauthentik.go/login")
			rw.WriteHeader(http.StatusFound)
		}))
		defer akServer.Close()

		next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			// check that the next handler was not called
			t.Fatalf("expected next handler not to be called")
		})

		config := &config.Config{Address: akServer.URL, KeepPrefix: "/test"}
		handler, _ := plugin.New(context.Background(), next, config, "test")

		req := httptest.NewRequest("GET", "http://example.com/api/users", nil)

		rw := httptest.NewRecorder()
		handler.ServeHTTP(rw, req)

		// check that the authentik server was called
		if !akCalled {
			t.Fatalf("expected authentik server to be called")
		}

		// check that the response status code comes from the authentik server
		if rw.Code != http.StatusFound {
			t.Errorf("expected status 302, got %d", rw.Code)
		}

		// check that the location header comes from the authentik server
		if rw.Header().Get("Location") != "/test/outpost.goauthentik.go/login" {
			t.Errorf("expected location header to be /test/outpost.goauthentik.go/login, got %s", rw.Header().Get("Location"))
		}
	})

	t.Run("authenticated request", func(t *testing.T) {
		akCalled := true
		akServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			akCalled = true

			// check that the forwarded host header is set
			if req.Header.Get("X-Forwarded-Host") != "example.com" {
				t.Errorf("expected X-Forwarded-Host header")
			}

			// check that the original uri header is set
			if req.Header.Get("X-Original-URI") != "http://example.com/api/users" {
				t.Errorf("expected X-Original-URI header")
			}

			// check that the authentication cookie is set
			if req.Header.Get("Cookie") != "authentik_proxy_user=testuser" {
				t.Errorf("expected Cookie header to be authentik_proxy_user=testuser")
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
			if req.Header.Get("X-Authentik-User") != "testuser" {
				t.Error("expected X-Authentik-User header to be added to request")
			}

			// check that authentik cookies were added to the request
			if _, err := req.Cookie("authentik_proxy_user"); err != nil {
				t.Error("expected authentik_proxy_user cookie to be added to request")
			}

			rw.WriteHeader(http.StatusTeapot)
		})

		config := &config.Config{Address: akServer.URL, KeepPrefix: "/test"}
		handler, _ := plugin.New(context.Background(), next, config, "test")

		req := httptest.NewRequest("GET", "http://example.com/api/users", nil)
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
		if rw.Code != http.StatusTeapot {
			t.Errorf("expected status 418, got %d", rw.Code)
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

func TestServeHTTP_AuthentikPaths(t *testing.T) {
	t.Run("explicitly allowed authentik path", func(t *testing.T) {
		akCalled := true
		akServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			akCalled = true

			rw.WriteHeader(http.StatusTeapot)
			rw.Write([]byte("i'm a teapot"))
		}))
		defer akServer.Close()

		config := &config.Config{Address: akServer.URL, KeepPrefix: "/test"}
		handler, _ := plugin.New(context.Background(), nil, config, "test")

		req := httptest.NewRequest("GET", "http://example.com/test/outpost.goauthentik.go/auth/start", nil)

		rw := httptest.NewRecorder()

		handler.ServeHTTP(rw, req)

		// check that the authentik server was called
		if !akCalled {
			t.Fatalf("expected authentik server to be called")
		}

		// check that the response status code comes from the authentik server
		if rw.Code != http.StatusTeapot {
			t.Errorf("expected status 418, got %d", rw.Code)
		}

		// check that the response body comes from the authentik server
		if rw.Body.String() != "i'm a teapot" {
			t.Errorf("expected content to be i'm a teapot, got %s", rw.Body.String())
		}
	})

	t.Run("explicitly denied authentik path", func(t *testing.T) {
		akServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			// check that the authentik server was not called
			t.Fatalf("expected authentik server not to be called")
		}))
		defer akServer.Close()

		config := &config.Config{Address: akServer.URL, KeepPrefix: "/test"}
		handler, _ := plugin.New(context.Background(), nil, config, "test")

		req := httptest.NewRequest("GET", "http://example.com/test/outpost.goauthentik.go/auth/nginx", nil)

		rw := httptest.NewRecorder()
		handler.ServeHTTP(rw, req)

		// check that the response status code is 404
		if rw.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", rw.Code)
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

		config := &config.Config{Address: akServer.URL, KeepPrefix: "/test"}
		handler, _ := plugin.New(context.Background(), nil, config, "test")

		req := httptest.NewRequest("GET", "http://example.com/test/outpost.goauthentik.go/static/styles.css", nil)

		rw := httptest.NewRecorder()
		handler.ServeHTTP(rw, req)

		// check that the authentik server was called
		if !akCalled {
			t.Fatalf("expected authentik server to be called")
		}

		// check that the response status code comes from the authentik server
		if rw.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rw.Code)
		}

		// check that the response body comes from the authentik server
		if rw.Body.String() != "i'm a teapot" {
			t.Errorf("expected content to be i'm a teapot, got %s", rw.Body.String())
		}
	})
}
