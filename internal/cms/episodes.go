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

// getSeriesEpisodes godoc
// @Summary      List episodes by series with pagination
// @Description  Get a paginated list of all episodes for a specific series
// @Tags         Episodes
// @Accept       json
// @Produce      json
// @Param        series_id  query     string  true   "Series ID"
// @Param        page       query     int     false  "Page number (default: 1)"
// @Param        page_size  query     int     false  "Page size (default: 20, max: 100)"
// @Success      200        {object}  v1.PaginatedEpisodeResponse
// @Failure      400        {object}  map[string]string
// @Failure      500        {object}  map[string]string
// @Router       /series/episodes [get]
func (h *Handler) listSeriesEpisodes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	seriesIDParam := r.URL.Query().Get("series_id")
	if seriesIDParam == "" {
		response.RespondWithError(ctx, w, http.StatusBadRequest, "Series ID is required.")
		return
	}

	seriesID, err := uuid.Parse(seriesIDParam)
	if err != nil {
		response.RespondWithError(ctx, w, http.StatusBadRequest, "Invalid series ID format.")
		return
	}

	pagination := middleware.GetPagination(ctx)
	offset := (pagination.Page - 1) * pagination.PageSize

	fetchCount := func(ctx context.Context) (int64, error) {
		return h.s.Queries.CountEpisodesBySeries(ctx, seriesID)
	}

	fetchEpisodes := func(ctx context.Context) ([]sqlc.ListEpisodesWithAssetsBySeriesPaginatedRow, error) {
		params := sqlc.ListEpisodesWithAssetsBySeriesPaginatedParams{
			SeriesID: seriesID,
			Limit:    int32(pagination.PageSize),
			Offset:   int32(offset),
		}
		return h.s.Queries.ListEpisodesWithAssetsBySeriesPaginated(ctx, params)
	}

	itemCount, dbEpisodes, err := util.FetchPaginatedData(ctx, fetchCount, fetchEpisodes)
	if err != nil {
		response.HandleDBError(ctx, w, err, "We couldn't retrieve the episodes.")
		return
	}

	episodesData := mapping.EpisodesWithAssets(dbEpisodes)

	paginationMeta := util.CalculatePaginationResponse(pagination.Page, pagination.PageSize, itemCount)
	res := util.PaginatedResponse[v1.EpisodeResponse]{
		Data:       episodesData,
		Pagination: paginationMeta,
	}

	response.RespondWithJSON(ctx, w, http.StatusOK, res)
}

// getSeriesEpisode godoc
// @Summary      Get episode by ID
// @Description  Get a single episode by its ID
// @Tags         Episodes
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Episode ID"
// @Success      200  {object}  v1.EpisodeResponse
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /series/episodes/{id} [get]
func (h *Handler) getSeriesEpisode(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idParam := chi.URLParam(r, "id")
	if idParam == "" {
		response.RespondWithError(ctx, w, http.StatusBadRequest, "Episode ID is required.")
		return
	}

	episodeID, err := uuid.Parse(idParam)
	if err != nil {
		response.RespondWithError(ctx, w, http.StatusBadRequest, "Invalid episode ID format.")
		return
	}

	dbEpisode, err := h.s.Queries.GetEpisode(ctx, episodeID)
	if err != nil {
		response.HandleDBError(ctx, w, err, "Episode not found.")
		return
	}

	assets, err := h.s.Queries.ListAssetsByEpisode(ctx, episodeID)
	if err != nil {
		response.HandleDBError(ctx, w, err, "We couldn't retrieve the episode assets.")
		return
	}

	res := mapping.Episode(dbEpisode, assets)
	response.RespondWithJSON(ctx, w, http.StatusOK, res)
}

// postSeriesEpisode godoc
// @Summary      Create a new episode
// @Description  Create a new episode for a series
// @Tags         Episodes
// @Accept       json
// @Produce      json
// @Param        episode  body      v1.CreateEpisodeRequest  true  "Episode data"
// @Success      201      {object}  v1.EpisodeResponse
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /series/episodes [post]
func (h *Handler) postSeriesEpisode(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req, err := validation.DecodeAndValidate[v1.CreateEpisodeRequest](r, h.v)
	if err != nil {
		response.RespondWithError(ctx, w, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	seriesID, err := uuid.Parse(req.SeriesID)
	if err != nil {
		response.RespondWithError(ctx, w, http.StatusBadRequest, "Invalid series ID format.")
		return
	}

	params := sqlc.CreateEpisodeParams{
		SeriesID:        seriesID,
		Title:           req.Title,
		Description:     req.Description,
		DurationSeconds: req.DurationSeconds,
		PublishDate:     req.PublishDate,
	}

	dbEpisode, err := h.s.Queries.CreateEpisode(ctx, params)
	if err != nil {
		response.HandleDBError(ctx, w, err, "We couldn't create the episode.")
		return
	}

	if err := h.q.EnqueueIndexEpisode(ctx, dbEpisode, []sqlc.EpisodeAsset{}); err != nil {
		slog.ErrorContext(ctx, "failed to enqueue index episode task", "err", err, "episode_id", dbEpisode.ID)
	}

	assets, err := h.s.Queries.ListAssetsByEpisode(ctx, dbEpisode.ID)
	if err != nil {
		response.HandleDBError(ctx, w, err, "We couldn't retrieve the episode assets.")
		return
	}

	res := mapping.Episode(dbEpisode, assets)
	response.RespondWithJSON(ctx, w, http.StatusCreated, res)
}

// putSeriesEpisode godoc
// @Summary      Update episode by ID
// @Description  Update an existing episode with the provided data
// @Tags         Episodes
// @Accept       json
// @Produce      json
// @Param        id       path      string                   true  "Episode ID"
// @Param        episode  body      v1.UpdateEpisodeRequest  true  "Episode data"
// @Success      200      {object}  v1.EpisodeResponse
// @Failure      400      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /series/episodes/{id} [put]
func (h *Handler) putSeriesEpisode(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idParam := chi.URLParam(r, "id")
	if idParam == "" {
		response.RespondWithError(ctx, w, http.StatusBadRequest, "Episode ID is required.")
		return
	}

	episodeID, err := uuid.Parse(idParam)
	if err != nil {
		response.RespondWithError(ctx, w, http.StatusBadRequest, "Invalid episode ID format.")
		return
	}

	req, err := validation.DecodeAndValidate[v1.UpdateEpisodeRequest](r, h.v)
	if err != nil {
		response.RespondWithError(ctx, w, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	params := sqlc.UpdateEpisodeParams{
		ID:              episodeID,
		Title:           req.Title,
		Description:     req.Description,
		DurationSeconds: req.DurationSeconds,
		PublishDate:     req.PublishDate,
	}

	dbEpisode, err := h.s.Queries.UpdateEpisode(ctx, params)
	if err != nil {
		response.HandleDBError(ctx, w, err, "Episode not found.")
		return
	}

	assets, err := h.s.Queries.ListAssetsByEpisode(ctx, episodeID)
	if err != nil {
		response.HandleDBError(ctx, w, err, "We couldn't retrieve the episode assets.")
		return
	}

	if err := h.q.EnqueueIndexEpisode(ctx, dbEpisode, assets); err != nil {
		slog.ErrorContext(ctx, "failed to enqueue index episode task", "err", err, "episode_id", dbEpisode.ID)
	}

	res := mapping.Episode(dbEpisode, assets)
	response.RespondWithJSON(ctx, w, http.StatusOK, res)
}

// deleteSeriesEpisode godoc
// @Summary      Delete episode by ID
// @Description  Soft delete an episode by its ID
// @Tags         Episodes
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Episode ID"
// @Success      204  "No Content"
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /series/episodes/{id} [delete]
func (h *Handler) deleteSeriesEpisode(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idParam := chi.URLParam(r, "id")
	if idParam == "" {
		response.RespondWithError(ctx, w, http.StatusBadRequest, "Episode ID is required.")
		return
	}

	episodeID, err := uuid.Parse(idParam)
	if err != nil {
		response.RespondWithError(ctx, w, http.StatusBadRequest, "Invalid episode ID format.")
		return
	}

	err = h.s.Queries.DeleteEpisode(ctx, episodeID)
	if err != nil {
		response.HandleDBError(ctx, w, err, "We couldn't delete the episode.")
		return
	}

	if err := h.q.EnqueueDeleteEpisode(ctx, episodeID.String()); err != nil {
		slog.ErrorContext(ctx, "failed to enqueue delete episode task", "err", err, "episode_id", episodeID)
	}

	w.WriteHeader(http.StatusNoContent)
}
