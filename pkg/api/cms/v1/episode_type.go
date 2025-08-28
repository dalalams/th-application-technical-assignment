package v1

import (
	"th-application-technical-assignment/pkg/util"
	"time"
)

type PaginatedEpisodeResponse = util.PaginatedResponse[EpisodeResponse]

type EpisodeResponse struct {
	ID              string     `json:"id"`
	SeriesID        string     `json:"series_id"`
	Title           string     `json:"title"`
	Description     *string    `json:"description,omitempty"`
	DurationSeconds *int32     `json:"duration_seconds,omitempty"`
	PublishDate     *time.Time `json:"publish_date,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	Assets          []EpisodeAssetResponse `json:"assets"`
}

type CreateEpisodeRequest struct {
	SeriesID        string     `json:"series_id" validate:"required,uuid"`
	Title           string     `json:"title" validate:"required,min=1,max=255"`
	Description     *string    `json:"description,omitempty" validate:"omitempty,max=2000"`
	DurationSeconds *int32     `json:"duration_seconds,omitempty" validate:"omitempty,min=0,max=86400"`
	PublishDate     *time.Time `json:"publish_date,omitempty"`
}

type CreateEpisodeResponse struct {
	ID string `json:"id"`
}

type UpdateEpisodeRequest struct {
	Title           string     `json:"title" validate:"required,min=1,max=255"`
	Description     *string    `json:"description,omitempty" validate:"omitempty,max=2000"`
	DurationSeconds *int32     `json:"duration_seconds,omitempty" validate:"omitempty,min=0,max=86400"`
	PublishDate     *time.Time `json:"publish_date,omitempty"`
}
