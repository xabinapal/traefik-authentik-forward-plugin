package config_test

import (
	"testing"
	"time"

	"github.com/xabinapal/traefik-authentik-forward-plugin/internal/config"
)

func TestParse_CacheDuration(t *testing.T) {
	t.Run("with empty value", func(t *testing.T) {
		config := config.Config{
			Address: "https://authentik.example.com",
		}

		pc, err := config.Parse()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		expectedCacheDuration := time.Duration(0)
		if pc.Authentik.CacheDuration != expectedCacheDuration {
			t.Errorf("expected cache duration to be %v, got %v", expectedCacheDuration, pc.Authentik.CacheDuration)
		}
	})

	t.Run("with valid value", func(t *testing.T) {
		config := config.Config{
			Address:       "https://authentik.example.com",
			CacheDuration: "1h",
		}

		pc, err := config.Parse()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		expectedCacheDuration := time.Hour
		if pc.Authentik.CacheDuration != expectedCacheDuration {
			t.Errorf("expected cache duration to be %v, got %v", expectedCacheDuration, pc.Authentik.CacheDuration)
		}
	})

	t.Run("with invalid value", func(t *testing.T) {
		config := config.Config{
			Address:       "https://authentik.example.com",
			CacheDuration: "invalid",
		}

		_, err := config.Parse()
		if err == nil {
			t.Fatal("expected error for invalid cache duration, got none")
		}
	})
}
