package httpclient_test

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/xabinapal/traefik-authentik-forward-plugin/internal/httpclient"
)

//nolint:gochecknoglobals
var (
	TestCA = `
	-----BEGIN CERTIFICATE-----
	MIIBcTCCARugAwIBAgIUODWb/tp162q2N4Rwfea2H1KKoyQwDQYJKoZIhvcNAQEL
	BQAwDTELMAkGA1UEAwwCQ0EwHhcNMjUwNzA1MTI1ODI4WhcNMjUwODA0MTI1ODI4
	WjANMQswCQYDVQQDDAJDQTBcMA0GCSqGSIb3DQEBAQUAA0sAMEgCQQDLy6Z7emnC
	mxjgGfNCHuuPjWlP4juSRQzEvsqMmoEzhR83E/eJFmAg/rfgoO3CgGfbf5XJirLp
	kKzYM8kFbR3tAgMBAAGjUzBRMB0GA1UdDgQWBBQye8L6lpoBL16EnpMpCt1EDTdG
	jzAfBgNVHSMEGDAWgBQye8L6lpoBL16EnpMpCt1EDTdGjzAPBgNVHRMBAf8EBTAD
	AQH/MA0GCSqGSIb3DQEBCwUAA0EAAaCfAWXqpKYuHOSp/uS4Nyp1GY3QQ/7ifofb
	N+OChG9Y8pInpgVaxNlczADoSgAUv6jM61k+u9Cnlr+cwCRLqQ==
	-----END CERTIFICATE-----
	`

	TestCert = `
	-----BEGIN CERTIFICATE-----
	MIIBZDCCAQ6gAwIBAgIUFPQ7ZxxSkoMDtUH+FSKsohnjMe8wDQYJKoZIhvcNAQEL
	BQAwDTELMAkGA1UEAwwCQ0EwHhcNMjUwNzA1MTI1OTA0WhcNMjUwODA0MTI1OTA0
	WjARMQ8wDQYDVQQDDAZDbGllbnQwXDANBgkqhkiG9w0BAQEFAANLADBIAkEAude1
	BGgJRLYugeN0h1XOWD5W94+6YRABA5ascSth+BA4bS5EUOGQ3NuzrMyT6SJkZwyP
	JxyA60YeXKeOf+cZZwIDAQABo0IwQDAdBgNVHQ4EFgQUNr4UrC/8/Akyk9z7oHvi
	Z+6XTZ4wHwYDVR0jBBgwFoAUMnvC+paaAS9ehJ6TKQrdRA03Ro8wDQYJKoZIhvcN
	AQELBQADQQAHuBfou0DLFLNe1TDNEDfmturX3IvXftKnMg1zdjm/6UYkX3LQkI6K
	Kuh0h0g8IBhpolcVR8z+xIg5sCQYYeSK
	-----END CERTIFICATE-----
	`

	TestKey = `
	-----BEGIN PRIVATE KEY-----
	MIIBUwIBADANBgkqhkiG9w0BAQEFAASCAT0wggE5AgEAAkEAude1BGgJRLYugeN0
	h1XOWD5W94+6YRABA5ascSth+BA4bS5EUOGQ3NuzrMyT6SJkZwyPJxyA60YeXKeO
	f+cZZwIDAQABAkB1k4Jj6kpK3ZQo+zXDVcc5zx8Iezdop05s7cvlwZO287/3kW54
	cDT4NNkg9gcRsgtGFrAvjefEwB3PVTpG4bHBAiEA4wjbEB3dtimht+zNKGmJywtO
	cgyu9S1aJVfsO/f+QccCIQDRjXxd7m2XPQwzJVojDvjhIE9qRdaHlSXwSqXdBF5r
	YQIgcC9wEAayB9GKl9friIyeCjcMDE84JO7EHK/Vi8x/VwECIHMRQTCHI1B/6joP
	ka5co1djmZgen022Ha4UH338zyghAiB77ck7rJIoAeS3K+2KL/aamXaQ/JeYCvQY
	wS0rgmjmEQ==
	-----END PRIVATE KEY-----
	`
)

func fakeReader(path string) ([]byte, error) {
	var data string

	switch path {
	case "testdata/ca.crt":
		data = TestCA
	case "testdata/cert.crt":
		data = TestCert
	case "testdata/key.key":
		data = TestKey
	}

	data = strings.ReplaceAll(data, "\t", "")
	data = strings.TrimPrefix(data, "\n")
	return []byte(data), nil
}

func TestNew(t *testing.T) {
	t.Run("with default config", func(t *testing.T) {
		cfg := &httpclient.Config{}
		client, err := httpclient.New(cfg)

		// check that there is no error
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

	t.Run("with timeout settings", func(t *testing.T) {
		cfg := &httpclient.Config{
			Timeout: 10 * time.Second,
		}
		client, err := httpclient.New(cfg)

		// check that there is no error
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// check that the client is not nil
		if client == nil {
			t.Fatalf("expected client to be not nil")
		}

		// check that the client has the correct timeout
		expectedTimeout := 10 * time.Second
		if client.Timeout != expectedTimeout {
			t.Fatalf("expected timeout %v, got %v", expectedTimeout, client.Timeout)
		}
	})

	t.Run("with tls settings", func(t *testing.T) {
		cfg := &httpclient.Config{
			TLS: httpclient.TLSConfig{
				CA:                 "testdata/ca.crt",
				Cert:               "testdata/cert.crt",
				Key:                "testdata/key.key",
				MinVersion:         tls.VersionTLS10,
				MaxVersion:         tls.VersionTLS11,
				InsecureSkipVerify: true,
			},
		}
		client, err := httpclient.NewWithReader(cfg, fakeReader)

		// check that there is no error
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// check that the client is not nil
		if client == nil {
			t.Fatalf("expected client to be not nil")
		}

		transport, ok := client.Transport.(*http.Transport)
		if !ok {
			t.Fatalf("expected transport to be an *http.Transport")
		}

		// check that the client has the correct tls config
		ca, _ := fakeReader("testdata/ca.crt")
		expectedRootCAs := x509.NewCertPool()
		expectedRootCAs.AppendCertsFromPEM(ca)
		if !transport.TLSClientConfig.RootCAs.Equal(expectedRootCAs) {
			t.Fatalf("expected ca pool to contain the test ca")
		}

		cert, _ := fakeReader("testdata/cert.crt")
		key, _ := fakeReader("testdata/key.key")
		kp, _ := tls.X509KeyPair(cert, key)
		expectedCertificates := []tls.Certificate{kp}
		if !reflect.DeepEqual(transport.TLSClientConfig.Certificates, expectedCertificates) {
			t.Fatalf("expected certificates to contain the test cert")
		}

		var expectedTLSMinVersion uint16 = tls.VersionTLS10
		if transport.TLSClientConfig.MinVersion != expectedTLSMinVersion {
			t.Fatalf("expected min version %v, got %v", expectedTLSMinVersion, transport.TLSClientConfig.MinVersion)
		}

		var expectedTLSMaxVersion uint16 = tls.VersionTLS11
		if transport.TLSClientConfig.MaxVersion != expectedTLSMaxVersion {
			t.Fatalf("expected max version %v, got %v", expectedTLSMaxVersion, transport.TLSClientConfig.MaxVersion)
		}

		expectedTLSInsecureSkipVerify := true
		if transport.TLSClientConfig.InsecureSkipVerify != expectedTLSInsecureSkipVerify {
			t.Fatalf("expected insecure skip verify to be %v, got %v", expectedTLSInsecureSkipVerify, transport.TLSClientConfig.InsecureSkipVerify)
		}
	})
}
