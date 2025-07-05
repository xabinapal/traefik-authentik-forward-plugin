package config_test

import (
	"crypto/tls"
	"errors"
	"testing"
	"time"

	"github.com/xabinapal/traefik-authentik-forward-plugin/internal/config"
)

func TestParse_Timeout(t *testing.T) {
	t.Run("with empty value", func(t *testing.T) {
		config := config.Config{
			Address: "https://authentik.example.com",
			Timeout: "",
		}

		pc, err := config.Parse()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		expectedTimeout := time.Duration(0)
		if pc.HTTPClient.Timeout != expectedTimeout {
			t.Errorf("expected timeout %v, got %v", expectedTimeout, pc.HTTPClient.Timeout)
		}
	})

	t.Run("with valid value", func(t *testing.T) {
		config := config.Config{
			Address: "https://authentik.example.com",
			Timeout: "30s",
		}

		pc, err := config.Parse()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		expectedTimeout := 30 * time.Second
		if pc.HTTPClient.Timeout != expectedTimeout {
			t.Errorf("expected timeout %v, got %v", expectedTimeout, pc.HTTPClient.Timeout)
		}
	})

	t.Run("with invalid value", func(t *testing.T) {
		config := config.Config{
			Address: "https://authentik.example.com",
			Timeout: "invalid",
		}

		_, err := config.Parse()
		if err == nil {
			t.Fatal("expected error for invalid timeout, got none")
		}
	})
}

func TestParse_TLS(t *testing.T) {
	t.Run("with default config", func(t *testing.T) {
		config := &config.Config{
			Address: "https://authentik.example.com",
		}

		pc, err := config.Parse()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// check that the ca is the default one
		if pc.HTTPClient.TLS.CA != "" {
			t.Errorf("expected ca to be empty, got %s", pc.HTTPClient.TLS.CA)
		}

		// check that the cert is the default one
		if pc.HTTPClient.TLS.Cert != "" {
			t.Errorf("expected cert to be empty, got %s", pc.HTTPClient.TLS.Cert)
		}

		// check that the key is the default one
		if pc.HTTPClient.TLS.Key != "" {
			t.Errorf("expected key to be empty, got %s", pc.HTTPClient.TLS.Key)
		}

		// check that the min version is the default one
		var expectedMinVersion uint16 = tls.VersionTLS12
		if pc.HTTPClient.TLS.MinVersion != expectedMinVersion {
			t.Errorf("expected MinVersion %d, got %d", expectedMinVersion, pc.HTTPClient.TLS.MinVersion)
		}

		// check that the max version is the default one
		var expectedMaxVersion uint16 = tls.VersionTLS13
		if pc.HTTPClient.TLS.MaxVersion != expectedMaxVersion {
			t.Errorf("expected MaxVersion %d, got %d", expectedMaxVersion, pc.HTTPClient.TLS.MaxVersion)
		}

		// check that the insecure skip verify is the default one
		expectedInsecureSkipVerify := false
		if pc.HTTPClient.TLS.InsecureSkipVerify != expectedInsecureSkipVerify {
			t.Errorf("expected InsecureSkipVerify %v, got %v", expectedInsecureSkipVerify, pc.HTTPClient.TLS.InsecureSkipVerify)
		}
	})

	t.Run("with custom parameters", func(t *testing.T) {
		config := &config.Config{
			Address: "https://authentik.example.com",
			TLS: config.TLSConfig{
				CA:                 "/path/to/ca.pem",
				Cert:               "/path/to/cert.pem",
				Key:                "/path/to/key.pem",
				MinVersion:         11,
				MaxVersion:         12,
				InsecureSkipVerify: true,
			},
		}

		pc, err := config.Parse()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// check that the ca is the configured one
		expectedCA := "/path/to/ca.pem"
		if pc.HTTPClient.TLS.CA != expectedCA {
			t.Errorf("expected ca %s, got %s", expectedCA, pc.HTTPClient.TLS.CA)
		}

		// check that the cert is the configured one
		expectedCert := "/path/to/cert.pem"
		if pc.HTTPClient.TLS.Cert != expectedCert {
			t.Errorf("expected cert %s, got %s", expectedCert, pc.HTTPClient.TLS.Cert)
		}

		// check that the key is the configured one
		expectedKey := "/path/to/key.pem"
		if pc.HTTPClient.TLS.Key != expectedKey {
			t.Errorf("expected key %s, got %s", expectedKey, pc.HTTPClient.TLS.Key)
		}

		// check that the min version is the configured one
		var expectedMinVersion uint16 = tls.VersionTLS11
		if pc.HTTPClient.TLS.MinVersion != expectedMinVersion {
			t.Errorf("expected min version %d, got %d", expectedMinVersion, pc.HTTPClient.TLS.MinVersion)
		}

		// check that the max version is the configured one
		var expectedMaxVersion uint16 = tls.VersionTLS12
		if pc.HTTPClient.TLS.MaxVersion != expectedMaxVersion {
			t.Errorf("expected max version %d, got %d", expectedMaxVersion, pc.HTTPClient.TLS.MaxVersion)
		}

		// check that the insecure skip verify is the configured one
		expectedInsecureSkipVerify := true
		if pc.HTTPClient.TLS.InsecureSkipVerify != expectedInsecureSkipVerify {
			t.Errorf("expected insecure skip verify %v, got %v", expectedInsecureSkipVerify, pc.HTTPClient.TLS.InsecureSkipVerify)
		}
	})

	t.Run("with invalid parameters", func(t *testing.T) {
		tests := []struct {
			name          string
			config        *config.TLSConfig
			expectedError string
		}{
			{
				name: "invalid min version",
				config: &config.TLSConfig{
					MinVersion: 99,
				},
				expectedError: "tls.minVersion is not valid",
			},
			{
				name: "invalid max version",
				config: &config.TLSConfig{
					MaxVersion: 99,
				},
				expectedError: "tls.maxVersion is not valid",
			},
			{
				name: "min version higher than max version",
				config: &config.TLSConfig{
					MinVersion: 13,
					MaxVersion: 12,
				},
				expectedError: "tls.minVersion cannot be higher than tls.maxVersion",
			},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				cfg := &config.Config{
					Address: "https://authentik.example.com",
					TLS:     *test.config,
				}

				_, err := cfg.Parse()
				if err == nil {
					t.Fatal("expected error, got none")
				}

				if !errors.Is(err, config.ErrConfigParse) {
					t.Errorf("expected error %q, got %q", config.ErrConfigParse, err)
				}

				expectedError := "invalid config: " + test.expectedError
				if err.Error() != expectedError {
					t.Errorf("expected error %q, got %q", expectedError, err.Error())
				}
			})
		}
	})
}
