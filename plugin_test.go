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
	t.Run("request to unobserved prefix", func(t *testing.T) {
		akServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			// check that the authentik server was not called
			t.Fatalf("expected authentik server not to be called")
		}))
		defer akServer.Close()

		nextCalled := false
		next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			nextCalled = true

			rw.WriteHeader(http.StatusAccepted)
		})

		config := &config.Config{Address: akServer.URL, KeepPrefix: "/test"}
		handler, _ := plugin.New(context.Background(), next, config, "test")

		req := httptest.NewRequest("GET", "http://example.com/other/test/users", nil)

		rw := httptest.NewRecorder()
		handler.ServeHTTP(rw, req)

		// check that the next handler was called
		if !nextCalled {
			t.Fatalf("expected next handler to be called")
		}

		// check that the response status code comes from the upstream
		expectedCode := http.StatusAccepted
		if rw.Code != expectedCode {
			t.Errorf("expected status %d, got %d", expectedCode, rw.Code)
		}
	})

	t.Run("unauthenticated request without redirect", func(t *testing.T) {
		akCalled := true
		akServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			akCalled = true

			// check that the forwarded host header is set
			expectedHost := "example.com"
			if req.Header.Get("X-Forwarded-Host") != expectedHost {
				t.Errorf("expected X-Forwarded-Host header to be %s, got %s", expectedHost, req.Header.Get("X-Forwarded-Host"))
			}

			// check that the original uri header is set
			expectedURI := "http://example.com/test/users"
			if req.Header.Get("X-Original-URI") != expectedURI {
				t.Errorf("expected X-Original-URI header to be %s, got %s", expectedURI, req.Header.Get("X-Original-URI"))
			}

			rw.WriteHeader(http.StatusUnauthorized)
		}))
		defer akServer.Close()

		next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			// check that the next handler was not called
			t.Fatalf("expected next handler not to be called")
		})

		config := &config.Config{Address: akServer.URL, KeepPrefix: "/test", UnauthorizedStatusCode: http.StatusForbidden}
		handler, _ := plugin.New(context.Background(), next, config, "test")

		req := httptest.NewRequest("GET", "http://example.com/test/users", nil)

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
			if req.Header.Get("X-Forwarded-Host") != expectedHost {
				t.Errorf("expected X-Forwarded-Host header to be %s, got %s", expectedHost, req.Header.Get("X-Forwarded-Host"))
			}

			// check that the original uri header is set
			expectedURI := "http://example.com/test/users"
			if req.Header.Get("X-Original-URI") != expectedURI {
				t.Errorf("expected X-Original-URI header to be %s, got %s", expectedURI, req.Header.Get("X-Original-URI"))
			}

			rw.WriteHeader(http.StatusUnauthorized)
		}))
		defer akServer.Close()

		next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			// check that the next handler was not called
			t.Fatalf("expected next handler not to be called")
		})

		config := &config.Config{Address: akServer.URL, KeepPrefix: "/test", UnauthorizedStatusCode: http.StatusMovedPermanently}
		handler, _ := plugin.New(context.Background(), next, config, "test")

		req := httptest.NewRequest("GET", "http://example.com/test/users", nil)

		rw := httptest.NewRecorder()
		handler.ServeHTTP(rw, req)

		// check that the authentik server was called
		if !akCalled {
			t.Fatalf("expected authentik server to be called")
		}

		// check that the response status code is the one configured
		expectedCode := http.StatusMovedPermanently
		if rw.Code != expectedCode {
			t.Errorf("expected status %d, got %d", expectedCode, rw.Code)
		}

		// check that the location header starts authorization flow
		expectedLocation := "http://example.com/test/outpost.goauthentik.go/start?rd=http%3A%2F%2Fexample.com%2Ftest%2Fusers"
		if rw.Header().Get("Location") != expectedLocation {
			t.Errorf("expected location header to be %s, got %s", expectedLocation, rw.Header().Get("Location"))
		}
	})

	t.Run("authenticated request", func(t *testing.T) {
		akCalled := true
		akServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			akCalled = true

			// check that the forwarded host header is set
			expectedHost := "example.com"
			if req.Header.Get("X-Forwarded-Host") != expectedHost {
				t.Errorf("expected X-Forwarded-Host header to be %s, got %s", expectedHost, req.Header.Get("X-Forwarded-Host"))
			}

			// check that the original uri header is set
			expectedURI := "http://example.com/test/users"
			if req.Header.Get("X-Original-URI") != expectedURI {
				t.Errorf("expected X-Original-URI header to be %s, got %s", expectedURI, req.Header.Get("X-Original-URI"))
			}

			// check that the authentication cookie is set
			expectedCookie := "authentik_proxy_user=testuser"
			if req.Header.Get("Cookie") != expectedCookie {
				t.Errorf("expected Cookie header to be %s, got %s", expectedCookie, req.Header.Get("Cookie"))
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
			if req.Header.Get("X-Authentik-User") != expectedUser {
				t.Errorf("expected X-Authentik-User header to be %s, got %s", expectedUser, req.Header.Get("X-Authentik-User"))
			}

			// check that authentik cookies were added to the request
			if _, err := req.Cookie("authentik_proxy_user"); err != nil {
				t.Error("expected authentik_proxy_user cookie to be added to request")
			}

			rw.WriteHeader(http.StatusAccepted)
		})

		config := &config.Config{Address: akServer.URL, KeepPrefix: "/test"}
		handler, _ := plugin.New(context.Background(), next, config, "test")

		req := httptest.NewRequest("GET", "http://example.com/test/users", nil)
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
		if rw.Code != expectedCode {
			t.Errorf("expected status %d, got %d", expectedCode, rw.Code)
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

		config := &config.Config{
			Address:                akServer.URL,
			KeepPrefix:             "/test",
			UnauthorizedStatusCode: http.StatusUnauthorized,
			UnauthorizedPathStatusCodes: map[string]uint{
				"/test/users/wrong": http.StatusNotFound,
			},
		}
		handler, _ := plugin.New(context.Background(), nil, config, "test")

		req := httptest.NewRequest("GET", "http://example.com/test/users", nil)

		rw := httptest.NewRecorder()
		handler.ServeHTTP(rw, req)

		// check that the response status code is the one configured
		expectedCode := http.StatusUnauthorized
		if rw.Code != expectedCode {
			t.Errorf("expected status %d, got %d", expectedCode, rw.Code)
		}
	})

	t.Run("unauthenticated request with path status code", func(t *testing.T) {
		akServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			rw.WriteHeader(http.StatusUnauthorized)
		}))
		defer akServer.Close()

		config := &config.Config{
			Address:                akServer.URL,
			KeepPrefix:             "/test",
			UnauthorizedStatusCode: http.StatusUnauthorized,
			UnauthorizedPathStatusCodes: map[string]uint{
				"/test/users": http.StatusNotFound,
			},
		}
		handler, _ := plugin.New(context.Background(), nil, config, "test")

		req := httptest.NewRequest("GET", "http://example.com/test/users", nil)

		rw := httptest.NewRecorder()
		handler.ServeHTTP(rw, req)

		// check that the response status code is the one configured
		expectedCode := http.StatusNotFound
		if rw.Code != expectedCode {
			t.Errorf("expected status %d, got %d", expectedCode, rw.Code)
		}
	})

	t.Run("unauthenticated request with multiple path status codes", func(t *testing.T) {
		akServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			rw.WriteHeader(http.StatusUnauthorized)
		}))
		defer akServer.Close()

		config := &config.Config{
			Address:                akServer.URL,
			KeepPrefix:             "/test",
			UnauthorizedStatusCode: http.StatusUnauthorized,
			UnauthorizedPathStatusCodes: map[string]uint{
				"/test/users":   http.StatusBadRequest,
				"/test/users/?": http.StatusNotFound,
			},
		}
		handler, _ := plugin.New(context.Background(), nil, config, "test")

		req := httptest.NewRequest("GET", "http://example.com/test/users", nil)

		rw := httptest.NewRecorder()
		handler.ServeHTTP(rw, req)

		// check that the response status code is the one configured
		expectedCode := http.StatusNotFound
		if rw.Code != expectedCode {
			t.Errorf("expected status %d, got %d", expectedCode, rw.Code)
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
		expectedCode := http.StatusTeapot
		if rw.Code != expectedCode {
			t.Errorf("expected status %d, got %d", expectedCode, rw.Code)
		}

		// check that the response body comes from the authentik server
		expectedBody := "i'm a teapot"
		if rw.Body.String() != expectedBody {
			t.Errorf("expected content to be %s, got %s", expectedBody, rw.Body.String())
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
		expectedCode := http.StatusNotFound
		if rw.Code != expectedCode {
			t.Errorf("expected status %d, got %d", expectedCode, rw.Code)
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
		expectedCode := http.StatusOK
		if rw.Code != expectedCode {
			t.Errorf("expected status %d, got %d", expectedCode, rw.Code)
		}

		// check that the response body comes from the authentik server
		expectedBody := "i'm a teapot"
		if rw.Body.String() != expectedBody {
			t.Errorf("expected content to be %s, got %s", expectedBody, rw.Body.String())
		}
	})
}
