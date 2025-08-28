package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"th-application-technical-assignment/pkg/database"
	"th-application-technical-assignment/pkg/tasks"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/hibiken/asynq"
)

type Config struct {
	Redis    tasks.RedisConfig `envPrefix:"REDIS_"`
	Queue    tasks.QueueConfig `envPrefix:"QUEUE_"`
	Database database.Config   `envPrefix:"DB_"`
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		slog.ErrorContext(ctx, "failed to parse config", "err", err)
		os.Exit(1)
	}

	p, err := database.NewPgPoolFromCfg(ctx, &cfg.Database)
	if err != nil {
		slog.ErrorContext(ctx, "failed to create database pool", "err", err)
	}

	store := database.New(ctx, p)
	defer store.Close(ctx)

	redisOpt := asynq.RedisClientOpt{
		Addr:     cfg.Redis.RedisAddr,
		Password: cfg.Redis.RedisPassword,
		DB:       cfg.Redis.RedisDB,
	}

	client, err := tasks.NewClient(&cfg.Redis)
	if err != nil {
		slog.ErrorContext(ctx, "failed to create queue client", "err", err)
		os.Exit(1)
	}
	defer func() {
		if err := client.Close(); err != nil {
			slog.ErrorContext(ctx, "failed to close queue client", "err", err)
		}
	}()

	srv := asynq.NewServer(redisOpt, asynq.Config{
		Concurrency: cfg.Queue.Concurrency,
		RetryDelayFunc: func(n int, err error, task *asynq.Task) time.Duration {
			return cfg.Queue.RetryDelay
		},
	})

	mux := asynq.NewServeMux()

	importProcessor := tasks.NewImportEpisodeTaskProcessor(store, client)

	mux.Handle(tasks.TypeImportContent, importProcessor)

	go func() {
		if err := srv.Start(mux); err != nil {
			slog.ErrorContext(ctx, "queue server error", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	slog.InfoContext(shutdownCtx, "shutting down worker...")

	srv.Shutdown()
	slog.InfoContext(shutdownCtx, "worker stopped")
}
