package config

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/xabinapal/traefik-authentik-forward-plugin/internal/authentik"
	"github.com/xabinapal/traefik-authentik-forward-plugin/internal/httpclient"
)

const (
	minValidTLSVersion = 10
	maxValidTLSVersion = 13
)

const (
	DefaultUnauthorizedStatusCode = http.StatusUnauthorized
	DefaultRedirectStatusCode     = http.StatusFound

	DefaultTimeout               = "0s"
	DefaultTLSMinVersion         = 12
	DefaultTLSMaxVersion         = 13
	DefaultTLSInsecureSkipVerify = false
)

//nolint:gochecknoglobals
var (
	DefaultUnauthorizedPaths = []string{"^/.*$"}
	DefaultRedirectPaths     = []string{}
)

func (c *Config) Parse() (*PluginConfig, error) {
	var err error
	var authentikCfg *authentik.Config
	var httpClientCfg *httpclient.Config

	authentikCfg, err = parseAuthentikConfig(c)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrConfigParse, err)
	}

	httpClientCfg, err = parseHTTPClientConfig(c)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrConfigParse, err)
	}

	return &PluginConfig{
		Authentik:  authentikCfg,
		HTTPClient: httpClientCfg,
	}, nil
}

func parseAuthentikConfig(c *Config) (*authentik.Config, error) {
	cfg := &authentik.Config{}

	if c.Address == "" {
		return nil, fmt.Errorf("%w: address is required", ErrConfigParse)
	}
	cfg.Address = c.Address

	// set default unauthorized status code
	if c.UnauthorizedStatusCode == 0 {
		c.UnauthorizedStatusCode = DefaultUnauthorizedStatusCode
	}

	cfg.UnauthorizedStatusCode = int(c.UnauthorizedStatusCode)

	// set default redirect status code
	if c.RedirectStatusCode == 0 {
		c.RedirectStatusCode = DefaultRedirectStatusCode
	}

	cfg.RedirectStatusCode = int(c.RedirectStatusCode)

	// parse unauthorized paths
	cfg.UnauthorizedPaths = make([]*regexp.Regexp, 0, len(c.UnauthorizedPaths))
	for idx, path := range c.UnauthorizedPaths {
		re, err := regexp.Compile(path)
		if err != nil {
			return nil, fmt.Errorf("unauthorizedPaths[%d] is not valid: %w", idx, err)
		}

		cfg.UnauthorizedPaths = append(cfg.UnauthorizedPaths, re)
	}

	// parse redirect paths
	cfg.RedirectPaths = make([]*regexp.Regexp, 0, len(c.RedirectPaths))
	for idx, path := range c.RedirectPaths {
		re, err := regexp.Compile(path)
		if err != nil {
			return nil, fmt.Errorf("redirectPaths[%d] is not valid: %w", idx, err)
		}

		cfg.RedirectPaths = append(cfg.RedirectPaths, re)
	}

	return cfg, nil
}

func parseHTTPClientConfig(c *Config) (*httpclient.Config, error) {
	cfg := &httpclient.Config{}

	// parse timeout
	var timeout time.Duration
	if c.Timeout == "" {
		c.Timeout = DefaultTimeout
	}

	timeout, err := time.ParseDuration(c.Timeout)
	if err != nil {
		return nil, fmt.Errorf("timeout is not valid: %w", err)
	}

	cfg.Timeout = timeout

	// parse tls ca config
	cfg.TLS.CA = c.TLS.CA

	// parse tls client cert config
	cfg.TLS.Cert = c.TLS.Cert

	// parse tls client key config
	cfg.TLS.Key = c.TLS.Key

	// parse tls min version
	if c.TLS.MinVersion == 0 {
		c.TLS.MinVersion = DefaultTLSMinVersion
	} else if c.TLS.MinVersion < minValidTLSVersion || c.TLS.MinVersion > maxValidTLSVersion {
		return nil, errors.New("tls.minVersion is not valid")
	}

	cfg.TLS.MinVersion = c.TLS.MinVersion - minValidTLSVersion + tls.VersionTLS10

	// parse tls max version
	if c.TLS.MaxVersion == 0 {
		c.TLS.MaxVersion = DefaultTLSMaxVersion
	} else if c.TLS.MaxVersion < minValidTLSVersion || c.TLS.MaxVersion > maxValidTLSVersion {
		return nil, errors.New("tls.maxVersion is not valid")
	}

	cfg.TLS.MaxVersion = c.TLS.MaxVersion - minValidTLSVersion + tls.VersionTLS10

	if cfg.TLS.MinVersion > cfg.TLS.MaxVersion {
		return nil, errors.New("tls.minVersion cannot be higher than tls.maxVersion")
	}

	// parse tls insecure skip verify
	cfg.TLS.InsecureSkipVerify = c.TLS.InsecureSkipVerify

	return cfg, nil
}
