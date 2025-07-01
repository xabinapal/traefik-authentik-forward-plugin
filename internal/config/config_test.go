package config_test

import (
	"net/http"
	"testing"

	"github.com/xabinapal/traefik-authentik-forward-plugin/internal/config"
)

func TestGetUnauthorizedStatusCode(t *testing.T) {
	t.Run("with no matching paths", func(t *testing.T) {
		config := config.RawConfig{
			Address:                "https://authentik.example.com",
			UnauthorizedStatusCode: http.StatusForbidden,
			RedirectStatusCode:     http.StatusFound,
			UnauthorizedPaths:      []string{"^/admin"},
			RedirectPaths:          []string{"^/login"},
		}

		pc, err := config.Parse()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		statusCode := pc.GetUnauthorizedStatusCode("/test")

		// check that the status code is ok
		expectedStatusCode := http.StatusOK
		if statusCode != expectedStatusCode {
			t.Errorf("expected status %d, got %d", expectedStatusCode, statusCode)
		}
	})

	t.Run("with matching unauthorized path", func(t *testing.T) {
		config := config.RawConfig{
			Address:                "https://authentik.example.com",
			UnauthorizedStatusCode: http.StatusForbidden,
			RedirectStatusCode:     http.StatusMovedPermanently,
			UnauthorizedPaths:      []string{"^/admin", "^/test"},
			RedirectPaths:          []string{"^/login"},
		}

		pc, err := config.Parse()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		statusCode := pc.GetUnauthorizedStatusCode("/test")

		// check that the status code is the configured unauthorized status
		expectedStatusCode := http.StatusForbidden
		if statusCode != expectedStatusCode {
			t.Errorf("expected status %d, got %d", expectedStatusCode, statusCode)
		}
	})

	t.Run("with matching redirect path", func(t *testing.T) {
		config := config.RawConfig{
			Address:                "https://authentik.example.com",
			UnauthorizedStatusCode: http.StatusForbidden,
			RedirectStatusCode:     http.StatusMovedPermanently,
			UnauthorizedPaths:      []string{"^/admin"},
			RedirectPaths:          []string{"^/login", "^/test"},
		}

		pc, err := config.Parse()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		statusCode := pc.GetUnauthorizedStatusCode("/test")

		// check that the status code is the configured redirect status
		expectedStatusCode := http.StatusMovedPermanently
		if statusCode != expectedStatusCode {
			t.Errorf("expected status %d, got %d", expectedStatusCode, statusCode)
		}
	})

	t.Run("with unauthorized path taking precedence over redirect path", func(t *testing.T) {
		config := config.RawConfig{
			Address:                "https://authentik.example.com",
			UnauthorizedStatusCode: http.StatusForbidden,
			RedirectStatusCode:     http.StatusMovedPermanently,
			UnauthorizedPaths:      []string{"^/admin", "^/test"},
			RedirectPaths:          []string{"^/.*"},
		}

		pc, err := config.Parse()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		statusCode := pc.GetUnauthorizedStatusCode("/test")

		// check that the status code is the configured unauthorized status
		expectedStatusCode := http.StatusForbidden
		if statusCode != expectedStatusCode {
			t.Errorf("expected status %d, got %d", expectedStatusCode, statusCode)
		}
	})

	t.Run("with unauthorized path taking precedence over redirect path", func(t *testing.T) {
		config := config.RawConfig{
			Address:                "https://authentik.example.com",
			UnauthorizedStatusCode: http.StatusForbidden,
			RedirectStatusCode:     http.StatusMovedPermanently,
			UnauthorizedPaths:      []string{"^/.*"},
			RedirectPaths:          []string{"^/admin", "^/test"},
		}

		pc, err := config.Parse()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		statusCode := pc.GetUnauthorizedStatusCode("/test")

		// check that the status code is the configured unauthorized status
		expectedStatusCode := http.StatusMovedPermanently
		if statusCode != expectedStatusCode {
			t.Errorf("expected status %d, got %d", expectedStatusCode, statusCode)
		}
	})

	t.Run("with regex pattern matching", func(t *testing.T) {
		config := config.RawConfig{
			Address:                "https://authentik.example.com",
			UnauthorizedStatusCode: http.StatusForbidden,
			RedirectStatusCode:     http.StatusFound,
			UnauthorizedPaths:      []string{`^/api/v\d+/admin`},
			RedirectPaths:          []string{`^/api/v\d+/login`},
		}

		pc, err := config.Parse()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Test unauthorized path regex matching
		statusCode := pc.GetUnauthorizedStatusCode("/api/v1/admin")
		expectedStatusCode := http.StatusForbidden
		if statusCode != expectedStatusCode {
			t.Errorf("expected status %d for unauthorized path, got %d", expectedStatusCode, statusCode)
		}

		// Test redirect path regex matching
		statusCode = pc.GetUnauthorizedStatusCode("/api/v2/login")
		expectedStatusCode = http.StatusFound
		if statusCode != expectedStatusCode {
			t.Errorf("expected status %d for redirect path, got %d", expectedStatusCode, statusCode)
		}

		// Test non-matching path
		statusCode = pc.GetUnauthorizedStatusCode("/api/v1/users")
		expectedStatusCode = http.StatusOK
		if statusCode != expectedStatusCode {
			t.Errorf("expected status %d for non-matching path, got %d", expectedStatusCode, statusCode)
		}
	})
}
