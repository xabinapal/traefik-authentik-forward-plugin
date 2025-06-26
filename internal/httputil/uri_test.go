package httputil_test

import (
	"net/http/httptest"
	"testing"

	"github.com/xabinapal/traefik-authentik-forward-plugin/internal/httputil"
)

func TestGetRequestURI(t *testing.T) {

	t.Run("request with no scheme", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/path?query=value", nil)

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

	t.Run("request with http scheme", func(t *testing.T) {
		req := httptest.NewRequest("GET", "http://example.com/path?query=value", nil)

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

	t.Run("request with https scheme", func(t *testing.T) {
		req := httptest.NewRequest("GET", "https://example.com/path?query=value", nil)

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

	t.Run("request parts", func(t *testing.T) {
		req := httptest.NewRequest("GET", "http://example.com/path?query=value", nil)

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
}
