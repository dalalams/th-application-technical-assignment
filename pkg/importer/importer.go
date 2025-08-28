package importer

import (
	"context"
	"th-application-technical-assignment/sqlc"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type Importer interface {
	FetchEpisode(ctx context.Context, url, seriesID string) (*sqlc.Episode, *sqlc.EpisodeAsset, error)
}

var importers = map[string]Importer{
	"youtube": NewYouTubeImporter(),
}

func GetImporter(source string) (Importer, error) {
	imp, ok := importers[source]
	if !ok {
		return nil, errors.New("importer not found")
	}
	return imp, nil
}

type YouTubeImporter struct {
}

func NewYouTubeImporter() Importer {
	return &YouTubeImporter{}
}

func (i *YouTubeImporter) FetchEpisode(ctx context.Context, url, seriesID string) (*sqlc.Episode, *sqlc.EpisodeAsset, error) {
	// youtube import logic, skipped
	seriesUuid, err := uuid.Parse(seriesID)
	if err != nil {
		return nil, nil, errors.Wrap(err, "invalid series ID")
	}

	episodeID := uuid.New()
	ep := &sqlc.Episode{
		ID:       episodeID,
		SeriesID: seriesUuid,
		Title:    "YouTube Import",
	}

	asset := &sqlc.EpisodeAsset{
		EpisodeID: episodeID,
		AssetType: "video",
		MimeType:  "video/mp4",
		Url:       &url,
	}

	return ep, asset, nil
}
