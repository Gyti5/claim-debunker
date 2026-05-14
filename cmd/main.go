package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"claim-debunker/config"
	"claim-debunker/pkg/composition"
	"claim-debunker/pkg/server"
	"claim-debunker/pkg/status"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load("config/base.yaml")
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	compositionService := composition.NewService(
		composition.NewOpenAIChatGPTClient(composition.OpenAIClientConfig{
			APIKey:  cfg.OpenAI.APIKey,
			Model:   cfg.OpenAI.Model,
			BaseURL: cfg.OpenAI.BaseURL,
			Timeout: cfg.OpenAI.Timeout,
		}),
	)
	compositionHandler := composition.NewHandlerWithMaxUploadBytes(
		compositionService,
		cfg.Composition.MaxUploadBytes,
	)

	httpAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	httpServer := server.New(
		httpAddr,
		status.NewHandler(),
		compositionHandler,
		server.Options{
			CompositionAPIKey:             cfg.Security.CompositionAPIKey,
			CompositionRateLimitPerMinute: cfg.Security.CompositionRateLimitPerMinute,
		},
	)

	serverErrCh := make(chan error, 1)
	go func() {
		log.Printf("http server listening on %s", httpAddr)
		if err := httpServer.ListenAndServe(); err != nil {
			serverErrCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		log.Printf("shutdown signal received")
	case err := <-serverErrCh:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("http server failed: %v", err)
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("http server shutdown: %v", err)
	}
}
