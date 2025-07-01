package config

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
)

var ErrConfigParse = errors.New("invalid config")

type RawConfig struct {
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
}

type Config struct {
	RawConfig
	UnauthorizedStatusCode int
	RedirectStatusCode     int
	UnauthorizedPaths      []*regexp.Regexp
	RedirectPaths          []*regexp.Regexp
}

func (c *RawConfig) Parse() (*Config, error) {
	if c.Address == "" {
		return nil, fmt.Errorf("%w: address is required", ErrConfigParse)
	}

	if c.UnauthorizedStatusCode == 0 {
		c.UnauthorizedStatusCode = http.StatusUnauthorized
	}

	if c.RedirectStatusCode == 0 {
		c.RedirectStatusCode = http.StatusFound
	}

	unauthorizedPaths := make([]*regexp.Regexp, 0, len(c.UnauthorizedPaths))
	for _, path := range c.UnauthorizedPaths {
		re, err := regexp.Compile(path)
		if err != nil {
			return nil, fmt.Errorf("%w: unauthorizedPaths[%s] is not valid: %w", ErrConfigParse, path, err)
		}

		unauthorizedPaths = append(unauthorizedPaths, re)
	}

	redirectPaths := make([]*regexp.Regexp, 0, len(c.RedirectPaths))
	for _, path := range c.RedirectPaths {
		re, err := regexp.Compile(path)
		if err != nil {
			return nil, fmt.Errorf("%w: redirectPaths[%s] is not valid: %w", ErrConfigParse, path, err)
		}

		redirectPaths = append(redirectPaths, re)
	}

	pc := &Config{
		RawConfig:              *c,
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
