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

func TestGetAuthentikStartPath(t *testing.T) {
	tests := []struct {
		request  string
		response string
	}{
		{
			request:  "https://example.com/protected",
			response: "https://example.com/outpost.goauthentik.io/start?rd=https%3A%2F%2Fexample.com%2Fprotected",
		},
		{
			request:  "https://example.com/protected?query=value",
			response: "https://example.com/outpost.goauthentik.io/start?rd=https%3A%2F%2Fexample.com%2Fprotected%3Fquery%3Dvalue",
		},
	}
	for _, tt := range tests {
		t.Run("with path "+tt.request, func(t *testing.T) {
			url, err := url.Parse(tt.request)
			if err != nil {
				t.Fatalf("failed to parse url: %v", err)
			}

			path := authentik.GetAuthentikStartPath(url)
			if path != tt.response {
				t.Errorf("expected path to be %s, got %s", tt.response, path)
			}
		})
	}
}
