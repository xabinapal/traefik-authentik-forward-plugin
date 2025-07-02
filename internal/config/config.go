package config

import (
	"fmt"
	"net/http"
	"regexp"
	"time"
)

const (
	DefaultTimeout                = "0s"
	DefaultUnauthorizedStatusCode = http.StatusUnauthorized
	DefaultRedirectStatusCode     = http.StatusFound
)

//nolint:gochecknoglobals
var (
	DefaultUnauthorizedPaths = []string{"^/.*$"}
	DefaultRedirectPaths     = []string{}
)

type RawConfig struct {
	// The address of the Authentik server to forward requests to.
	Address string `json:"address"`

	// Connection timeout duration as a string (e.g., "30s", "1m")
	Timeout string `json:"timeout,omitempty"`

	// TLS configuration
	TLS *RawTLSConfig `json:"tls,omitempty"`

	// The status code to return when the request is unauthorized.
	UnauthorizedStatusCode uint16 `json:"unauthorizedStatusCode,omitempty"`

	// The status code to return when unauthorized requests must be redirected.
	RedirectStatusCode uint16 `json:"redirectStatusCode,omitempty"`

	// List of path regexes that will be treated as unauthorized.
	UnauthorizedPaths []string `json:"unauthorizedPaths,omitempty"`

	// List of path regexes that will be treated as redirections.
	RedirectPaths []string `json:"redirectPaths,omitempty"`
}

type Config struct {
	RawConfig
	Timeout                time.Duration
	TLS                    *TLSConfig
	UnauthorizedStatusCode int
	RedirectStatusCode     int
	UnauthorizedPaths      []*regexp.Regexp
	RedirectPaths          []*regexp.Regexp
}

func (c *RawConfig) Parse() (*Config, error) {
	var err error

	if c.Address == "" {
		return nil, fmt.Errorf("%w: address is required", ErrConfigParse)
	}

	// parse timeout
	var timeout time.Duration
	if c.Timeout == "" {
		c.Timeout = DefaultTimeout
	}

	timeout, err = time.ParseDuration(c.Timeout)
	if err != nil {
		return nil, fmt.Errorf("%w: timeout is not valid: %w", ErrConfigParse, err)
	}

	c.Timeout = timeout.String()

	// parse tls config
	var tlsConfig *TLSConfig
	if c.TLS == nil {
		c.TLS = &RawTLSConfig{}
	}

	tlsConfig, err = c.TLS.Parse()
	if err != nil {
		return nil, fmt.Errorf("%w: tls configuration is not valid: %w", ErrConfigParse, err)
	}

	// set default unauthorized status code
	if c.UnauthorizedStatusCode == 0 {
		c.UnauthorizedStatusCode = DefaultUnauthorizedStatusCode
	}

	// set default redirect status code
	if c.RedirectStatusCode == 0 {
		c.RedirectStatusCode = DefaultRedirectStatusCode
	}

	// parse unauthorized paths
	unauthorizedPaths := make([]*regexp.Regexp, 0, len(c.UnauthorizedPaths))
	for _, path := range c.UnauthorizedPaths {
		re, err := regexp.Compile(path)
		if err != nil {
			return nil, fmt.Errorf("%w: unauthorizedPaths[%s] is not valid: %w", ErrConfigParse, path, err)
		}

		unauthorizedPaths = append(unauthorizedPaths, re)
	}

	// parse redirect paths
	redirectPaths := make([]*regexp.Regexp, 0, len(c.RedirectPaths))
	for _, path := range c.RedirectPaths {
		re, err := regexp.Compile(path)
		if err != nil {
			return nil, fmt.Errorf("%w: redirectPaths[%s] is not valid: %w", ErrConfigParse, path, err)
		}

		redirectPaths = append(redirectPaths, re)
	}

	// create config
	pc := &Config{
		RawConfig:              *c,
		Timeout:                timeout,
		TLS:                    tlsConfig,
		UnauthorizedStatusCode: int(c.UnauthorizedStatusCode),
		RedirectStatusCode:     int(c.RedirectStatusCode),
		UnauthorizedPaths:      unauthorizedPaths,
		RedirectPaths:          redirectPaths,
	}

	return pc, nil
}

func (c *Config) GetUnauthorizedStatusCode(path string) int {
	var longestMatch *regexp.Regexp
	var longestMatchLength int
	var longestMatchStatusCode int

	for _, re := range c.UnauthorizedPaths {
		if re.MatchString(path) {
			l := len(re.String())
			if l > longestMatchLength {
				longestMatch = re
				longestMatchLength = l
				longestMatchStatusCode = c.UnauthorizedStatusCode
			}
		}
	}

	for _, re := range c.RedirectPaths {
		if re.MatchString(path) {
			l := len(re.String())
			if l > longestMatchLength {
				longestMatch = re
				longestMatchLength = l
				longestMatchStatusCode = c.RedirectStatusCode
			}
		}
	}

	if longestMatch != nil {
		return longestMatchStatusCode
	}

	return http.StatusOK
}
