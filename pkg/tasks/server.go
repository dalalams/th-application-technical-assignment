package tasks

import (
	"context"
	"fmt"
	"log/slog"
	"th-application-technical-assignment/pkg/search"
	"time"

	"github.com/hibiken/asynq"
	"github.com/pkg/errors"
)

type Server struct {
	server  *asynq.Server
	handler *Handler
	mux     *asynq.ServeMux
}

func NewServer(redisCfg *RedisConfig, queueCfg *QueueConfig, searchClient search.Searcher, searchConfig *search.Config) (*Server, error) {
	redisOpt := asynq.RedisClientOpt{
		Addr:     redisCfg.RedisAddr,
		Password: redisCfg.RedisPassword,
		DB:       redisCfg.RedisDB,
	}

	server := asynq.NewServer(redisOpt, asynq.Config{
		Concurrency:  queueCfg.Concurrency,
		RetryDelayFunc: func(n int, err error, task *asynq.Task) time.Duration {
			return queueCfg.RetryDelay
		},
	})

	handler := NewHandler(searchClient, searchConfig)
	mux := asynq.NewServeMux()

	mux.HandleFunc(TypeIndexSeries, handler.HandleIndexSeries)
	mux.HandleFunc(TypeIndexEpisode, handler.HandleIndexEpisode)
	mux.HandleFunc(TypeDeleteSeries, handler.HandleDeleteSeries)
	mux.HandleFunc(TypeDeleteEpisode, handler.HandleDeleteEpisode)

	return &Server{
		server:  server,
		handler: handler,
		mux:     mux,
	}, nil
}

func (s *Server) Start(ctx context.Context) error {
	slog.InfoContext(ctx, "starting asynq server")
	
	if err := s.initializeIndices(ctx); err != nil {
		return errors.Wrap(err, "failed to initialize search indices")
	}

	return s.server.Start(s.mux)
}

func (s *Server) Shutdown() {
	slog.Info("shutting down asynq server")
	s.server.Shutdown()
}

func (s *Server) initializeIndices(ctx context.Context) error {
	indices := map[string]string{
		fmt.Sprintf("%s-series", s.handler.config.IndexPrefix):   search.SeriesMapping,
		fmt.Sprintf("%s-episodes", s.handler.config.IndexPrefix): search.EpisodeMapping,
	}

	for index, mapping := range indices {
		exists, err := s.handler.searchClient.IndexExists(ctx, index)
		if err != nil {
			return errors.Wrapf(err, "failed to check if index %s exists", index)
		}

		if !exists {
			err = s.handler.searchClient.CreateIndex(ctx, index, mapping)
			if err != nil {
				return errors.Wrapf(err, "failed to create index %s", index)
			}
			slog.InfoContext(ctx, "created search index", "index", index)
		}
	}

	return nil
}
