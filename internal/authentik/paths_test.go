package authentik_test

import (
	"testing"

	"github.com/xabinapal/traefik-authentik-forward-plugin/internal/authentik"
)

func TestIsPathAllowed(t *testing.T) {
	t.Run("allowed paths", func(t *testing.T) {
		allowed := authentik.IsPathAllowed("/outpost.goauthentik.go/auth/start")
		if !allowed {
			t.Errorf("expected path to be allowed")
		}
	})

	t.Run("restricted paths", func(t *testing.T) {
		allowed := authentik.IsPathAllowed("/outpost.goauthentik.go/auth/nginx")
		if allowed {
			t.Errorf("expected path to be restricted")
		}
	})

	t.Run("default paths", func(t *testing.T) {
		allowed := authentik.IsPathAllowed("/outpost.goauthentik.go/callback")
		if !allowed {
			t.Errorf("expected path to be allowed")
		}
	})
}
