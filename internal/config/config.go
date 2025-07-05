package config

import (
	"github.com/xabinapal/traefik-authentik-forward-plugin/internal/authentik"
	"github.com/xabinapal/traefik-authentik-forward-plugin/internal/httpclient"
)

type Config struct {
	// The address of the Authentik server to forward requests to.
	Address string `json:"address"`

	// The status code to return when the request is unauthorized.
	UnauthorizedStatusCode uint16 `json:"unauthorizedStatusCode,omitempty"`

	// The status code to return when unauthorized requests must be redirected.
	RedirectStatusCode uint16 `json:"redirectStatusCode,omitempty"`

	// List of path regexes that will be treated as unauthorized.
	UnauthorizedPaths []string `json:"unauthorizedPaths,omitempty"`

	// List of path regexes that will be treated as redirections.
	RedirectPaths []string `json:"redirectPaths,omitempty"`

	// Connection timeout duration as a string (e.g., "30s", "1m")
	Timeout string `json:"timeout,omitempty"`

	// TLS configuration
	TLS TLSConfig `json:"tls,omitempty"`
}

type TLSConfig struct {
	// Path to the CA certificate file
	CA string `json:"ca,omitempty"`

	// Path to the client certificate file
	Cert string `json:"cert,omitempty"`

	// Path to the client private key file
	Key string `json:"key,omitempty"`

	// Minimum TLS version (10=1.0, 11=1.1, 12=1.2, 13=1.3)
	MinVersion uint16 `json:"minVersion,omitempty"`

	// Maximum TLS version (10=1.0, 11=1.1, 12=1.2, 13=1.3)
	MaxVersion uint16 `json:"maxVersion,omitempty"`

	// Skip certificate verification
	InsecureSkipVerify bool `json:"insecureSkipVerify,omitempty"`
}

type PluginConfig struct {
	Authentik  *authentik.Config
	HTTPClient *httpclient.Config
}
