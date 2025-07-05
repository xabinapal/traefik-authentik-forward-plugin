package httpclient

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net/http"
	"os"
)

func New(cfg *Config) (*http.Client, error) {
	return NewWithReader(cfg, os.ReadFile)
}

func NewWithReader(cfg *Config, reader func(string) ([]byte, error)) (*http.Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("%w: config is required", ErrClientCreate)
	}

	transport, err := createHTTPTransport(cfg, reader)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrClientCreate, err)
	}

	client := &http.Client{
		Timeout:   cfg.Timeout,
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// don't follow redirects
			return http.ErrUseLastResponse
		},
	}

	return client, nil
}

func createHTTPTransport(cfg *Config, reader func(string) ([]byte, error)) (*http.Transport, error) {
	transport := &http.Transport{}

	tlsConfig, err := createTLSConfig(cfg, reader)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrClientCreate, err)
	}

	transport.TLSClientConfig = tlsConfig

	return transport, nil
}

func createTLSConfig(cfg *Config, reader func(string) ([]byte, error)) (*tls.Config, error) {
	tlsConfig := &tls.Config{} //nolint:gosec

	err := setCertificates(cfg, tlsConfig, reader)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrClientCreate, err)
	}

	err = setVersions(cfg, tlsConfig)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrClientCreate, err)
	}

	// set insecure skip verify
	tlsConfig.InsecureSkipVerify = cfg.TLS.InsecureSkipVerify

	return tlsConfig, nil
}

func setCertificates(cfg *Config, tlsConfig *tls.Config, reader func(string) ([]byte, error)) error {
	// load ca certificate if provided
	if cfg.TLS.CA != "" {
		caData, err := reader(cfg.TLS.CA)
		if err != nil {
			return fmt.Errorf("failed to load ca certificate: %w", err)
		}

		ca := x509.NewCertPool()
		if ok := ca.AppendCertsFromPEM(caData); !ok {
			return errors.New("failed to parse ca certificate")
		}

		tlsConfig.RootCAs = ca
	}

	// load client key pair if provided
	if cfg.TLS.Cert != "" && cfg.TLS.Key != "" {
		certData, err := reader(cfg.TLS.Cert)
		if err != nil {
			return fmt.Errorf("failed to load client certificate: %w", err)
		}

		keyData, err := reader(cfg.TLS.Key)
		if err != nil {
			return fmt.Errorf("failed to load client key: %w", err)
		}

		cert, err := tls.X509KeyPair(certData, keyData)
		if err != nil {
			return fmt.Errorf("failed to parse client key pair: %w", err)
		}

		tlsConfig.Certificates = []tls.Certificate{cert}
	} else if cfg.TLS.Cert != "" || cfg.TLS.Key != "" {
		return errors.New("both cert and key must be provided")
	}

	return nil
}

func setVersions(cfg *Config, tlsConfig *tls.Config) error {
	// set minimum tls version
	switch {
	case cfg.TLS.MinVersion == 0:
		tlsConfig.MinVersion = tls.VersionTLS12
	case cfg.TLS.MaxVersion >= tls.VersionTLS10 && cfg.TLS.MaxVersion <= tls.VersionTLS13:
		tlsConfig.MinVersion = cfg.TLS.MinVersion
	default:
		return errors.New("tls.minVersion is not valid")
	}

	// set maximum tls version
	switch {
	case cfg.TLS.MaxVersion == 0:
		tlsConfig.MaxVersion = tls.VersionTLS13
	case cfg.TLS.MaxVersion >= tls.VersionTLS10 && cfg.TLS.MaxVersion <= tls.VersionTLS13:
		tlsConfig.MaxVersion = cfg.TLS.MaxVersion
	default:
		return errors.New("tls.maxVersion is not valid")
	}

	if tlsConfig.MinVersion > tlsConfig.MaxVersion {
		return errors.New("tls.minVersion is greater than tls.maxVersion")
	}

	return nil
}
