package config

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/xabinapal/traefik-authentik-forward-plugin/internal/authentik"
	"github.com/xabinapal/traefik-authentik-forward-plugin/internal/httpclient"
)

const (
	minValidTLSVersion = 10
	maxValidTLSVersion = 13
)

const (
	DefaultCookiePolicy           = "lax"
	DefaultUnauthorizedStatusCode = http.StatusUnauthorized
	DefaultRedirectStatusCode     = http.StatusFound

	DefaultTimeout               = "0s"
	DefaultTLSMinVersion         = 12
	DefaultTLSMaxVersion         = 13
	DefaultTLSInsecureSkipVerify = false
)

//nolint:gochecknoglobals
var (
	DefaultSkippedPaths      = []string{}
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

	// parse cookie policy
	if c.CookiePolicy == "" {
		c.CookiePolicy = DefaultCookiePolicy
	}

	c.CookiePolicy = strings.ToLower(c.CookiePolicy)

	switch c.CookiePolicy {
	case "none":
		cfg.CookiePolicy = http.SameSiteNoneMode
	case "lax":
		cfg.CookiePolicy = http.SameSiteLaxMode
	case "strict":
		cfg.CookiePolicy = http.SameSiteStrictMode
	default:
		return nil, fmt.Errorf("cookiePolicy is not valid: %s", c.CookiePolicy)
	}

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

	// parse skipped paths
	if skippedPaths, err := parsePathRegexes("skippedPaths", c.SkippedPaths); err != nil {
		return nil, err
	} else {
		cfg.SkippedPaths = skippedPaths
	}

	// parse unauthorized paths
	if unauthorizedPaths, err := parsePathRegexes("unauthorizedPaths", c.UnauthorizedPaths); err != nil {
		return nil, err
	} else {
		cfg.UnauthorizedPaths = unauthorizedPaths
	}

	// parse redirect paths
	if redirectPaths, err := parsePathRegexes("redirectPaths", c.RedirectPaths); err != nil {
		return nil, err
	} else {
		cfg.RedirectPaths = redirectPaths
	}

	return cfg, nil
}

func parsePathRegexes(name string, paths []string) ([]*regexp.Regexp, error) {
	pathRegexes := make([]*regexp.Regexp, 0, len(paths))
	for idx, path := range paths {
		re, err := regexp.Compile(path)
		if err != nil {
			return nil, fmt.Errorf("%s[%d] is not valid: %w", name, idx, err)
		}

		pathRegexes = append(pathRegexes, re)
	}

	return pathRegexes, nil
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
