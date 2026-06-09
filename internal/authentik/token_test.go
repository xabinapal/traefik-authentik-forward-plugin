package authentik_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/xabinapal/traefik-authentik-forward-plugin/internal/authentik"
)

const tokenUserResponse = `{
	"user": {
		"pk": 42,
		"username": "testuser",
		"email": "user@example.com",
		"groups": [
			{"name": "admins"},
			{"name": "users"}
		]
	}
}`

func tokenRequestMeta(token string) *authentik.RequestMeta {
	return &authentik.RequestMeta{
		URL: &url.URL{
			Scheme: "https",
			Host:   "example.com",
			Path:   "/protected",
		},
		Cookies: []*http.Cookie{},
		Token:   token,
	}
}

func TestCheckToken(t *testing.T) {
	t.Run("with valid token", func(t *testing.T) {
		akCalled := false
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			akCalled = true

			// check that the api path is used
			expectedPath := "/api/v3/core/users/me/"
			if r.URL.Path != expectedPath {
				t.Fatalf("expected path %s, got %s", expectedPath, r.URL.Path)
			}

			// check that the bearer token is forwarded
			expectedAuth := "Bearer test-token"
			if r.Header.Get("Authorization") != expectedAuth {
				t.Fatalf("expected authorization %s, got %s", expectedAuth, r.Header.Get("Authorization"))
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(tokenUserResponse))
		}))
		defer server.Close()

		config := &authentik.Config{Address: server.URL}
		client, _ := authentik.NewClient(context.Background(), server.Client(), config)

		resMeta, err := client.Check(tokenRequestMeta("test-token"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !akCalled {
			t.Fatalf("expected authentik server to be called")
		}

		// check that the request was authenticated
		if !resMeta.Session.IsAuthenticated {
			t.Error("expected request to be authenticated")
		}

		// check that the user data is mapped to authentik headers
		expectedHeaders := map[string]string{
			"X-Authentik-Uid":      "42",
			"X-Authentik-Username": "testuser",
			"X-Authentik-Email":    "user@example.com",
			"X-Authentik-Groups":   "admins|users",
		}

		for k, v := range expectedHeaders {
			if got := resMeta.Session.Headers.Get(k); got != v {
				t.Errorf("expected header %s=%s, got %s", k, v, got)
			}
		}
	})

	t.Run("with invalid token", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden)
		}))
		defer server.Close()

		config := &authentik.Config{Address: server.URL}
		client, _ := authentik.NewClient(context.Background(), server.Client(), config)

		resMeta, err := client.Check(tokenRequestMeta("bad-token"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// check that the request was not authenticated
		if resMeta.Session.IsAuthenticated {
			t.Error("expected request to be unauthenticated")
		}

		// check that no authentik headers are set
		if len(resMeta.Session.Headers) != 0 {
			t.Errorf("expected 0 headers, got %d", len(resMeta.Session.Headers))
		}
	})

	t.Run("with caching", func(t *testing.T) {
		akCalls := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			akCalls++

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(tokenUserResponse))
		}))
		defer server.Close()

		config := &authentik.Config{Address: server.URL, CacheDuration: time.Minute}
		client, _ := authentik.NewClient(context.Background(), server.Client(), config)

		// first request hits authentik
		resMeta, err := client.Check(tokenRequestMeta("test-token"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resMeta.Cached {
			t.Error("expected first request to not be cached")
		}

		// second request with the same token is served from cache
		resMeta, err = client.Check(tokenRequestMeta("test-token"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !resMeta.Cached {
			t.Error("expected second request to be cached")
		}

		// authentik must only be queried once
		if akCalls != 1 {
			t.Errorf("expected authentik to be called once, got %d", akCalls)
		}

		// a different token must not be served from the first token's cache
		resMeta, err = client.Check(tokenRequestMeta("other-token"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resMeta.Cached {
			t.Error("expected request with different token to not be cached")
		}
		if akCalls != 2 {
			t.Errorf("expected authentik to be called twice, got %d", akCalls)
		}
	})
}

func TestGetBearerToken(t *testing.T) {
	cases := []struct {
		name     string
		header   string
		expected string
	}{
		{"with bearer token", "Bearer abc123", "abc123"},
		{"with lowercase scheme", "bearer abc123", "abc123"},
		{"with surrounding spaces", "Bearer   abc123  ", "abc123"},
		{"without authorization", "", ""},
		{"with basic scheme", "Basic dXNlcjpwYXNz", ""},
		{"with empty token", "Bearer ", ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
			if tc.header != "" {
				req.Header.Set("Authorization", tc.header)
			}

			if got := authentik.GetBearerToken(req); got != tc.expected {
				t.Errorf("expected token %q, got %q", tc.expected, got)
			}
		})
	}
}
