package discovery

import (
	"log/slog"
	"net/http"
	"strconv"
	"th-application-technical-assignment/internal/response"
	"th-application-technical-assignment/pkg/api/discovery/v1"
	"th-application-technical-assignment/pkg/search"
)

// searchSeries godoc
// @Summary      Search series
// @Description  Search for series using full-text search
// @Tags         Discovery
// @Accept       json
// @Produce      json
// @Param        q           query     string  false  "Search query"
// @Param        page        query     int     false  "Page number (default: 1)"
// @Param        page_size   query     int     false  "Page size (default: 20, max: 100)"
// @Param        category_id query     string  false  "Filter by category ID"
// @Param        type        query     string  false  "Filter by series type"
// @Param        language    query     string  false  "Filter by language"
// @Success      200         {object}  v1.SearchResponse
// @Failure      400         {object}  map[string]string
// @Failure      500         {object}  map[string]string
// @Router       /search/series [get]
func (h *Handler) searchSeries(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req := v1.SearchSeriesRequest{
		Query:    r.URL.Query().Get("q"),
		Page:     1,
		PageSize: 20,
	}

	if page := r.URL.Query().Get("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			req.Page = p
		}
	}

	if pageSize := r.URL.Query().Get("page_size"); pageSize != "" {
		if ps, err := strconv.Atoi(pageSize); err == nil && ps > 0 && ps <= 100 {
			req.PageSize = ps
		}
	}

	if categoryID := r.URL.Query().Get("category_id"); categoryID != "" {
		req.CategoryID = &categoryID
	}

	if seriesType := r.URL.Query().Get("type"); seriesType != "" {
		req.Type = &seriesType
	}

	if language := r.URL.Query().Get("language"); language != "" {
		req.Language = &language
	}

	if err := h.v.Struct(&req); err != nil {
		response.RespondWithError(ctx, w, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	searchReq := search.SearchRequest{
		Query:    req.Query,
		Page:     req.Page,
		PageSize: req.PageSize,
		Filters:  make(map[string]interface{}),
	}

	if req.CategoryID != nil {
		searchReq.Filters["category_id"] = *req.CategoryID
	}
	if req.Type != nil {
		searchReq.Filters["type"] = *req.Type
	}
	if req.Language != nil {
		searchReq.Filters["language"] = *req.Language
	}

	searchResult, err := h.searchClient.SearchSeries(ctx, searchReq)
	if err != nil {
		slog.ErrorContext(ctx, "failed to search series", "err", err)
		response.RespondWithError(ctx, w, http.StatusInternalServerError, "Search failed.")
		return
	}

	pageCount := int((searchResult.Total + int64(req.PageSize) - 1) / int64(req.PageSize))
	res := v1.SearchResponse{
		Query:     req.Query,
		Total:     searchResult.Total,
		Page:      req.Page,
		PageSize:  req.PageSize,
		PageCount: pageCount,
		Results:   searchResult.Hits,
	}

	response.RespondWithJSON(ctx, w, http.StatusOK, res)
}

// searchEpisodes godoc
// @Summary      Search episodes
// @Description  Search for episodes using full-text search
// @Tags         Discovery
// @Accept       json
// @Produce      json
// @Param        q         query     string  false  "Search query"
// @Param        page      query     int     false  "Page number (default: 1)"
// @Param        page_size query     int     false  "Page size (default: 20, max: 100)"
// @Param        series_id query     string  false  "Filter by series ID"
// @Success      200       {object}  v1.SearchResponse
// @Failure      400       {object}  map[string]string
// @Failure      500       {object}  map[string]string
// @Router       /search/episodes [get]
func (h *Handler) searchEpisodes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req := v1.SearchEpisodesRequest{
		Query:    r.URL.Query().Get("q"),
		Page:     1,
		PageSize: 20,
	}

	if page := r.URL.Query().Get("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			req.Page = p
		}
	}

	if pageSize := r.URL.Query().Get("page_size"); pageSize != "" {
		if ps, err := strconv.Atoi(pageSize); err == nil && ps > 0 && ps <= 100 {
			req.PageSize = ps
		}
	}

	if seriesID := r.URL.Query().Get("series_id"); seriesID != "" {
		req.SeriesID = &seriesID
	}

	if err := h.v.Struct(&req); err != nil {
		response.RespondWithError(ctx, w, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	searchReq := search.SearchRequest{
		Query:    req.Query,
		Page:     req.Page,
		PageSize: req.PageSize,
		Filters:  make(map[string]interface{}),
	}

	if req.SeriesID != nil {
		searchReq.Filters["series_id"] = *req.SeriesID
	}

	searchResult, err := h.searchClient.SearchEpisodes(ctx, searchReq)
	if err != nil {
		slog.ErrorContext(ctx, "failed to search episodes", "err", err)
		response.RespondWithError(ctx, w, http.StatusInternalServerError, "Search failed.")
		return
	}

	pageCount := int((searchResult.Total + int64(req.PageSize) - 1) / int64(req.PageSize))
	res := v1.SearchResponse{
		Query:     req.Query,
		Total:     searchResult.Total,
		Page:      req.Page,
		PageSize:  req.PageSize,
		PageCount: pageCount,
		Results:   searchResult.Hits,
	}

	response.RespondWithJSON(ctx, w, http.StatusOK, res)
}
