package search

import (
	"encoding/json"
	"time"
)

type SeriesDocument struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description *string   `json:"description,omitempty"`
	CategoryID  string    `json:"category_id"`
	Language    *string   `json:"language,omitempty"`
	Type        string    `json:"type"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	IndexedAt   time.Time `json:"indexed_at"`
}

type AssetDocument struct {
	ID        string  `json:"id"`
	AssetType string  `json:"asset_type"`
	MimeType  string  `json:"mime_type"`
	SizeBytes *int64  `json:"size_bytes,omitempty"`
	URL       *string `json:"url,omitempty"`
}

type EpisodeDocument struct {
	ID              string          `json:"id"`
	SeriesID        string          `json:"series_id"`
	Title           string          `json:"title"`
	Description     *string         `json:"description,omitempty"`
	DurationSeconds *int32          `json:"duration_seconds,omitempty"`
	PublishDate     *time.Time      `json:"publish_date,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
	IndexedAt       time.Time       `json:"indexed_at"`
	Assets          []AssetDocument `json:"assets"`
}

func (s SeriesDocument) ToJSON() ([]byte, error) {
	s.IndexedAt = time.Now()
	return json.Marshal(s)
}

func (e EpisodeDocument) ToJSON() ([]byte, error) {
	e.IndexedAt = time.Now()
	return json.Marshal(e)
}
