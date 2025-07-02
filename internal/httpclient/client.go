package httpclient

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/xabinapal/traefik-authentik-forward-plugin/internal/config"
)

func CreateHTTPClient(config *config.Config) (*http.Client, error) {
	transport, err := createHTTPTransport(config.TLS)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrClientCreate, err)
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   config.Timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // don't follow redirects
		},
	}

	return client, nil
}

func createHTTPTransport(tlsConfig *config.TLSConfig) (*http.Transport, error) {
	if tlsConfig == nil {
		return nil, errors.New("tls config is required")
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion:         tlsConfig.MinVersion,
			MaxVersion:         tlsConfig.MaxVersion,
			InsecureSkipVerify: tlsConfig.InsecureSkipVerify, //nolint:gosec
		},
	}

	// load ca certificate if provided
	if tlsConfig.CA != "" {
		caCert, err := os.ReadFile(tlsConfig.CA)
		if err != nil {
			return nil, fmt.Errorf("failed to load ca certificate: %w", err)
		}

		transport.TLSClientConfig.RootCAs = x509.NewCertPool()
		if !transport.TLSClientConfig.RootCAs.AppendCertsFromPEM(caCert) {
			return nil, errors.New("failed to parse ca certificate")
		}
	}

	// load client certificate and key if provided
	if tlsConfig.Cert != "" && tlsConfig.Key != "" {
		cert, err := tls.LoadX509KeyPair(tlsConfig.Cert, tlsConfig.Key)
		if err != nil {
			return nil, fmt.Errorf("failed to load client certificate: %w", err)
		}

		transport.TLSClientConfig.Certificates = []tls.Certificate{cert}
	} else if tlsConfig.Cert != "" || tlsConfig.Key != "" {
		return nil, errors.New("both cert and key must be provided for client certificate authentication")
	}

	return transport, nil
}
