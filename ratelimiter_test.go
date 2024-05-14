package ratelimiter

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServeHTTP(t *testing.T) {
	tests := []struct {
		name           string
		haveResponse   int
		haveURL        string
		haveDryRun     bool
		expectCall     bool
		expectResponse int
		expectBody     string
	}{
		{
			name:           "bypass if ratelimiter config is malformed",
			haveURL:        "https://api.xxx.com/cgi-bin/%%32%65%%32%65/%%32%65%%32%65/%%32%65%%",
			expectCall:     true,
			expectResponse: http.StatusOK,
			expectBody:     "OK",
		},
		{
			name:           "failed calling ratelimiter",
			haveURL:        "127.0.0.1",
			expectCall:     true,
			expectResponse: http.StatusOK,
			expectBody:     "OK",
		},
		{
			name:           "denied by ratelimiter",
			haveResponse:   http.StatusTooManyRequests,
			expectCall:     false,
			expectResponse: http.StatusTooManyRequests,
			expectBody:     "Too many requests\n",
		},
		{
			name:           "denied by ratelimiter; dry run",
			haveResponse:   http.StatusTooManyRequests,
			haveDryRun:     true,
			expectCall:     true,
			expectResponse: http.StatusOK,
			expectBody:     "OK",
		},
		{
			name:           "ratelimiter returns OK",
			haveResponse:   http.StatusOK,
			expectCall:     true,
			expectResponse: http.StatusOK,
			expectBody:     "OK",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ratelimiterServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				user, pass, ok := req.BasicAuth()
				if !ok || user != "user" || pass != "pass" {
					t.FailNow()
				}
				rw.WriteHeader(tt.haveResponse)
				_, _ = rw.Write([]byte(http.StatusText(tt.haveResponse)))
			}))
			defer ratelimiterServer.Close()

			url := ratelimiterServer.URL
			if tt.haveURL != "" {
				url = tt.haveURL
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
				t.FailNow()
			}

			req, err := http.NewRequest(http.MethodGet, "some.url.com", nil)
			if err != nil {
				t.FailNow()
			}
			req.SetBasicAuth("user", "pass")

			rw := httptest.NewRecorder()
			rl.ServeHTTP(rw, req)

			if tt.expectCall != called || tt.expectResponse != rw.Code || tt.expectBody != rw.Body.String() {
				t.FailNow()
			}
		})
	}
}
