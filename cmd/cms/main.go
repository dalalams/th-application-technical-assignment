package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	_ "th-application-technical-assignment/docs/cms"
	"th-application-technical-assignment/internal/cms"

	"github.com/pkg/errors"
)

func main() {
	ctx := context.Background()
	if err := run(ctx); err != nil {
		slog.ErrorContext(ctx, "failed to run server", "err", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	srv, err := cms.NewServer(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to create server")
	}

    srv.MountMiddlewares(ctx)
    srv.MountRoutes(ctx)

	server := &http.Server{
		Addr:    srv.Config.HTTP.Addr,
		Handler: srv.Router,
	}

	go func() {
		slog.InfoContext(ctx, "starting server", "addr", srv.Config.HTTP.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.ErrorContext(ctx, "server error", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.InfoContext(ctx, "shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	srv.Close(shutdownCtx)
	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.ErrorContext(ctx, "server shutdown error", "err", err)
		os.Exit(1)
	}

	slog.InfoContext(shutdownCtx, "server shutdown")
	return nil
}
