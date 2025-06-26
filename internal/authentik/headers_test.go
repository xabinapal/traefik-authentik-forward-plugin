package authentik_test

import (
	"net/http"
	"slices"
	"testing"

	"github.com/xabinapal/traefik-authentik-forward-plugin/internal/authentik"
)

func TestGetHeaders(t *testing.T) {
	t.Run("header filtering", func(t *testing.T) {
		// Add headers to the response
		responseHeaders := http.Header{
			"Content-Type":      []string{"application/json"},
			"Cache-Control":     []string{"no-cache"},
			"X-Custom-Header":   []string{"value"},
			"X-Authentik":       []string{"user123"},
			"X-AuthentikUser":   []string{"user123"},
			"X-Authentik-User":  []string{"user456"},
			"X-Authentik-Email": []string{"user@example.com"},
		}

		// Create a mock response
		resp := &http.Response{
			Header: responseHeaders,
		}

		// Call the function
		result := authentik.GetHeaders(resp)

		// Check headers
		expectedHeaders := http.Header{
			"X-Authentik-User":  []string{"user456"},
			"X-Authentik-Email": []string{"user@example.com"},
		}

		actualHeaders := make([]string, 0, len(result))
		for header := range result {
			actualHeaders = append(actualHeaders, header)
		}

		// Check each expected header is present
		for key := range expectedHeaders {
			found := slices.Contains(actualHeaders, key)
			if !found {
				t.Errorf("expected header %s not found in result", key)
			}
		}

		// Check that no unexpected headers are present
		for key := range result {
			if !slices.Contains(actualHeaders, key) {
				t.Errorf("unexpected header %s found in result", key)
			}
		}
	})
}
