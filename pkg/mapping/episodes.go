package mapping

import (
	"th-application-technical-assignment/pkg/api/cms/v1"
	"th-application-technical-assignment/sqlc"

	"github.com/google/uuid"
)

func Episode(ep sqlc.Episode, assets []sqlc.EpisodeAsset) v1.EpisodeResponse {
	resp := v1.EpisodeResponse{
		ID:              ep.ID.String(),
		SeriesID:        ep.SeriesID.String(),
		Description:     ep.Description,
		Title:           ep.Title,
		DurationSeconds: ep.DurationSeconds,
		PublishDate:     ep.PublishDate,
		CreatedAt:       ep.CreatedAt,
		UpdatedAt:       ep.UpdatedAt,
	}

	for _, a := range assets {
		resp.Assets = append(resp.Assets, v1.EpisodeAssetResponse{
			ID:        a.ID.String(),
			EpisodeID: a.EpisodeID.String(),
			AssetType: a.AssetType,
			MimeType:  a.MimeType,
			SizeBytes: a.SizeBytes,
			URL:       a.Url,
			CreatedAt: a.CreatedAt,
		})
	}

	return resp
}

func EpisodesWithAssets(rows []sqlc.ListEpisodesWithAssetsBySeriesPaginatedRow) []v1.EpisodeResponse {
	episodeMap := make(map[uuid.UUID]*v1.EpisodeResponse)

	for _, row := range rows {
		ep, ok := episodeMap[row.EpisodeID]
		if !ok {
			ep = &v1.EpisodeResponse{
				ID:              row.EpisodeID.String(),
				SeriesID:        row.SeriesID.String(),
				Title:           row.Title,
				Description:     row.Description,
				DurationSeconds: row.DurationSeconds,
				PublishDate:     row.PublishDate,
				CreatedAt:       row.EpisodeCreatedAt,
				UpdatedAt:       row.EpisodeUpdatedAt,
				Assets:          []v1.EpisodeAssetResponse{},
			}
			episodeMap[row.EpisodeID] = ep
		}

		if row.AssetID != nil {
				ep.Assets = append(ep.Assets, v1.EpisodeAssetResponse{
					ID:        (*row.AssetID).String(),
					EpisodeID: row.EpisodeID.String(),
					AssetType: *row.AssetType,
					MimeType:  *row.MimeType,
					SizeBytes: row.SizeBytes,
					URL:       row.Url,
					CreatedAt: *row.AssetCreatedAt,
				})
			}
	}

	episodes := make([]v1.EpisodeResponse, 0, len(episodeMap))
	for _, ep := range episodeMap {
		episodes = append(episodes, *ep)
	}

	return episodes
}