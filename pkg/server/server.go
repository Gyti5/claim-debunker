package server

import (
	"net/http"
	"time"
)

type StatusEndpoint interface {
	Ping(http.ResponseWriter, *http.Request)
}

type CompositionEndpoint interface {
	GetComposition(http.ResponseWriter, *http.Request)
}

type Options struct {
	CompositionAPIKey             string
	CompositionRateLimitPerMinute int

	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
}

func New(
	addr string,
	statusEndpoint StatusEndpoint,
	compositionEndpoint CompositionEndpoint,
	options Options,
) *http.Server {
	options = withDefaultOptions(options)
	mux := http.NewServeMux()

	if statusEndpoint != nil {
		mux.HandleFunc("GET /ping", statusEndpoint.Ping)
	}
	if compositionEndpoint != nil {
		compositionHandler := http.Handler(http.HandlerFunc(compositionEndpoint.GetComposition))
		compositionRateLimiter := newClientRateLimiter(options.CompositionRateLimitPerMinute, time.Minute)
		compositionHandler = withRateLimit(compositionRateLimiter, compositionHandler)
		compositionHandler = withAPIKeyAuth(options.CompositionAPIKey, compositionHandler)
		mux.Handle("POST /get-composition", compositionHandler)
	}

	return &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: options.ReadHeaderTimeout,
		ReadTimeout:       options.ReadTimeout,
		WriteTimeout:      options.WriteTimeout,
		IdleTimeout:       options.IdleTimeout,
	}
}

func withDefaultOptions(options Options) Options {
	if options.CompositionRateLimitPerMinute <= 0 {
		options.CompositionRateLimitPerMinute = 60
	}
	if options.ReadHeaderTimeout <= 0 {
		options.ReadHeaderTimeout = 5 * time.Second
	}
	if options.ReadTimeout <= 0 {
		options.ReadTimeout = 30 * time.Second
	}
	if options.WriteTimeout <= 0 {
		options.WriteTimeout = 30 * time.Second
	}
	if options.IdleTimeout <= 0 {
		options.IdleTimeout = 60 * time.Second
	}

	return options
}
