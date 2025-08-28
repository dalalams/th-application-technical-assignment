package tasks

import (
	"context"
	"encoding/json"
	"log/slog"
	"th-application-technical-assignment/sqlc"

	"github.com/hibiken/asynq"
	"github.com/pkg/errors"
)

const (
	TypeIndexSeries    = "search:index_series"
	TypeIndexEpisode   = "search:index_episode"
	TypeDeleteSeries   = "search:delete_series"
	TypeDeleteEpisode  = "search:delete_episode"
	TypeImportContent  = "import:content"
)

type TaskQueue interface {
    Enqueue(ctx context.Context, typename string, taskPayload any) error
    EnqueueIndexSeries(ctx context.Context, series sqlc.Series) error
    EnqueueIndexEpisode(ctx context.Context, episode sqlc.Episode, assets []sqlc.EpisodeAsset) error
    EnqueueDeleteSeries(ctx context.Context, seriesID string) error
    EnqueueDeleteEpisode(ctx context.Context, episodeID string) error
    EnqueueImportContent(ctx context.Context, payload ImportContentPayload) error
    Close() error
}

type AsynqQueue struct {
	client *asynq.Client
}

func NewClient(cfg *RedisConfig) (*AsynqQueue, error) {
	if cfg == nil {
		return nil, errors.New("redis config is nil")
	}

	redisOpt := asynq.RedisClientOpt{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	}

	client := asynq.NewClient(redisOpt)
	if err := client.Ping(); err != nil {
		return nil, errors.Wrap(err, "failed to ping asynq client")
	}

	return &AsynqQueue{client: client}, nil
}

func (c *AsynqQueue) Close() error {
	return c.client.Close()
}

type IndexSeriesPayload struct {
	Series sqlc.Series `json:"series"`
}

type IndexEpisodePayload struct {
	Episode sqlc.Episode       `json:"episode"`
	Assets  []sqlc.EpisodeAsset `json:"assets"`
}

type DeleteSeriesPayload struct {
	SeriesID string `json:"series_id"`
}

type DeleteEpisodePayload struct {
	EpisodeID string `json:"episode_id"`
}

type ImportContentPayload struct {
	SourceType string `json:"source_type"`
	SourceURL  string `json:"source_url"`
	SeriesID   string `json:"series_id"`
}

func (c *AsynqQueue) Enqueue(ctx context.Context, typename string, taskPayload any) error {
	data, err := json.Marshal(taskPayload)
	if err != nil {
		return errors.Wrap(err, "failed to marshal payload")
	}

	t := asynq.NewTask(typename, data)
	_, err = c.client.EnqueueContext(ctx, t)
	if err != nil {
		slog.ErrorContext(ctx, "failed to enqueue task", "err", err, "typename", typename)
		return errors.Wrap(err, "failed to enqueue task")
	}

	slog.InfoContext(ctx, "enqueued task", "typename", typename)
	return nil
}

func (c *AsynqQueue) EnqueueIndexSeries(ctx context.Context, series sqlc.Series) error {
	payload := IndexSeriesPayload{Series: series}
	return c.Enqueue(ctx, TypeIndexSeries, payload)
}

func (c *AsynqQueue) EnqueueIndexEpisode(ctx context.Context, episode sqlc.Episode, assets []sqlc.EpisodeAsset) error {
	payload := IndexEpisodePayload{Episode: episode, Assets: assets}
	return c.Enqueue(ctx, TypeIndexEpisode, payload)
}

func (c *AsynqQueue) EnqueueDeleteSeries(ctx context.Context, seriesID string) error {
	payload := DeleteSeriesPayload{SeriesID: seriesID}
	return c.Enqueue(ctx, TypeDeleteSeries, payload)
}

func (c *AsynqQueue) EnqueueDeleteEpisode(ctx context.Context, episodeID string) error {
	payload := DeleteEpisodePayload{EpisodeID: episodeID}
	return c.Enqueue(ctx, TypeDeleteEpisode, payload)
}

func (c *AsynqQueue) EnqueueImportContent(ctx context.Context, payload ImportContentPayload) error {
	return c.Enqueue(ctx, TypeImportContent, payload)
}
