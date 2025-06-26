package config_test

import (
	"net/http"
	"testing"

	"github.com/xabinapal/traefik-authentik-forward-plugin/internal/config"
)

func TestGetUnauthorizedStatusCode(t *testing.T) {
	t.Run("with global status code", func(t *testing.T) {
		config := config.Config{
			Address:                "https://authentik.example.com",
			UnauthorizedStatusCode: http.StatusUnauthorized,
		}

		pc, err := config.Parse()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		statusCode := pc.GetUnauthorizedStatusCode("/test/users")

		// check that the status code is the one configured
		expectedStatusCode := http.StatusUnauthorized
		if statusCode != expectedStatusCode {
			t.Errorf("expected status %d, got %d", expectedStatusCode, statusCode)
		}
	})

	t.Run("with unmatched path status code", func(t *testing.T) {
		config := config.Config{
			Address:                "https://authentik.example.com",
			UnauthorizedStatusCode: http.StatusUnauthorized,
			UnauthorizedPathStatusCodes: map[string]uint{
				"/test/users/wrong": http.StatusNotFound,
			},
		}

		pc, err := config.Parse()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		statusCode := pc.GetUnauthorizedStatusCode("/test/users")

		// check that the status code is the one configured
		expectedStatusCode := http.StatusUnauthorized
		if statusCode != expectedStatusCode {
			t.Errorf("expected status %d, got %d", expectedStatusCode, statusCode)
		}
	})

	t.Run("with matched path status code", func(t *testing.T) {
		config := config.Config{
			Address:                "https://authentik.example.com",
			UnauthorizedStatusCode: http.StatusUnauthorized,
			UnauthorizedPathStatusCodes: map[string]uint{
				"/test/users": http.StatusNotFound,
			},
		}

		pc, err := config.Parse()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		statusCode := pc.GetUnauthorizedStatusCode("/test/users")

		// check that the status code is the one configured
		expectedStatusCode := http.StatusNotFound
		if statusCode != expectedStatusCode {
			t.Errorf("expected status %d, got %d", expectedStatusCode, statusCode)
		}
	})

	t.Run("with multiple matched path status codes", func(t *testing.T) {
		config := config.Config{
			Address:                "https://authentik.example.com",
			UnauthorizedStatusCode: http.StatusUnauthorized,
			UnauthorizedPathStatusCodes: map[string]uint{
				"/test/users":   http.StatusBadRequest,
				"/test/users/?": http.StatusNotFound,
			},
		}

		pc, err := config.Parse()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		statusCode := pc.GetUnauthorizedStatusCode("/test/users")

		// check that the status code is the one configured
		expectedStatusCode := http.StatusNotFound
		if statusCode != expectedStatusCode {
			t.Errorf("expected status %d, got %d", expectedStatusCode, statusCode)
		}
	})
}
