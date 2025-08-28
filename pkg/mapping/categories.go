package mapping

import (
	"th-application-technical-assignment/pkg/api/cms/v1"
	"th-application-technical-assignment/sqlc"
)

func Category(c sqlc.Category) v1.CategoryResponse {
	return v1.CategoryResponse{
		ID:   c.ID.String(),
		Slug: c.Slug,
	}
}

