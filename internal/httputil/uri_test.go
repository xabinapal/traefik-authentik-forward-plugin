package httputil_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xabinapal/traefik-authentik-forward-plugin/internal/httputil"
)

func TestGetRequestURI(t *testing.T) {
	t.Run("with invalid value", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)
		req.RequestURI = "://localhost"

		uri, err := httputil.GetRequestURI(req)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}

		// check that the uri is nil
		if uri != nil {
			t.Errorf("expected uri to be nil, got %v", uri)
		}
	})

	t.Run("with valid value", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "http://example.com/path?query=value", nil)

		uri, err := httputil.GetRequestURI(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// check that the scheme is the same as the request
		expectedScheme := "http"
		if uri.Scheme != expectedScheme {
			t.Errorf("expected scheme to be %s, got %s", expectedScheme, uri.Scheme)
		}

		// check that the host is the same as the request
		expectedHost := "example.com"
		if uri.Host != expectedHost {
			t.Errorf("expected host to be %s, got %s", expectedHost, uri.Host)
		}

		// check that the path is the same as the request
		expectedPath := "/path"
		if uri.Path != expectedPath {
			t.Errorf("expected path to be %s, got %s", expectedPath, uri.Path)
		}

		// check that the query is the same as the request
		expectedQuery := "query=value"
		if uri.RawQuery != expectedQuery {
			t.Errorf("expected query to be %s, got %s", expectedQuery, uri.RawQuery)
		}
	})

	t.Run("with no scheme", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "http://example.com/path", nil)
		req.RequestURI = "example.com/path"

		uri, err := httputil.GetRequestURI(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// check that the scheme is http
		expectedScheme := "http"
		if uri.Scheme != expectedScheme {
			t.Errorf("expected scheme to be %s, got %s", expectedScheme, uri.Scheme)
		}
	})

	t.Run("with http scheme", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "http://example.com/path", nil)

		uri, err := httputil.GetRequestURI(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// check that the scheme is http
		expectedScheme := "http"
		if uri.Scheme != expectedScheme {
			t.Errorf("expected scheme to be %s, got %s", expectedScheme, uri.Scheme)
		}
	})

	t.Run("with https scheme", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "https://example.com/path", nil)

		uri, err := httputil.GetRequestURI(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// check that the scheme is https
		expectedScheme := "https"
		if uri.Scheme != expectedScheme {
			t.Errorf("expected scheme to be %s, got %s", expectedScheme, uri.Scheme)
		}
	})
}
