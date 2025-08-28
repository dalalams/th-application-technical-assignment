package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"th-application-technical-assignment/pkg/search"

	"github.com/hibiken/asynq"
	"github.com/pkg/errors"
)

type Handler struct {
	searchClient search.Searcher
	config       *search.Config
}

func NewHandler(searchClient search.Searcher, config *search.Config) *Handler {
	return &Handler{
		searchClient: searchClient,
		config:       config,
	}
}

func (h *Handler) HandleIndexSeries(ctx context.Context, t *asynq.Task) error {
	var payload IndexSeriesPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return errors.Wrap(err, "failed to unmarshal payload")
	}

	series := payload.Series

	doc := search.SeriesDocument{
		ID:          series.ID.String(),
		Title:       series.Title,
		Description: series.Description,
		CategoryID:  series.CategoryID.String(),
		Language:    series.Language,
		Type:        string(series.SeriesType),
		CreatedAt:   series.CreatedAt,
		UpdatedAt:   series.UpdatedAt,
	}

	docJSON, err := doc.ToJSON()
	if err != nil {
		return errors.Wrap(err, "failed to convert document to JSON")
	}

	index := fmt.Sprintf("%s-series", h.config.IndexPrefix)
	err = h.searchClient.IndexDocument(ctx, index, series.ID.String(), docJSON)
	if err != nil {
		return errors.Wrap(err, "failed to index series document")
	}

	slog.InfoContext(ctx, "indexed series document", "series_id", series.ID.String(), "index", index)
	return nil
}

func (h *Handler) HandleIndexEpisode(ctx context.Context, t *asynq.Task) error {
	var payload IndexEpisodePayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return errors.Wrap(err, "failed to unmarshal payload")
	}

	episode := payload.Episode
	assets := payload.Assets

	doc := search.EpisodeDocument{
		ID:              episode.ID.String(),
		SeriesID:        episode.SeriesID.String(),
		Title:           episode.Title,
		Description:     episode.Description,
		DurationSeconds: episode.DurationSeconds,
		PublishDate:     episode.PublishDate,
		CreatedAt:       episode.CreatedAt,
		UpdatedAt:       episode.UpdatedAt,
	}

	for _, a := range assets {
		doc.Assets = append(doc.Assets, search.AssetDocument{
			ID:        a.ID.String(),
			AssetType: a.AssetType,
			MimeType:  a.MimeType,
			SizeBytes: a.SizeBytes,
			URL:       a.Url,
		})
	}

	docJSON, err := doc.ToJSON()
	if err != nil {
		return errors.Wrap(err, "failed to convert document to JSON")
	}

	index := fmt.Sprintf("%s-episodes", h.config.IndexPrefix)
	err = h.searchClient.IndexDocument(ctx, index, episode.ID.String(), docJSON)
	if err != nil {
		return errors.Wrap(err, "failed to index episode document")
	}

	slog.InfoContext(ctx, "indexed episode document", "episode_id", episode.ID.String(), "index", index)
	return nil
}

func (h *Handler) HandleDeleteSeries(ctx context.Context, t *asynq.Task) error {
	var payload DeleteSeriesPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return errors.Wrap(err, "failed to unmarshal payload")
	}

	index := fmt.Sprintf("%s-series", h.config.IndexPrefix)
	err := h.searchClient.DeleteDocument(ctx, index, payload.SeriesID)
	if err != nil {
		return errors.Wrap(err, "failed to delete series document")
	}

	slog.InfoContext(ctx, "deleted series document", "series_id", payload.SeriesID, "index", index)
	return nil
}

func (h *Handler) HandleDeleteEpisode(ctx context.Context, t *asynq.Task) error {
	var payload DeleteEpisodePayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return errors.Wrap(err, "failed to unmarshal payload")
	}

	index := fmt.Sprintf("%s-episodes", h.config.IndexPrefix)
	err := h.searchClient.DeleteDocument(ctx, index, payload.EpisodeID)
	if err != nil {
		return errors.Wrap(err, "failed to delete episode document")
	}

	slog.InfoContext(ctx, "deleted episode document", "episode_id", payload.EpisodeID, "index", index)
	return nil
}
