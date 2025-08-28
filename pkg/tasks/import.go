package tasks

import (
	"context"
	"encoding/json"
	"log/slog"
	"th-application-technical-assignment/pkg/database"
	"th-application-technical-assignment/pkg/importer"
	"th-application-technical-assignment/sqlc"

	"github.com/hibiken/asynq"
	"github.com/pkg/errors"
)

type ImportEpisodeTaskProcessor struct {
	store *database.Store
	queue TaskQueue
}

func NewImportEpisodeTaskProcessor(store *database.Store, queue TaskQueue) *ImportEpisodeTaskProcessor {
	return &ImportEpisodeTaskProcessor{store, queue}
}

func (p *ImportEpisodeTaskProcessor) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var payload ImportContentPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return errors.Wrap(err, "failed to unmarshal payload")
	}

	var i importer.Importer
	i, err := importer.GetImporter(payload.SourceType)
	if err != nil {
		return errors.Wrap(err, "unsupported import source")
	}

	ep, asset, err := i.FetchEpisode(ctx, payload.SourceURL, payload.SeriesID)
	if err != nil {
		return errors.Wrap(err, "failed to fetch episodes")
	}

	params := sqlc.CreateEpisodeParams{
		SeriesID: ep.SeriesID,
		Title:    ep.Title,
	}

	episode, err := p.store.Queries.CreateEpisode(ctx, params)
	if err != nil {
		return errors.Wrap(err, "failed to create episode")
	}

	asset.EpisodeID = episode.ID
	assetParams := sqlc.CreateAssetParams{
		EpisodeID: asset.EpisodeID,
		AssetType: asset.AssetType,
		MimeType:  asset.MimeType,
		Url:       asset.Url,
	}

	dbAsset, err := p.store.Queries.CreateAsset(ctx, assetParams)
	if err != nil {
		return errors.Wrap(err, "failed to create asset")
	}

	if err := p.queue.EnqueueIndexEpisode(ctx, episode, []sqlc.EpisodeAsset{dbAsset}); err != nil {
		return errors.Wrap(err, "failed to enqueue index episode task")
	}

	slog.InfoContext(ctx, "imported episode", "episode_id", episode.ID.String())
	return nil
}
