package authentik_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xabinapal/traefik-authentik-forward-plugin/internal/authentik"
	"github.com/xabinapal/traefik-authentik-forward-plugin/internal/httputil"
)

func contains(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

func TestGetHeaders(t *testing.T) {
	t.Run("with response", func(t *testing.T) {
		res := &http.Response{
			Header: http.Header{
				"Content-Type":      []string{"application/json"},
				"Cache-Control":     []string{"no-cache"},
				"X-Custom-Header":   []string{"value"},
				"X-Authentik":       []string{"user123"},
				"X-AuthentikUser":   []string{"user123"},
				"X-Authentik-User":  []string{"user456"},
				"X-Authentik-Email": []string{"user@example.com"},
			},
		}

		result := authentik.GetHeaders(res)

		// check that the expected headers are present
		expectedHeaders := map[string]string{
			"X-Authentik-User":  "user456",
			"X-Authentik-Email": "user@example.com",
		}

		if len(result) != len(expectedHeaders) {
			t.Errorf("expected %d headers, got %d", len(expectedHeaders), len(result))
		}

		for k, v := range result {
			actual, ok := expectedHeaders[k]
			if !ok {
				t.Fatalf("expected header %s not found in result", k)
			}

			if len(v) != 1 {
				t.Errorf("expected 1 value for header %s, got %d", k, len(v))
			}

			if v[0] != actual {
				t.Errorf("expected value %s for header %s, got %s", actual, k, v[0])
			}
		}
	})
}

func TestGetCookies(t *testing.T) {
	t.Run("with request", func(t *testing.T) {
		req := &http.Request{
			Header: http.Header{
				"Cookie": []string{
					"session_id=abc123",
					"csrf_token=xyz789",
					"authentik_proxy=value",
					"authentik_proxy_=value",
					"authentik_proxy_session1=session1",
					"authentik_proxy_session2=session2"},
			},
		}

		result := authentik.GetCookies(req)

		// check that the expected cookies are present
		expectedCookies := map[string]string{
			"authentik_proxy_session1": "session1",
			"authentik_proxy_session2": "session2",
		}

		if len(result) != len(expectedCookies) {
			t.Errorf("expected %d cookies, got %d", len(expectedCookies), len(result))
		}

		for _, v := range result {
			actual, ok := expectedCookies[v.Name]
			if !ok {
				t.Fatalf("expected cookie %s not found in result", v.Name)
			}

			if v.Value != actual {
				t.Errorf("expected value %s for cookie %s, got %s", actual, v.Name, v.Value)
			}
		}
	})

	t.Run("with response", func(t *testing.T) {
		res := &http.Response{
			Header: http.Header{
				"Set-Cookie": []string{
					"session_id=abc123",
					"csrf_token=xyz789",
					"authentik_proxy=value",
					"authentik_proxy_=value",
					"authentik_proxy_session1=session1",
					"authentik_proxy_session2=session2"},
			},
		}

		result := authentik.GetCookies(res)

		// check that the expected cookies are present
		expectedCookies := map[string]string{
			"authentik_proxy_session1": "session1",
			"authentik_proxy_session2": "session2",
		}

		if len(result) != len(expectedCookies) {
			t.Errorf("expected %d cookies, got %d", len(expectedCookies), len(result))
		}

		for _, v := range result {
			actual, ok := expectedCookies[v.Name]
			if !ok {
				t.Fatalf("expected cookie %s not found in result", v.Name)
			}

			if v.Value != actual {
				t.Errorf("expected value %s for cookie %s, got %s", actual, v.Name, v.Value)
			}
		}
	})
}

func TestRequestMangle(t *testing.T) {
	t.Run("with downstream headers", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "http://authentik.example.com", nil)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}

		req.Header.Set("X-Authentik-User", "user123")
		req.Header.Set("X-Authentik-Email", "user@example.com")
		req.Header.Set("X-Other", "value")

		authentik.RequestMangle(req)

		// check that the authentik headers are removed
		if _, ok := req.Header["X-Authentik-User"]; ok {
			t.Errorf("expected authentik header to be removed")
		}

		if _, ok := req.Header["X-Authentik-Email"]; ok {
			t.Errorf("expected authentik header to be removed")
		}

		// check that the other headers are not removed
		if _, ok := req.Header["X-Other"]; !ok {
			t.Errorf("expected other header to be preserved")
		}
	})

	t.Run("with downstream cookies", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "http://authentik.example.com", nil)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}

		req.AddCookie(&http.Cookie{
			Name:  "authentik_proxy_session1",
			Value: "session1",
		})

		req.AddCookie(&http.Cookie{
			Name:  "authentik_proxy_session2",
			Value: "session2",
		})

		req.AddCookie(&http.Cookie{
			Name:  "other",
			Value: "value",
		})

		authentik.RequestMangle(req)

		// check that the authentik cookies are removed
		if _, err := req.Cookie("authentik_proxy_session1"); err == nil {
			t.Errorf("expected authentik cookie to be removed")
		}

		// check that the authentik cookies are removed
		if _, err := req.Cookie("authentik_proxy_session2"); err == nil {
			t.Errorf("expected authentik cookie to be removed")
		}

		// check that the other cookies are not removed
		if _, err := req.Cookie("other"); err != nil {
			t.Errorf("expected other cookie to be preserved")
		}
	})
}

func TestGetResponseMangler(t *testing.T) {
	t.Run("with upstream headers", func(t *testing.T) {
		mangler := authentik.GetResponseMangler(nil)

		rw := httptest.NewRecorder()
		rw.Header().Set("X-Authentik-User", "user123")
		rw.Header().Set("X-Authentik-Email", "user@example.com")
		rw.Header().Set("X-Other", "value")

		mangler(rw)

		// check that the authentik headers are removed
		if _, ok := rw.Header()["X-Authentik-User"]; ok {
			t.Errorf("expected authentik header to be removed")
		}

		if _, ok := rw.Header()["X-Authentik-Email"]; ok {
			t.Errorf("expected authentik header to be removed")
		}

		// check that the other headers are not removed
		if _, ok := rw.Header()["X-Other"]; !ok {
			t.Errorf("expected other header to be preserved")
		}
	})

	t.Run("with upstream cookies", func(t *testing.T) {
		mangler := authentik.GetResponseMangler(nil)

		rw := httptest.NewRecorder()
		rw.Header().Add("Set-Cookie", "authentik_proxy_session1=session1")
		rw.Header().Add("Set-Cookie", "authentik_proxy_session2=session2")
		rw.Header().Add("Set-Cookie", "other=value")

		mangler(rw)

		cookies := rw.Header().Values("Set-Cookie")
		cookieNames := make([]string, 0, len(cookies))
		for _, cookie := range cookies {
			cookieNames = append(cookieNames, httputil.ParseCookieName(cookie))
		}

		// check that the authentik cookies are removed
		if ok := contains(cookieNames, "authentik_proxy_session1"); ok {
			t.Errorf("expected upstream cookie authentik_proxy_session1 to be removed")
		}

		if ok := contains(cookieNames, "authentik_proxy_session2"); ok {
			t.Errorf("expected upstream cookie authentik_proxy_session2 to be removed")
		}

		// check that the other cookies are not removed
		if ok := contains(cookieNames, "other"); !ok {
			t.Errorf("expected other cookie to be preserved")
		}
	})

	t.Run("with authentik cookies", func(t *testing.T) {
		mangler := authentik.GetResponseMangler([]*http.Cookie{
			{Name: "authentik_proxy_session1", Value: "session1"},
			{Name: "authentik_proxy_session2", Value: "session2"},
		})

		rw := httptest.NewRecorder()
		rw.Header().Add("Set-Cookie", "authentik_proxy_session3=session3")
		rw.Header().Add("Set-Cookie", "other=value5")

		mangler(rw)

		cookies := rw.Header().Values("Set-Cookie")
		cookieNames := make([]string, 0, len(cookies))
		for _, cookie := range cookies {
			cookieNames = append(cookieNames, httputil.ParseCookieName(cookie))
		}

		// check that the upstream authentik cookies are removed
		if ok := contains(cookieNames, "authentik_proxy_session3"); ok {
			t.Errorf("expected upstream cookie authentik_proxy_session3 to be removed")
		}
		if ok := contains(cookieNames, "authentik_proxy_session4"); ok {
			t.Errorf("expected upstream cookie authentik_proxy_session4 to be removed")
		}

		// check that the upstream other cookies are preserved
		if ok := contains(cookieNames, "other"); !ok {
			t.Errorf("expected upstream cookie other to be preserved")
		}

		// check that the authentik cookies are preserved
		if ok := contains(cookieNames, "authentik_proxy_session1"); !ok {
			t.Errorf("expected authentik cookie authentik_proxy_session1 to be preserved")
		}
		if ok := contains(cookieNames, "authentik_proxy_session2"); !ok {
			t.Errorf("expected authentik cookie authentik_proxy_session2 to be preserved")
		}
	})
}
