package authentik_test

import (
	"net/http"
	"slices"
	"testing"

	"github.com/xabinapal/traefik-authentik-forward-plugin/internal/authentik"
)

func TestGetCookies(t *testing.T) {
	t.Run("cookie filtering", func(t *testing.T) {
		// Create a mock request
		req, err := http.NewRequest("GET", "http://authentik.example.com", nil)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}

		// Add cookies to the request
		cookies := []*http.Cookie{
			{Name: "session_id", Value: "abc123"},
			{Name: "csrf_token", Value: "xyz789"},
			{Name: "authentik_proxy", Value: "value"},
			{Name: "authentik_proxy_", Value: "value"},
			{Name: "authentik_proxy_session1", Value: "session1"},
			{Name: "authentik_proxy_session2", Value: "session2"},
		}

		for _, cookie := range cookies {
			req.AddCookie(cookie)
		}

		// Call the function
		result := authentik.GetCookies(req)

		// Check cookies
		expectedCookies := []string{
			"authentik_proxy_session1",
			"authentik_proxy_session2",
		}

		actualCookies := make([]string, 0, len(result))
		for _, cookie := range result {
			actualCookies = append(actualCookies, cookie.Name)
		}

		// Check each expected cookie is present
		for _, expectedCookie := range expectedCookies {
			found := slices.Contains(actualCookies, expectedCookie)
			if !found {
				t.Errorf("expected cookie %s not found in actual cookies", expectedCookie)
			}
		}

		// Check that no unexpected cookies are present
		for _, cookie := range result {
			if !slices.Contains(expectedCookies, cookie.Name) {
				t.Errorf("unexpected cookie %s found in result", cookie.Name)
			}
		}
	})
}
