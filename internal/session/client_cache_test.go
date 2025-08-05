package session_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/xabinapal/traefik-authentik-forward-plugin/internal/session"
)

func TestNewCacheClient(t *testing.T) {
	t.Run("with no duration", func(t *testing.T) {
		defer func() {
			// check that the constructor panics
			if r := recover(); r == nil {
				t.Fatal("expected panic")
			}
		}()

		session.NewCacheClient(context.Background(), 0)
	})

	t.Run("with duration", func(t *testing.T) {
		client := session.NewCacheClient(context.Background(), 10*time.Second)

		// check that the client is not nil
		if client == nil {
			t.Fatal("expected client to be not nil")
		}
	})
}

func TestCacheClient(t *testing.T) {
	t.Run("retrieve without store", func(t *testing.T) {
		client := session.NewCacheClient(context.Background(), 10*time.Second)

		session := client.Get([]*http.Cookie{
			{
				Name:  "test",
				Value: "test",
			},
		})

		// check that the session is nil
		if session != nil {
			t.Fatal("expected session to be nil")
		}
	})

	t.Run("retrieve after store", func(t *testing.T) {
		client := session.NewCacheClient(context.Background(), 10*time.Second)

		session := &session.Session{
			IsAuthenticated: true,
			Headers: http.Header{
				"X-Test": []string{"test"},
			},
			Cookies: []*http.Cookie{
				{
					Name:  "test",
					Value: "test",
				},
			},
		}
		client.Set([]*http.Cookie{
			{
				Name:  "test",
				Value: "test",
			},
		}, session)

		// check that the session is not nil
		session = client.Get([]*http.Cookie{
			{
				Name:  "test",
				Value: "test",
			},
		})
		if session == nil {
			t.Fatal("expected session to be not nil")
		}

		// check that the session has the expected values
		if !session.IsAuthenticated {
			t.Errorf("expected session to be authenticated")
		}

		if session.Headers.Get("X-Test") != "test" {
			t.Errorf("expected session to have original headers")
		}

		if len(session.Cookies) != 1 || session.Cookies[0].Name != "test" || session.Cookies[0].Value != "test" {
			t.Errorf("expected session to have original cookie")
		}
	})

	t.Run("retrieve after delete", func(t *testing.T) {
		client := session.NewCacheClient(context.Background(), 10*time.Second)

		session := &session.Session{
			IsAuthenticated: true,
			Headers:         http.Header{},
			Cookies: []*http.Cookie{
				{
					Name:  "test",
					Value: "test",
				},
			},
		}
		client.Set([]*http.Cookie{
			{
				Name:  "test",
				Value: "test",
			},
		}, session)
		client.Delete([]*http.Cookie{
			{
				Name:  "test",
				Value: "test",
			},
		})

		// check that the session is nil
		session = client.Get([]*http.Cookie{
			{
				Name:  "test",
				Value: "test",
			},
		})
		if session != nil {
			t.Errorf("expected session to be nil")
		}
	})

	t.Run("retrieve after expiration", func(t *testing.T) {
		client := session.NewCacheClient(context.Background(), 10*time.Millisecond)

		session := &session.Session{
			IsAuthenticated: true,
			Headers:         http.Header{},
			Cookies:         []*http.Cookie{},
		}
		client.Set([]*http.Cookie{
			{
				Name:  "test",
				Value: "test",
			},
		}, session)

		// wait for the session to expire
		time.Sleep(30 * time.Millisecond)

		// check that the session is nil
		session = client.Get([]*http.Cookie{
			{
				Name:  "test",
				Value: "test",
			},
		})
		if session != nil {
			t.Errorf("expected session to be nil")
		}
	})

	t.Run("retrieve after expiration cancel", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		client := session.NewCacheClient(ctx, 10*time.Millisecond)

		// cancel the context
		cancel()

		session := &session.Session{
			IsAuthenticated: true,
			Headers:         http.Header{},
			Cookies:         []*http.Cookie{},
		}
		client.Set([]*http.Cookie{
			{
				Name:  "test",
				Value: "test",
			},
		}, session)

		// wait for the session to expire
		time.Sleep(30 * time.Millisecond)

		// check that the session is not nil
		session = client.Get([]*http.Cookie{
			{
				Name:  "test",
				Value: "test",
			},
		})
		if session == nil {
			t.Errorf("expected session to be not nil")
		}
	})

	t.Run("delete before store", func(t *testing.T) {
		client := session.NewCacheClient(context.Background(), 10*time.Second)

		client.Delete([]*http.Cookie{
			{
				Name:  "test",
				Value: "test",
			},
		})

		// check that the session is nil
		session := client.Get([]*http.Cookie{
			{
				Name:  "test",
				Value: "test",
			},
		})
		if session != nil {
			t.Errorf("expected session to be nil")
		}
	})
}
