package httpclient_test

import (
	"crypto/tls"
	"net/http"
	"testing"
	"time"

	"github.com/xabinapal/traefik-authentik-forward-plugin/internal/config"
	"github.com/xabinapal/traefik-authentik-forward-plugin/internal/httpclient"
)

func TestCreateHTTPClient(t *testing.T) {
	t.Run("with default config", func(t *testing.T) {
		config := &config.RawConfig{
			Address: "https://authentik.example.com",
		}

		pc, _ := config.Parse()

		client, err := httpclient.CreateHTTPClient(pc)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// check that the client is not nil
		if client == nil {
			t.Fatalf("expected client to be not nil")
		}

		// check that the client has the correct timeout
		expectedTimeout := 0 * time.Second
		if client.Timeout != expectedTimeout {
			t.Fatalf("expected timeout %v, got %v", expectedTimeout, client.Timeout)
		}

		transport, ok := client.Transport.(*http.Transport)
		if !ok {
			t.Fatalf("expected transport to be an *http.Transport")
		}

		// check that the client has the correct tls config
		if transport.TLSClientConfig.RootCAs != nil {
			t.Fatalf("expected ca pool to be nil, got %v", transport.TLSClientConfig.RootCAs)
		}

		// check certificates is nil
		if transport.TLSClientConfig.Certificates != nil {
			t.Fatalf("expected certificates to be nil, got %v", transport.TLSClientConfig.Certificates)
		}

		var expectedTLSMinVersion uint16 = tls.VersionTLS12
		if transport.TLSClientConfig.MinVersion != expectedTLSMinVersion {
			t.Fatalf("expected min version %v, got %v", expectedTLSMinVersion, transport.TLSClientConfig.MinVersion)
		}

		var expectedTLSMaxVersion uint16 = tls.VersionTLS13
		if transport.TLSClientConfig.MaxVersion != expectedTLSMaxVersion {
			t.Fatalf("expected max version %v, got %v", expectedTLSMaxVersion, transport.TLSClientConfig.MaxVersion)
		}

		expectedTLSInsecureSkipVerify := false
		if transport.TLSClientConfig.InsecureSkipVerify != expectedTLSInsecureSkipVerify {
			t.Fatalf("expected insecure skip verify to be %v, got %v", expectedTLSInsecureSkipVerify, transport.TLSClientConfig.InsecureSkipVerify)
		}
	})
}
