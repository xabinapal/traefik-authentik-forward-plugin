package config

import (
	"crypto/tls"
	"fmt"
)

const (
	DefaultTLSMinVersion         = 12
	DefaultTLSMaxVersion         = 13
	DefaultTLSInsecureSkipVerify = false
)

const (
	minValidTLSVersion = 10
	maxValidTLSVersion = 13
)

type RawTLSConfig struct {
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

type TLSConfig struct {
	RawTLSConfig
	MinVersion uint16
	MaxVersion uint16
}

func (c *RawTLSConfig) Parse() (*TLSConfig, error) {
	pc := &TLSConfig{
		RawTLSConfig: *c,
	}

	if c.MinVersion == 0 {
		c.MinVersion = DefaultTLSMinVersion
	} else if c.MinVersion < minValidTLSVersion || c.MinVersion > maxValidTLSVersion {
		return nil, fmt.Errorf("%w: tls.minVersion is not valid", ErrConfigParse)
	}

	pc.MinVersion = c.MinVersion - minValidTLSVersion + tls.VersionTLS10

	if c.MaxVersion == 0 {
		c.MaxVersion = DefaultTLSMaxVersion
	} else if c.MaxVersion < minValidTLSVersion || c.MaxVersion > maxValidTLSVersion {
		return nil, fmt.Errorf("%w: tls.maxVersion is not valid", ErrConfigParse)
	}

	pc.MaxVersion = c.MaxVersion - minValidTLSVersion + tls.VersionTLS10

	if pc.MinVersion > pc.MaxVersion {
		return nil, fmt.Errorf("%w: tls.minVersion cannot be higher than tls.maxVersion", ErrConfigParse)
	}

	return pc, nil
}
