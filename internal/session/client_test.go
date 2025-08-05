package session_test

import (
	"context"
	"testing"
	"time"

	"github.com/xabinapal/traefik-authentik-forward-plugin/internal/session"
)

func TestNewClient(t *testing.T) {
	t.Run("with no duration", func(t *testing.T) {
		client := session.NewClient(context.Background(), 0)

		// check that the client is not nil
		if client == nil {
			t.Fatal("expected client to be not nil")
		}

		// check that the client is a standard client
		if _, ok := client.(*session.StandardClient); !ok {
			t.Fatal("expected client to be a standard client")
		}
	})

	t.Run("with duration", func(t *testing.T) {
		client := session.NewClient(context.Background(), 10*time.Second)

		// check that the client is not nil
		if client == nil {
			t.Fatal("expected client to be not nil")
		}

		// check that the client is a cache client
		if _, ok := client.(*session.CacheClient); !ok {
			t.Fatal("expected client to be a cache client")
		}
	})
}
