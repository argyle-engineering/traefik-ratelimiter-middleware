//nolint:stylecheck,revive
package traefik_ratelimiter_middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

type TestServeHTTP struct {
	haveResponse   int
	rateLimiterURL string
	requestURL     string
	haveDryRun     bool
	expectCall     bool
	expectResponse int
	expectBody     string
	expectHost     string
	expectPath     string
}

func TestBypassIfRatelimiterConfigIsMalformed(t *testing.T) {
	RunTestServeHTTP(t, TestServeHTTP{
		rateLimiterURL: "https://api.xxx.com/cgi-bin/%%32%65%%32%65/%%32%65%%32%65/%%32%65%%",
		expectCall:     true,
		expectResponse: http.StatusOK,
		expectBody:     "OK",
	})
}

func TestFailedCallingRateLimiter(t *testing.T) {
	RunTestServeHTTP(t, TestServeHTTP{
		rateLimiterURL: "127.0.0.1",
		haveResponse:   0, // cause failure
		expectCall:     true,
		expectResponse: http.StatusOK,
		expectBody:     "OK",
	})
}

func TestDeniedByRateLimiter(t *testing.T) {
	RunTestServeHTTP(t, TestServeHTTP{
		haveResponse:   http.StatusTooManyRequests,
		expectCall:     false,
		expectResponse: http.StatusTooManyRequests,
		expectBody:     "Too many requests\n",
	})
}

func TestDeniedByRateLimiterDryRun(t *testing.T) {
	RunTestServeHTTP(t, TestServeHTTP{
		haveResponse:   http.StatusTooManyRequests,
		haveDryRun:     true,
		expectCall:     true,
		expectResponse: http.StatusOK,
		expectBody:     "OK",
	})
}

func TestRateLimiterReturnsOk(t *testing.T) {
	RunTestServeHTTP(t, TestServeHTTP{
		haveResponse:   http.StatusOK,
		expectCall:     true,
		expectResponse: http.StatusOK,
		expectBody:     "OK",
	})
}

func TestPathIsSent(t *testing.T) {
	RunTestServeHTTP(t, TestServeHTTP{
		requestURL:     "https://test.test/v1/resource",
		haveResponse:   http.StatusOK,
		expectCall:     true,
		expectResponse: http.StatusOK,
		expectBody:     "OK",
		expectPath:     "/v1/resource",
	})
}

func TestDefaultPath(t *testing.T) {
	RunTestServeHTTP(t, TestServeHTTP{
		requestURL:     "https://test.test",
		haveResponse:   http.StatusOK,
		expectCall:     true,
		expectResponse: http.StatusOK,
		expectBody:     "OK",
		expectPath:     "/",
	})
}

func TestHostIsSent(t *testing.T) {
	RunTestServeHTTP(t, TestServeHTTP{
		requestURL:     "https://test-host.test",
		haveResponse:   http.StatusOK,
		expectCall:     true,
		expectResponse: http.StatusOK,
		expectBody:     "OK",
		expectHost:     "test-host.test",
	})
}

func RunTestServeHTTP(t *testing.T, tt TestServeHTTP) {
	t.Helper()
	ratelimiterServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		user, pass, ok := req.BasicAuth()
		if !ok || user != "user" || pass != "pass" {
			t.Error("unexpected user/password")
		}
		if tt.expectPath != "" && tt.expectPath != req.URL.Path {
			t.Errorf("got path `%s` but expected `%s`", req.URL.Path, tt.expectPath)
		}
		if tt.expectHost != "" && tt.expectHost != req.Host {
			t.Errorf("got host `%s` but expected `%s`", req.Host, tt.expectHost)
		}
		rw.WriteHeader(tt.haveResponse)
		_, _ = rw.Write([]byte(http.StatusText(tt.haveResponse)))
	}))
	defer ratelimiterServer.Close()

	url := ratelimiterServer.URL
	if url == "" {
		url = tt.rateLimiterURL
	}
	called := false
	rl, err := New(context.TODO(), http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		called = true
		rw.WriteHeader(http.StatusOK)
		_, _ = rw.Write([]byte("OK"))
	}), &Config{
		URL:    url,
		DryRun: tt.haveDryRun,
	}, "")
	if err != nil {
		t.Error(err)
	}

	if tt.requestURL == "" {
		tt.requestURL = "https://some.url.com"
	}

	req, err := http.NewRequest(http.MethodGet, tt.requestURL, nil)
	if err != nil {
		t.Errorf("couldn't create request: %v", err)
	}
	req.SetBasicAuth("user", "pass")

	rw := httptest.NewRecorder()
	rl.ServeHTTP(rw, req)

	if tt.expectCall != called {
		t.Errorf("expectCall(%t)!=called(%t)", tt.expectCall, called)
	}

	if tt.expectResponse != rw.Code {
		t.Errorf("tt.expectResponse(%d)!=rw.Code(%d)", tt.expectResponse, rw.Code)
	}

	if tt.expectBody != rw.Body.String() {
		t.Errorf("tt.expectBody(%s)!=rw.Body(%s)", tt.expectBody, rw.Body.String())
	}
}
