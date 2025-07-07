package authentik_test

import (
	"net/url"
	"testing"

	"github.com/xabinapal/traefik-authentik-forward-plugin/internal/authentik"
)

func TestIsAuthentikPathAllowed(t *testing.T) {
	var tests []string

	tests = []string{
		"/outpost.goauthentik.io/start",
		"/outpost.goauthentik.io/sign_out",
		"/outpost.goauthentik.io/callback",
	}
	for _, tt := range tests {
		t.Run("with allowed path "+tt, func(t *testing.T) {
			allowed := authentik.IsAuthentikPathAllowed(tt)
			if !allowed {
				t.Errorf("expected path to be allowed")
			}
		})
	}

	tests = []string{
		"/outpost.goauthentik.io",
		"/outpost.goauthentik.io/auth/nginx",
		"/outpost.goauthentik.io/auth/traefik",
		"/outpost.goauthentik.io/auth/caddy",
		"/outpost.goauthentik.io/auth/envoy",
	}
	for _, tt := range tests {
		t.Run("with restricted path "+tt, func(t *testing.T) {
			allowed := authentik.IsAuthentikPathAllowed(tt)
			if allowed {
				t.Errorf("expected path to be restricted")
			}
		})
	}
}

func TestGetStartURL(t *testing.T) {
	tests := []string{
		"/protected",
		"/protected?query=value",
	}
	for _, tt := range tests {
		t.Run("with path "+tt, func(t *testing.T) {
			origUrl, err := url.Parse("https://example.com" + tt)
			if err != nil {
				t.Fatalf("failed to parse url: %v", err)
			}

			// check that the url is correct
			encodedUrl := url.QueryEscape(origUrl.String())
			expectedUrl := "https://example.com/outpost.goauthentik.io/start?rd=" + encodedUrl
			startUrl := authentik.GetStartURL(origUrl)
			if startUrl != expectedUrl {
				t.Errorf("expected path to be %s, got %s", expectedUrl, startUrl)
			}
		})
	}
}
