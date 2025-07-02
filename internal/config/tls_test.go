package config_test

import (
	"crypto/tls"
	"errors"
	"testing"

	"github.com/xabinapal/traefik-authentik-forward-plugin/internal/config"
)

func TestParse_TLS(t *testing.T) {
	t.Run("with default config", func(t *testing.T) {
		config := &config.RawTLSConfig{}

		pc, err := config.Parse()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// check that the ca is the default one
		if pc.CA != "" {
			t.Errorf("expected ca to be empty, got %s", pc.CA)
		}

		// check that the cert is the default one
		if pc.Cert != "" {
			t.Errorf("expected cert to be empty, got %s", pc.Cert)
		}

		// check that the key is the default one
		if pc.Key != "" {
			t.Errorf("expected key to be empty, got %s", pc.Key)
		}

		// check that the min version is the default one
		var expectedMinVersion uint16 = tls.VersionTLS12
		if pc.MinVersion != expectedMinVersion {
			t.Errorf("expected MinVersion %d, got %d", expectedMinVersion, pc.MinVersion)
		}

		// check that the max version is the default one
		var expectedMaxVersion uint16 = tls.VersionTLS13
		if pc.MaxVersion != expectedMaxVersion {
			t.Errorf("expected MaxVersion %d, got %d", expectedMaxVersion, pc.MaxVersion)
		}

		// check that the insecure skip verify is the default one
		expectedInsecureSkipVerify := false
		if pc.InsecureSkipVerify != expectedInsecureSkipVerify {
			t.Errorf("expected InsecureSkipVerify %v, got %v", expectedInsecureSkipVerify, pc.InsecureSkipVerify)
		}
	})

	t.Run("with custom parameters", func(t *testing.T) {
		config := &config.RawTLSConfig{
			CA:                 "/path/to/ca.pem",
			Cert:               "/path/to/cert.pem",
			Key:                "/path/to/key.pem",
			MinVersion:         11,
			MaxVersion:         12,
			InsecureSkipVerify: true,
		}

		pc, err := config.Parse()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// check that the ca is the configured one
		expectedCA := "/path/to/ca.pem"
		if pc.CA != expectedCA {
			t.Errorf("expected ca %s, got %s", expectedCA, pc.CA)
		}

		// check that the cert is the configured one
		expectedCert := "/path/to/cert.pem"
		if pc.Cert != expectedCert {
			t.Errorf("expected cert %s, got %s", expectedCert, pc.Cert)
		}

		// check that the key is the configured one
		expectedKey := "/path/to/key.pem"
		if pc.Key != expectedKey {
			t.Errorf("expected key %s, got %s", expectedKey, pc.Key)
		}

		// check that the min version is the configured one
		var expectedMinVersion uint16 = tls.VersionTLS11
		if pc.MinVersion != expectedMinVersion {
			t.Errorf("expected min version %d, got %d", expectedMinVersion, pc.MinVersion)
		}

		// check that the max version is the configured one
		var expectedMaxVersion uint16 = tls.VersionTLS12
		if pc.MaxVersion != expectedMaxVersion {
			t.Errorf("expected max version %d, got %d", expectedMaxVersion, pc.MaxVersion)
		}

		// check that the insecure skip verify is the configured one
		expectedInsecureSkipVerify := true
		if pc.InsecureSkipVerify != expectedInsecureSkipVerify {
			t.Errorf("expected insecure skip verify %v, got %v", expectedInsecureSkipVerify, pc.InsecureSkipVerify)
		}
	})

	t.Run("with invalid parameters", func(t *testing.T) {
		tests := []struct {
			name          string
			config        *config.RawTLSConfig
			expectedError string
		}{
			{
				name: "invalid min version",
				config: &config.RawTLSConfig{
					MinVersion: 99,
				},
				expectedError: "tls.minVersion is not valid",
			},
			{
				name: "invalid max version",
				config: &config.RawTLSConfig{
					MaxVersion: 99,
				},
				expectedError: "tls.maxVersion is not valid",
			},
			{
				name: "min version higher than max version",
				config: &config.RawTLSConfig{
					MinVersion: 13,
					MaxVersion: 12,
				},
				expectedError: "tls.minVersion cannot be higher than tls.maxVersion",
			},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				_, err := test.config.Parse()
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
