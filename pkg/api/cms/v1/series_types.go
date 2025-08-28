package v1

import (
	"th-application-technical-assignment/pkg/util"
	"time"
)

type PaginatedSeriesResponse = util.PaginatedResponse[SeriesResponse]

type SeriesResponse struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description *string   `json:"description,omitempty"`
	CategoryID  string    `json:"category_id"`
	Language    *string   `json:"language,omitempty"`
	Type        string    `json:"type"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}


type CreateSeriesRequest struct {
	Title       string  `json:"title" validate:"required,min=1,max=255"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=1000"`
	CategoryID  string `json:"category_id" validate:"required,uuid"`
	Language    *string `json:"language,omitempty" validate:"omitempty,min=2,max=10"`
	Type        string  `json:"type" validate:"required,oneof=documentary podcast"`
}

type UpdateSeriesRequest struct {
	Title       string  `json:"title" validate:"required,min=1,max=255"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=1000"`
	CategoryID  *string `json:"category_id,omitempty" validate:"omitempty,uuid"`
	Language    *string `json:"language,omitempty" validate:"omitempty,min=2,max=10"`
	Type        string  `json:"type" validate:"required,oneof=documentary podcast"`
}
