package config

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
)

var ErrConfigParse = errors.New("invalid config")

type RawConfig struct {
	// The address of the Authentik server to forward requests to.
	Address string `json:"address"`

	// The status code to return when the request is unauthorized.
	UnauthorizedStatusCode uint `json:"unauthorizedStatusCode,omitempty"`

	// Map of path regexes to customize their unauthorized status codes.
	UnauthorizedPathStatusCodes map[string]uint `json:"unauthorizedPathStatusCodes,omitempty"`
}

type Config struct {
	RawConfig
	UnauthorizedStatusCode      int
	UnauthorizedPathStatusCodes []*PathStatusCodesConfig
}

type PathStatusCodesConfig struct {
	PathRegex  *regexp.Regexp
	PathLength int
	StatusCode int
}

func (c *RawConfig) Parse() (*Config, error) {
	if c.Address == "" {
		return nil, errors.Join(
			ErrConfigParse,
			errors.New("address is required"),
		)
	}

	if c.UnauthorizedStatusCode == 0 {
		c.UnauthorizedStatusCode = http.StatusUnauthorized
	}

	pathStatusCodes := make([]*PathStatusCodesConfig, 0, len(c.UnauthorizedPathStatusCodes))
	for path, statusCode := range c.UnauthorizedPathStatusCodes {
		if statusCode == 0 {
			return nil, errors.Join(
				ErrConfigParse,
				fmt.Errorf("pathStatusCodes[%s] is not valid", path),
			)
		}

		re, err := regexp.Compile("^" + path)
		if err != nil {
			return nil, errors.Join(
				ErrConfigParse,
				fmt.Errorf("pathStatusCodes[%s] is not valid: %w", path, err),
			)
		}

		pathStatusCodes = append(pathStatusCodes, &PathStatusCodesConfig{
			PathRegex:  re,
			PathLength: len(path),
			StatusCode: int(statusCode),
		})
	}

	pc := &Config{
		RawConfig:                   *c,
		UnauthorizedStatusCode:      int(c.UnauthorizedStatusCode),
		UnauthorizedPathStatusCodes: pathStatusCodes,
	}

	return pc, nil
}

func (c *Config) GetUnauthorizedStatusCode(path string) int {
	var match *PathStatusCodesConfig

	for _, psc := range c.UnauthorizedPathStatusCodes {
		if psc.PathRegex.MatchString(path) {
			if match == nil || psc.PathLength > match.PathLength {
				match = psc
			}
		}
	}

	if match != nil {
		return match.StatusCode
	}

	return c.UnauthorizedStatusCode
}

func ValidatePrefix(prefix string) error {
	if prefix == "" {
		return nil
	}

	u, err := url.Parse(prefix)
	if err != nil {
		return err
	}

	isValid := u.Scheme == "" &&
		u.Host == "" &&
		u.User == nil &&
		u.RawQuery == "" &&
		u.Fragment == "" &&
		u.Path != "" &&
		prefix == u.Path

	if !isValid {
		return fmt.Errorf("must be a valid url path")
	}

	return nil
}
