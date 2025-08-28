package cms

import (
	"context"
	"net/http"
	"th-application-technical-assignment/internal/middleware"
	"th-application-technical-assignment/internal/response"
	"th-application-technical-assignment/pkg/api/cms/v1"
	"th-application-technical-assignment/pkg/mapping"
	"th-application-technical-assignment/pkg/util"
	"th-application-technical-assignment/pkg/validation"
	"th-application-technical-assignment/sqlc"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// listCategories godoc
// @Summary      List all categories with pagination
// @Description  Get a paginated list of all categories available in the system
// @Tags         Categories
// @Accept       json
// @Produce      json
// @Param        page       query     int  false  "Page number (default: 1)"
// @Param        page_size  query     int  false  "Page size (default: 20, max: 100)"
// @Success      200        {object}  v1.PaginatedCategoryResponse
// @Failure      400        {object}  map[string]string
// @Failure      500        {object}  map[string]string
// @Router       /categories [get]
func (h *Handler) listCategories(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pagination := middleware.GetPagination(ctx)
	offset := (pagination.Page - 1) * pagination.PageSize

	fetchCount := func(ctx context.Context) (int64, error) {
		return h.s.Queries.CountCategories(ctx)
	}

	fetchCategories := func(ctx context.Context) ([]sqlc.Category, error) {
		params := sqlc.ListCategoriesPaginatedParams{
			Limit:  int32(pagination.PageSize),
			Offset: int32(offset),
		}
		return h.s.Queries.ListCategoriesPaginated(ctx, params)
	}

	itemCount, dbCategories, err := util.FetchPaginatedData(ctx, fetchCount, fetchCategories)
	if err != nil {
		response.HandleDBError(ctx, w, err, "We couldn't retrieve the categories.")
		return
	}

	categoriesData := make([]v1.CategoryResponse, len(dbCategories))
	for i, c := range dbCategories {
		categoriesData[i] = mapping.Category(c)
	}

	paginationMeta := util.CalculatePaginationResponse(pagination.Page, pagination.PageSize, itemCount)
	res := util.PaginatedResponse[v1.CategoryResponse]{
		Data:       categoriesData,
		Pagination: paginationMeta,
	}

	response.RespondWithJSON(ctx, w, http.StatusOK, res)
}

// getCategory godoc
// @Summary      Get category by ID
// @Description  Get a single category by its ID
// @Tags         Categories
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Category ID"
// @Success      200  {object}  v1.CategoryResponse
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /categories/{id} [get]
func (h *Handler) getCategory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idParam := chi.URLParam(r, "id")
	if idParam == "" {
		response.RespondWithError(ctx, w, http.StatusBadRequest, "Category ID is required.")
		return
	}

	categoryID, err := uuid.Parse(idParam)
	if err != nil {
		response.RespondWithError(ctx, w, http.StatusBadRequest, "Invalid category ID format.")
		return
	}

	dbCategory, err := h.s.Queries.GetCategory(ctx, categoryID)
	if err != nil {
		response.HandleDBError(ctx, w, err, "Category not found.")
		return
	}

	res := mapping.Category(dbCategory)
	response.RespondWithJSON(ctx, w, http.StatusOK, res)
}

// postCategory godoc
// @Summary      Create a new category
// @Description  Create a new category with the provided data
// @Tags         Categories
// @Accept       json
// @Produce      json
// @Param        category  body      v1.CreateCategoryRequest  true  "Category data"
// @Success      201       {object}  v1.CategoryResponse
// @Failure      400       {object}  map[string]string
// @Failure      500       {object}  map[string]string
// @Router       /categories [post]
func (h *Handler) postCategory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req, err := validation.DecodeAndValidate[v1.CreateCategoryRequest](r, h.v)
	if err != nil {
		response.RespondWithError(ctx, w, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	slug := util.CreateSlug(req.Name)
	dbCategory, err := h.s.Queries.CreateCategory(ctx, slug)
	if err != nil {
		response.HandleDBError(ctx, w, err, "We couldn't create the category.")
		return
	}

	res := mapping.Category(dbCategory)
	response.RespondWithJSON(ctx, w, http.StatusCreated, res)
}

// putCategory godoc
// @Summary      Update category by ID
// @Description  Update an existing category with the provided data
// @Tags         Categories
// @Accept       json
// @Produce      json
// @Param        id        path      string                    true  "Category ID"
// @Param        category  body      v1.UpdateCategoryRequest  true  "Category data"
// @Success      200       {object}  v1.CategoryResponse
// @Failure      400       {object}  map[string]string
// @Failure      404       {object}  map[string]string
// @Failure      500       {object}  map[string]string
// @Router       /categories/{id} [put]
func (h *Handler) putCategory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idParam := chi.URLParam(r, "id")
	if idParam == "" {
		response.RespondWithError(ctx, w, http.StatusBadRequest, "Category ID is required.")
		return
	}

	categoryID, err := uuid.Parse(idParam)
	if err != nil {
		response.RespondWithError(ctx, w, http.StatusBadRequest, "Invalid category ID format.")
		return
	}

	req, err := validation.DecodeAndValidate[v1.UpdateCategoryRequest](r, h.v)
	if err != nil {
		response.RespondWithError(ctx, w, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	params := sqlc.UpdateCategoryParams{
		ID:   categoryID,
		Slug: util.CreateSlug(req.Name),
	}

	dbCategory, err := h.s.Queries.UpdateCategory(ctx, params)
	if err != nil {
		response.HandleDBError(ctx, w, err, "Category not found.")
		return
	}

	res := mapping.Category(dbCategory)
	response.RespondWithJSON(ctx, w, http.StatusOK, res)
}

// deleteCategory godoc
// @Summary      Delete category by ID
// @Description  Soft delete a category by its ID
// @Tags         Categories
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Category ID"
// @Success      204  "No Content"
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /categories/{id} [delete]
func (h *Handler) deleteCategory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idParam := chi.URLParam(r, "id")
	if idParam == "" {
		response.RespondWithError(ctx, w, http.StatusBadRequest, "Category ID is required.")
		return
	}

	categoryID, err := uuid.Parse(idParam)
	if err != nil {
		response.RespondWithError(ctx, w, http.StatusBadRequest, "Invalid category ID format.")
		return
	}

	err = h.s.Queries.DeleteCategory(ctx, categoryID)
	if err != nil {
		response.HandleDBError(ctx, w, err, "We couldn't delete the category.")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
