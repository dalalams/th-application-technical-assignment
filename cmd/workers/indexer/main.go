package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"th-application-technical-assignment/pkg/search"
	"th-application-technical-assignment/pkg/tasks"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Redis  tasks.RedisConfig `envPrefix:"REDIS_"`
	Queue  tasks.QueueConfig `envPrefix:"QUEUE_"`
	Search search.Config     `envPrefix:"OPENSEARCH_"`
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		slog.ErrorContext(ctx, "failed to parse config", "err", err)
		os.Exit(1)
	}

	searchClient, err := search.NewClient(&cfg.Search)
	if err != nil {
		slog.ErrorContext(ctx, "failed to create search client", "err", err)
		os.Exit(1)
	}

	queueServer, err := tasks.NewServer(&cfg.Redis, &cfg.Queue, searchClient, &cfg.Search)
	if err != nil {
		slog.ErrorContext(ctx, "failed to create queue server", "err", err)
		os.Exit(1)
	}

	go func() {
		if err := queueServer.Start(ctx); err != nil {
			slog.ErrorContext(ctx, "queue server error", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	slog.InfoContext(shutdownCtx, "shutting down worker...")

	// Shutdown server
	queueServer.Shutdown()
	slog.InfoContext(shutdownCtx, "worker stopped")
}
