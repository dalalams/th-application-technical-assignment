package v1

import "th-application-technical-assignment/pkg/util"

type PaginatedCategoryResponse = util.PaginatedResponse[CategoryResponse]

type CategoryResponse struct {
	ID   string `json:"id"`
	Slug string `json:"slug"`
}

type CreateCategoryRequest struct {
	Name string `json:"name" validate:"required,min=1,max=100"`
}

type UpdateCategoryRequest struct {
	Name string `json:"name" validate:"required,min=1,max=100"`
}
