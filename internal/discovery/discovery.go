package discovery

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"th-application-technical-assignment/pkg/search"
	"th-application-technical-assignment/pkg/telemetry"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"github.com/riandyrn/otelchi"
	slogchi "github.com/samber/slog-chi"
	httpSwagger "github.com/swaggo/http-swagger"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type Server struct {
	Config      *Config
	Validator   *validator.Validate
	Searcher    search.Searcher
	Router      chi.Router
	Middlewares chi.Middlewares
	Telemetry   *sdktrace.TracerProvider
}

func NewServer(ctx context.Context) (*Server, error) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	loggerCfg := &slogchi.Config{WithRequestID: true, WithSpanID: true, WithTraceID: true}

	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, errors.Wrap(err, "failed to parse config")
	}
	slog.DebugContext(ctx, "config", "cfg", cfg)

	tp, err := telemetry.InitTracer(ctx, &cfg.Telemetry)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize tracer")
	}

	s, err := search.NewClient(&cfg.Search)
	if err != nil {
		slog.ErrorContext(ctx, "failed to create search client", "err", err)
		os.Exit(1)
	}

	r := chi.NewRouter()

	mw := chi.Chain(
		otelchi.Middleware("cms-service", otelchi.WithChiRoutes(r)),
		middleware.RequestID,
		slogchi.NewWithConfig(logger, *loggerCfg),
		middleware.Recoverer,
		middleware.AllowContentType("application/json"),
		middleware.CleanPath,
		httprate.LimitByIP(100, 1*time.Minute),
	)

	return &Server{
		Config:      &cfg,
		Validator:   validator.New(),
		Router:      r,
        Searcher:    s,
		Middlewares: mw,
		Telemetry:   tp,
	}, nil
}

func (s *Server) MountMiddlewares(ctx context.Context) {
	s.Router.Use(s.Middlewares...)
}

func (s *Server) MountRoutes(ctx context.Context) {
	h := &Handler{s.Validator, s.Searcher}
	s.Router.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:4000/swagger/doc.json"),
	))

	s.Router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/index.html", http.StatusFound)
	})

	s.Router.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	s.Router.Mount("/api", Routes(ctx, h))
}

func (s *Server) Close(ctx context.Context) {
	if s.Telemetry != nil {
		if err := telemetry.Shutdown(ctx, s.Telemetry); err != nil {
			slog.ErrorContext(ctx, "failed to shutdown tracer", "err", err)
		}
	}
}
