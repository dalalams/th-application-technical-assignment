package cms

import (
	"context"
	"log/slog"
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

// listSeries godoc
// @Summary      List all series with pagination
// @Description  Get a paginated list of all series available in the system
// @Tags         Series
// @Accept       json
// @Produce      json
// @Param        page       query     int  false  "Page number (default: 1)"
// @Param        page_size  query     int  false  "Page size (default: 20, max: 100)"
// @Success      200        {object}  v1.PaginatedSeriesResponse
// @Failure      400        {object}  map[string]string
// @Failure      500        {object}  map[string]string
// @Router       /series [get]
func (h *Handler) listSeries(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	pagination := middleware.GetPagination(ctx)
	offset := (pagination.Page - 1) * pagination.PageSize

	fetchCount := func(ctx context.Context) (int64, error) {
		return h.s.Queries.CountSeries(ctx)
	}

	fetchSeries := func(ctx context.Context) ([]sqlc.Series, error) {
		params := sqlc.ListSeriesPaginatedParams{
			Limit:  int32(pagination.PageSize),
			Offset: int32(offset),
		}
		return h.s.Queries.ListSeriesPaginated(ctx, params)
	}

	itemCount, dbSeries, err := util.FetchPaginatedData(ctx, fetchCount, fetchSeries)
	if err != nil {
		response.HandleDBError(ctx, w, err, "We couldn't retrieve the series.")
		return
	}

	seriesData := make([]v1.SeriesResponse, len(dbSeries))
	for i, s := range dbSeries {
		seriesData[i] = mapping.Series(s)
	}

	paginationMeta := util.CalculatePaginationResponse(pagination.Page, pagination.PageSize, itemCount)
	res := util.PaginatedResponse[v1.SeriesResponse]{
		Data:       seriesData,
		Pagination: paginationMeta,
	}

	response.RespondWithJSON(ctx, w, http.StatusOK, res)
}

// getSeries godoc
// @Summary      Get series by ID
// @Description  Get a single series by its ID
// @Tags         Series
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Series ID"
// @Success      200  {object}  v1.SeriesResponse
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /series/{id} [get]
func (h *Handler) getSeries(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idParam := chi.URLParam(r, "id")
	if idParam == "" {
		response.RespondWithError(ctx, w, http.StatusBadRequest, "series id is required.")
		return
	}

	seriesID, err := uuid.Parse(idParam)
	if err != nil {
		response.RespondWithError(ctx, w, http.StatusBadRequest, "invalid series id format.")
		return
	}

	dbSeries, err := h.s.Queries.GetSeries(ctx, seriesID)
	if err != nil {
		response.HandleDBError(ctx, w, err, "Series not found.")
		return
	}

	res := mapping.Series(dbSeries)
	response.RespondWithJSON(ctx, w, http.StatusOK, res)
}

// postSeries godoc
// @Summary      Create a new series
// @Description  Create a new series with the provided data
// @Tags         Series
// @Accept       json
// @Produce      json
// @Param        series  body      v1.CreateSeriesRequest  true  "Series data"
// @Success      201     {object}  v1.SeriesResponse
// @Failure      400     {object}  map[string]string
// @Failure      500     {object}  map[string]string
// @Router       /series [post]
func (h *Handler) postSeries(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req, err := validation.DecodeAndValidate[v1.CreateSeriesRequest](r, h.v)
	if err != nil {
		response.RespondWithError(ctx, w, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	categoryID, err := uuid.Parse(req.CategoryID)
	if err != nil {
		response.RespondWithError(ctx, w, http.StatusBadRequest, "Invalid category ID format.")
		return
	}

	params := sqlc.CreateSeriesParams{
		Title:       req.Title,
		SeriesType:  req.Type,
		CategoryID:  categoryID,
		Description: req.Description,
		Language:    req.Language,
	}

	dbSeries, err := h.s.Queries.CreateSeries(ctx, params)
	if err != nil {
		response.HandleDBError(ctx, w, err, "We couldn't create the series.")
		return
	}

	if err := h.q.EnqueueIndexSeries(ctx, dbSeries); err != nil {
		slog.ErrorContext(ctx, "failed to enqueue index series task", "err", err, "series_id", dbSeries.ID)
	}

	res := mapping.Series(dbSeries)
	response.RespondWithJSON(ctx, w, http.StatusCreated, res)
}

// putSeries godoc
// @Summary      Update series by ID
// @Description  Update an existing series with the provided data
// @Tags         Series
// @Accept       json
// @Produce      json
// @Param        id      path      string                  true  "Series ID"
// @Param        series  body      v1.UpdateSeriesRequest  true  "Series data"
// @Success      200     {object}  v1.SeriesResponse
// @Failure      400     {object}  map[string]string
// @Failure      404     {object}  map[string]string
// @Failure      500     {object}  map[string]string
// @Router       /series/{id} [put]
func (h *Handler) putSeries(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idParam := chi.URLParam(r, "id")
	if idParam == "" {
		response.RespondWithError(ctx, w, http.StatusBadRequest, "Series ID is required.")
		return
	}

	seriesID, err := uuid.Parse(idParam)
	if err != nil {
		response.RespondWithError(ctx, w, http.StatusBadRequest, "Invalid series ID format.")
		return
	}

	req, err := validation.DecodeAndValidate[v1.UpdateSeriesRequest](r, h.v)
	if err != nil {
		response.RespondWithError(ctx, w, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	params := sqlc.UpdateSeriesParams{
		ID:          seriesID,
		Title:       req.Title,
		SeriesType:  req.Type,
		Description: req.Description,
		Language:    req.Language,
	}

	slog.InfoContext(ctx, "categoryID", "category", req.CategoryID)
	if req.CategoryID != nil {
		categoryID, err := uuid.Parse(*req.CategoryID)
		slog.InfoContext(ctx, "categoryID", "category", categoryID)
		if err != nil {
			response.RespondWithError(ctx, w, http.StatusBadRequest, "Invalid category ID format.")
			return
		}
		params.CategoryID = categoryID
	}

	dbSeries, err := h.s.Queries.UpdateSeries(ctx, params)
	if err != nil {
		response.HandleDBError(ctx, w, err, "Series not found.")
		return
	}

	if err := h.q.EnqueueIndexSeries(ctx, dbSeries); err != nil {
		slog.ErrorContext(ctx, "failed to enqueue index series task", "err", err, "series_id", dbSeries.ID)
	}

	res := mapping.Series(dbSeries)
	response.RespondWithJSON(ctx, w, http.StatusOK, res)
}

// deleteSeries godoc
// @Summary      Delete series by ID
// @Description  Soft delete a series by its ID
// @Tags         Series
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Series ID"
// @Success      204  "No Content"
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /series/{id} [delete]
func (h *Handler) deleteSeries(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idParam := chi.URLParam(r, "id")
	if idParam == "" {
		response.RespondWithError(ctx, w, http.StatusBadRequest, "Series ID is required.")
		return
	}

	seriesID, err := uuid.Parse(idParam)
	if err != nil {
		response.RespondWithError(ctx, w, http.StatusBadRequest, "Invalid series ID format.")
		return
	}

	err = h.s.Queries.DeleteSeries(ctx, seriesID)
	if err != nil {
		response.HandleDBError(ctx, w, err, "We couldn't delete the series.")
		return
	}

	if err := h.q.EnqueueDeleteSeries(ctx, seriesID.String()); err != nil {
		slog.ErrorContext(ctx, "failed to enqueue delete series task", "err", err, "series_id", seriesID)
	}

	w.WriteHeader(http.StatusNoContent)
}
