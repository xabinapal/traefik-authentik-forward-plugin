package config

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

var ErrConfigParse = errors.New("invalid config")

type Config struct {
	// The address of the Authentik server to forward requests to.
	Address string `json:"address"`

	// Part of the request path to keep, if required.
	KeepPrefix string `json:"keepPrefix,omitempty"`

	// The status code to return when the request is unauthorized.
	UnauthorizedStatusCode uint `json:"unauthorizedStatusCode,omitempty"`

	// Map of path regexes to customize their unauthorized status codes.
	UnauthorizedPathStatusCodes map[string]uint `json:"unauthorizedPathStatusCodes,omitempty"`
}

type ParsedConfig struct {
	Config
	UnauthorizedStatusCode      int
	UnauthorizedPathStatusCodes []*PathStatusCodesConfig
}

type PathStatusCodesConfig struct {
	PathRegex  *regexp.Regexp
	PathLength int
	StatusCode int
}

func (c *Config) Parse() (*ParsedConfig, error) {
	if c.Address == "" {
		return nil, errors.Join(
			ErrConfigParse,
			errors.New("address is required"),
		)
	}

	c.KeepPrefix = strings.Trim(c.KeepPrefix, "/")
	if err := ValidateKeepPrefix(c.KeepPrefix); err != nil {
		return nil, errors.Join(
			ErrConfigParse,
			fmt.Errorf("keepPrefix is not valid: %w", err),
		)
	} else if c.KeepPrefix != "" {
		c.KeepPrefix = "/" + c.KeepPrefix
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

	pc := &ParsedConfig{
		Config:                      *c,
		UnauthorizedStatusCode:      int(c.UnauthorizedStatusCode),
		UnauthorizedPathStatusCodes: pathStatusCodes,
	}

	return pc, nil
}

func (c *ParsedConfig) GetUnauthorizedStatusCode(path string) int {
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

func ValidateKeepPrefix(keepPrefix string) error {
	if keepPrefix == "" {
		return nil
	}

	u, err := url.Parse(keepPrefix)
	if err != nil {
		return err
	}

	isValid := u.Scheme == "" &&
		u.Host == "" &&
		u.User == nil &&
		u.RawQuery == "" &&
		u.Fragment == "" &&
		u.Path != "" &&
		keepPrefix == u.Path

	if !isValid {
		return fmt.Errorf("must be a valid url path")
	}

	return nil
}
