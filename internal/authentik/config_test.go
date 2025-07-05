package authentik_test

import (
	"net/http"
	"regexp"
	"testing"

	"github.com/xabinapal/traefik-authentik-forward-plugin/internal/authentik"
)

func TestIsSkippedPath(t *testing.T) {
	t.Run("with no matching paths", func(t *testing.T) {
		cfg := authentik.Config{
			SkippedPaths: []*regexp.Regexp{regexp.MustCompile("^/test")},
		}

		isSkipped := cfg.IsSkippedPath("/admin")

		// check that the path is not skipped
		expectedIsSkipped := false
		if isSkipped != expectedIsSkipped {
			t.Errorf("expected isSkipped to be %t, got %t", expectedIsSkipped, isSkipped)
		}
	})

	t.Run("with matching path", func(t *testing.T) {
		cfg := authentik.Config{
			SkippedPaths: []*regexp.Regexp{regexp.MustCompile("^/test")},
		}

		isSkipped := cfg.IsSkippedPath("/test")

		// check that the path is skipped
		expectedIsSkipped := true
		if isSkipped != expectedIsSkipped {
			t.Errorf("expected isSkipped to be %t, got %t", expectedIsSkipped, isSkipped)
		}
	})
}

func TestGetUnauthorizedStatusCode(t *testing.T) {
	t.Run("with no matching paths", func(t *testing.T) {
		cfg := authentik.Config{
			Address:                "https://authentik.example.com",
			UnauthorizedStatusCode: http.StatusForbidden,
			RedirectStatusCode:     http.StatusFound,
			UnauthorizedPaths:      []*regexp.Regexp{regexp.MustCompile("^/admin")},
			RedirectPaths:          []*regexp.Regexp{regexp.MustCompile("^/login")},
		}

		statusCode := cfg.GetUnauthorizedStatusCode("/test")

		// check that the status code is the allowed one
		expectedStatusCode := http.StatusOK
		if statusCode != expectedStatusCode {
			t.Errorf("expected status %d, got %d", expectedStatusCode, statusCode)
		}
	})

	t.Run("with matching unauthorized path", func(t *testing.T) {
		cfg := authentik.Config{
			Address:                "https://authentik.example.com",
			UnauthorizedStatusCode: http.StatusForbidden,
			RedirectStatusCode:     http.StatusMovedPermanently,
			UnauthorizedPaths:      []*regexp.Regexp{regexp.MustCompile("^/admin"), regexp.MustCompile("^/test")},
			RedirectPaths:          []*regexp.Regexp{regexp.MustCompile("^/login")},
		}

		statusCode := cfg.GetUnauthorizedStatusCode("/test")

		// check that the status code is the unauthorized one
		expectedStatusCode := http.StatusForbidden
		if statusCode != expectedStatusCode {
			t.Errorf("expected status %d, got %d", expectedStatusCode, statusCode)
		}
	})

	t.Run("with matching redirect path", func(t *testing.T) {
		cfg := authentik.Config{
			Address:                "https://authentik.example.com",
			UnauthorizedStatusCode: http.StatusForbidden,
			RedirectStatusCode:     http.StatusMovedPermanently,
			UnauthorizedPaths:      []*regexp.Regexp{regexp.MustCompile("^/admin")},
			RedirectPaths:          []*regexp.Regexp{regexp.MustCompile("^/login"), regexp.MustCompile("^/test")},
		}

		statusCode := cfg.GetUnauthorizedStatusCode("/test")

		// check that the status code is the redirect one
		expectedStatusCode := http.StatusMovedPermanently
		if statusCode != expectedStatusCode {
			t.Errorf("expected status %d, got %d", expectedStatusCode, statusCode)
		}
	})

	t.Run("with longest matching for unauthorized path", func(t *testing.T) {
		cfg := authentik.Config{
			Address:                "https://authentik.example.com",
			UnauthorizedStatusCode: http.StatusForbidden,
			RedirectStatusCode:     http.StatusMovedPermanently,
			UnauthorizedPaths:      []*regexp.Regexp{regexp.MustCompile("^/test")},
			RedirectPaths:          []*regexp.Regexp{regexp.MustCompile("^/.*")},
		}

		statusCode := cfg.GetUnauthorizedStatusCode("/test")

		// check that the status code is the unauthorized one
		expectedStatusCode := http.StatusForbidden
		if statusCode != expectedStatusCode {
			t.Errorf("expected status %d, got %d", expectedStatusCode, statusCode)
		}
	})

	t.Run("with longest matching for redirect path", func(t *testing.T) {
		cfg := authentik.Config{
			Address:                "https://authentik.example.com",
			UnauthorizedStatusCode: http.StatusForbidden,
			RedirectStatusCode:     http.StatusMovedPermanently,
			UnauthorizedPaths:      []*regexp.Regexp{regexp.MustCompile("^/.*")},
			RedirectPaths:          []*regexp.Regexp{regexp.MustCompile("^/test")},
		}

		statusCode := cfg.GetUnauthorizedStatusCode("/test")

		// check that the status code is the redirect one
		expectedStatusCode := http.StatusMovedPermanently
		if statusCode != expectedStatusCode {
			t.Errorf("expected status %d, got %d", expectedStatusCode, statusCode)
		}
	})

	t.Run("with same length matching for both", func(t *testing.T) {
		cfg := authentik.Config{
			Address:                "https://authentik.example.com",
			UnauthorizedStatusCode: http.StatusForbidden,
			RedirectStatusCode:     http.StatusFound,
			UnauthorizedPaths:      []*regexp.Regexp{regexp.MustCompile(`^/test/?`)},
			RedirectPaths:          []*regexp.Regexp{regexp.MustCompile(`^/test/+`)},
		}

		statusCode := cfg.GetUnauthorizedStatusCode("/test/")

		// check that the status code is the unauthorized one
		expectedStatusCode := http.StatusForbidden
		if statusCode != expectedStatusCode {
			t.Errorf("expected status %d, got %d", expectedStatusCode, statusCode)
		}
	})
}
