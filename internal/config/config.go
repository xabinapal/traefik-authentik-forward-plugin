package config

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
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
}

func (c *Config) Parse() error {
	if c.Address == "" {
		return errors.Join(
			ErrConfigParse,
			errors.New("address is required"),
		)
	}

	c.KeepPrefix = strings.Trim(c.KeepPrefix, "/")
	if err := ValidateKeepPrefix(c.KeepPrefix); err != nil {
		return errors.Join(
			ErrConfigParse,
			fmt.Errorf("keepPrefix is not valid: %w", err),
		)
	} else if c.KeepPrefix != "" {
		c.KeepPrefix = "/" + c.KeepPrefix
	}

	if c.UnauthorizedStatusCode == 0 {
		c.UnauthorizedStatusCode = http.StatusUnauthorized
	}

	return nil
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
