package v1

import "time"

type EpisodeAssetResponse struct {
	ID        string    `json:"id"`
	EpisodeID string    `json:"episode_id"`
	AssetType string    `json:"asset_type"`
	MimeType  string    `json:"mime_type"`
	SizeBytes *int64    `json:"size_bytes,omitempty"`
	URL       *string   `json:"url,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateEpisodeAssetRequest struct {
	EpisodeID string `json:"episode_id" validate:"required,uuid"`
	AssetType string `json:"asset_type" validate:"required,oneof=audio video thumbnail"`
	MimeType  string `json:"mime_type" validate:"required,min=3,max=100"`
	SizeBytes *int64 `json:"size_bytes,omitempty" validate:"omitempty,min=0"`
	URL       *string `json:"url,omitempty" validate:"omitempty,url"`
}
