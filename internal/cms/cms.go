package cms

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"th-application-technical-assignment/pkg/database"
	"th-application-technical-assignment/pkg/storage"
	"th-application-technical-assignment/pkg/tasks"
	"th-application-technical-assignment/pkg/telemetry"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"
	"github.com/go-playground/validator/v10"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/pkg/errors"
	"github.com/riandyrn/otelchi"
	slogchi "github.com/samber/slog-chi"
	httpSwagger "github.com/swaggo/http-swagger"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type Server struct {
    Config      *Config
	Store       *database.Store
	Queue       tasks.TaskQueue
	Storage     storage.ObjectStorage
	Auth        *jwtauth.JWTAuth
	Validator   *validator.Validate
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

	p, err := database.NewPgPoolFromCfg(ctx, &cfg.Database)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create database pool")
	}
	s := database.New(ctx, p)

	minioClient, err := storage.NewMinIOClient(&cfg.MinIO)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create minio client")
	}
	if err := minioClient.EnsureBucket(ctx); err != nil {
		return nil, errors.Wrap(err, "failed to ensure bucket existance")
	}

	tasksClient, err := tasks.NewClient(&cfg.Redis)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create queue client")
	}

	tokenAuth := jwtauth.New(cfg.Auth.SigningAlg, []byte(cfg.Auth.SigningKey), cfg.Auth.VerificationKey, jwt.WithAcceptableSkew(30*time.Second))

	tp, err := telemetry.InitTracer(ctx, &cfg.Telemetry)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize tracer")
	}

	r := chi.NewRouter()

	mw := chi.Chain(
		otelchi.Middleware("cms-service", otelchi.WithChiRoutes(r)),
		middleware.RequestID,
		slogchi.NewWithConfig(logger, *loggerCfg),
		middleware.Recoverer,
		middleware.AllowContentType("application/json"),
		middleware.CleanPath,
		// jwtauth.Verifier(tokenAuth),
		// jwtauth.Authenticator(tokenAuth),
	)

	return &Server{
        Config:      &cfg,
		Store:       s,
		Queue:       tasksClient,
		Storage:     minioClient,
		Auth:        tokenAuth,
		Validator:   validator.New(),
		Router:      r,
		Middlewares: mw,
        Telemetry:   tp,
	}, nil
}

func (s *Server) MountMiddlewares(ctx context.Context) {
	s.Router.Use(s.Middlewares...)
}

func (s *Server) MountRoutes(ctx context.Context) {
	h := &Handler{s.Store, s.Validator, s.Queue, s.Storage, s.Auth}
	s.Router.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:3000/swagger/doc.json"),
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
	if s.Store != nil {
		s.Store.Close(ctx)
	}
	if s.Queue != nil {
		if err := s.Queue.Close(); err != nil {
			slog.ErrorContext(ctx, "failed to close queue client", "err", err)
		}
	}
    if s.Telemetry != nil {
        if err := telemetry.Shutdown(ctx, s.Telemetry); err != nil {
            slog.ErrorContext(ctx, "failed to shutdown tracer", "err", err)
        }
    }
}
