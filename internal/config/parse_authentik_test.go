package config_test

import (
	"net/http"
	"testing"

	"github.com/xabinapal/traefik-authentik-forward-plugin/internal/config"
)

func TestParse_CookiePolicy(t *testing.T) {
	t.Run("with empty value", func(t *testing.T) {
		config := config.Config{
			Address:      "https://authentik.example.com",
			CookiePolicy: "",
		}

		pc, err := config.Parse()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		expectedCookiePolicy := http.SameSiteLaxMode
		if pc.Authentik.CookiePolicy != expectedCookiePolicy {
			t.Errorf("expected cookie policy %v, got %v", expectedCookiePolicy, pc.Authentik.CookiePolicy)
		}
	})

	t.Run("with valid value", func(t *testing.T) {
		config := config.Config{
			Address:      "https://authentik.example.com",
			CookiePolicy: "strict",
		}

		pc, err := config.Parse()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		expectedCookiePolicy := http.SameSiteStrictMode
		if pc.Authentik.CookiePolicy != expectedCookiePolicy {
			t.Errorf("expected cookie policy %v, got %v", expectedCookiePolicy, pc.Authentik.CookiePolicy)
		}
	})

	t.Run("with invalid value", func(t *testing.T) {
		config := config.Config{
			Address:      "https://authentik.example.com",
			CookiePolicy: "invalid",
		}

		_, err := config.Parse()
		if err == nil {
			t.Fatal("expected error for invalid cookie policy, got none")
		}
	})
}
