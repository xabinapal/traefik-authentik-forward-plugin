package httputil_test

import (
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"github.com/xabinapal/traefik-authentik-forward-plugin/internal/httputil"
)

func TestResponseModifier_WriteHeader(t *testing.T) {
	t.Run("cookie injection", func(t *testing.T) {
		// create a response recorder
		rw := httptest.NewRecorder()

		// add existing cookies to the response writer
		cookies := []string{
			"first_cookie=abc123; Path=/",
			"test_1=old_test_1; Path=/",
			"test_2=old_test_2; Path=/",
			"other_cookie=value; Path=/",
		}

		for _, cookie := range cookies {
			rw.Header().Add("Set-Cookie", cookie)
		}

		// create response modifier
		rcm := &httputil.ResponseCookieModifier{
			ResponseWriter: rw,
			CookiesPrefix:  "test_",
			Cookies: []*http.Cookie{
				{Name: "test_1", Value: "new_test_1", Path: "/"},
				{Name: "test_3", Value: "new_test_3", Path: "/"},
			},
		}

		// call the function
		rcm.WriteHeader(318)

		// check status code
		if rw.Code != 318 {
			t.Errorf("expected status code 318, got %d", rw.Code)
		}

		// check cookies
		expectedCookies := []string{
			"first_cookie=abc123; Path=/",
			"other_cookie=value; Path=/",
			"test_1=new_test_1; Path=/",
			"test_3=new_test_3; Path=/",
		}

		actualCookies := rw.Header().Values("Set-Cookie")

		if len(actualCookies) != len(expectedCookies) {
			t.Errorf("expected %d cookies, got %d", len(expectedCookies), len(actualCookies))
			t.Errorf("Expected: %v", expectedCookies)
			t.Errorf("Actual: %v", actualCookies)
			return
		}

		// check each expected cookie is present
		for _, expectedCookie := range expectedCookies {
			found := slices.Contains(actualCookies, expectedCookie)
			if !found {
				t.Errorf("expected cookie %q not found in actual cookies", expectedCookie)
			}
		}

		// check that no unexpected cookies are present
		for _, cookie := range actualCookies {
			if !slices.Contains(expectedCookies, cookie) {
				t.Errorf("unexpected cookie %q found in actual cookies", cookie)
			}
		}
	})
}
