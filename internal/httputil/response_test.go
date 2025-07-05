package httputil_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xabinapal/traefik-authentik-forward-plugin/internal/httputil"
)

func TestResponseMangler(t *testing.T) {
	t.Run("with mangle function", func(t *testing.T) {
		mangleCalled := false
		mangleFunc := func(rw http.ResponseWriter) {
			mangleCalled = true

			rw.Header().Set("X-Mangle-Test", "test_value")
		}

		responseWriter := httptest.NewRecorder()
		responseMangler := &httputil.ResponseMangler{
			ResponseWriter: responseWriter,
			MangleFunc:     mangleFunc,
		}

		responseMangler.WriteHeader(http.StatusOK)

		// check that the response code is the received one
		expectedCode := http.StatusOK
		if responseWriter.Code != expectedCode {
			t.Errorf("expected status code %d, got %d", expectedCode, responseWriter.Code)
		}

		// check that the mangle function was called
		if !mangleCalled {
			t.Fatalf("expected mangle function to be called")
		}

		// check that the response mangling is applied
		if responseWriter.Header().Get("X-Mangle-Test") != "test_value" {
			t.Errorf("expected X-Mangle-Test header to be set, got %s", responseWriter.Header().Get("X-Mangle-Test"))
		}
	})
}
