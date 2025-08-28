package mapping

import (
	"th-application-technical-assignment/pkg/api/cms/v1"
	"th-application-technical-assignment/sqlc"
)

func Series(s sqlc.Series) v1.SeriesResponse {
	resp := v1.SeriesResponse{
		ID:          s.ID.String(),
		Title:       s.Title,
		CategoryID:  s.CategoryID.String(),
		Type:        s.SeriesType,
		Description: s.Description,
		Language:    s.Language,
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.UpdatedAt,
	}

	return resp
}
