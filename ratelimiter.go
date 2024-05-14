// Package ratelimiter defines a middleware that integrates with ratelimiter service.
package ratelimiter

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
)

// Config the plugin configuration.
type Config struct {
	URL    string `json:"url"`
	DryRun bool   `json:"dryRun"`
}

// Ratelimiter plugin that calls a specified ratelimiter service URL.
type Ratelimiter struct {
	next   http.Handler
	url    string
	dryRun bool
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{}
}

// New created a new ratelimiter plugin.
func New(_ context.Context, next http.Handler, config *Config, _ string) (http.Handler, error) {
	if len(config.URL) == 0 {
		return nil, errors.New("URL cannot be empty")
	}

	return &Ratelimiter{
		next:   next,
		url:    config.URL,
		dryRun: config.DryRun,
	}, nil
}

func (r *Ratelimiter) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	client := http.Client{}
	newReq, err := http.NewRequest(http.MethodGet, r.url, nil)
	if err != nil {
		_, _ = os.Stderr.WriteString(fmt.Sprintf("error creating request: %v", err))
		r.next.ServeHTTP(rw, req)
		return
	}
	newReq.Header = req.Header

	resp, err := client.Do(newReq)
	if err != nil {
		_, _ = os.Stderr.WriteString(fmt.Sprintf("error making request: %v", err))
		r.next.ServeHTTP(rw, req)
		return
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		if !r.dryRun {
			http.Error(rw, "Too many requests", http.StatusTooManyRequests)
			return
		}
		_, _ = os.Stderr.WriteString("dry run: too many requests")
	}

	r.next.ServeHTTP(rw, req)
}
