package authentik_test

import (
	"testing"

	"github.com/xabinapal/traefik-authentik-forward-plugin/internal/authentik"
)

func TestIsPathAllowedDownstream_Allowed(t *testing.T) {
	tests := []string{
		"/outpost.goauthentik.io/start",
		"/outpost.goauthentik.io/sign_out",
		"/outpost.goauthentik.io/callback",
	}
	for _, tt := range tests {
		t.Run("allowed path "+tt, func(t *testing.T) {
			allowed := authentik.IsPathAllowedDownstream(tt)
			if !allowed {
				t.Errorf("expected path to be allowed")
			}
		})
	}
}

func TestIsPathAllowedDownstream_Restricted(t *testing.T) {
	tests := []string{
		"/outpost.goauthentik.io",
		"/outpost.goauthentik.io/auth/nginx",
		"/outpost.goauthentik.io/auth/traefik",
		"/outpost.goauthentik.io/auth/caddy",
		"/outpost.goauthentik.io/auth/envoy",
	}
	for _, tt := range tests {
		t.Run("restricted path "+tt, func(t *testing.T) {
			allowed := authentik.IsPathAllowedDownstream(tt)
			if allowed {
				t.Errorf("expected path to be restricted")
			}
		})
	}
}
