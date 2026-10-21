package httputil_test

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"testing"
	"time"

	pluginhttp "github.com/xabinapal/traefik-authentik-forward-plugin/internal/httputil"
)

func TestResponseMangler(t *testing.T) {
	t.Run("with mangle function", func(t *testing.T) {
		mangleCalled := false
		mangleFunc := func(rw http.ResponseWriter) {
			mangleCalled = true

			rw.Header().Set("X-Mangle-Test", "test_value")
		}

		responseWriter := httptest.NewRecorder()
		responseMangler := &pluginhttp.ResponseMangler{
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

func TestResponseManglerStreamsEvents(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Content-Type", "text/event-stream")
		if _, err := fmt.Fprint(rw, "data: connected\n\n"); err != nil {
			t.Error(err)
			return
		}
		if err := http.NewResponseController(rw).Flush(); err != nil {
			t.Error(err)
			return
		}
		<-req.Context().Done()
	}))
	defer upstream.Close()
	upstreamURL, err := url.Parse(upstream.URL)
	if err != nil {
		t.Fatal(err)
	}
	proxy := httputil.NewSingleHostReverseProxy(upstreamURL)
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		proxy.ServeHTTP(&pluginhttp.ResponseMangler{
			ResponseWriter: rw,
			MangleFunc: func(rw http.ResponseWriter) {
				rw.Header().Set("X-Mangle-Test", "stream")
			},
		}, req)
	}))
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	response, err := server.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = response.Body.Close() }()
	if response.Header.Get("X-Mangle-Test") != "stream" {
		t.Fatal("expected mangled headers on the streamed response")
	}
	line, err := bufio.NewReader(response.Body).ReadString('\n')
	if err != nil {
		t.Fatal(err)
	}
	if line != "data: connected\n" {
		t.Fatalf("unexpected event: %q", line)
	}
}

func TestResponseManglerDoesNotFlushOrdinaryResponses(t *testing.T) {
	recorder := httptest.NewRecorder()
	writer := &pluginhttp.ResponseMangler{
		ResponseWriter: recorder,
		MangleFunc:     func(http.ResponseWriter) {},
	}
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	n, err := writer.Write([]byte("{}"))
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 || recorder.Body.String() != "{}" {
		t.Fatalf("unexpected response: %d bytes, body %q", n, recorder.Body.String())
	}
	if recorder.Flushed {
		t.Fatal("ordinary responses must not be flushed on each write")
	}
}
