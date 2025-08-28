package cms

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"th-application-technical-assignment/internal/response"
	v1 "th-application-technical-assignment/pkg/api/cms/v1"
	"th-application-technical-assignment/pkg/tasks"
	"th-application-technical-assignment/pkg/validation"

	"github.com/google/uuid"
)

// postImportContent godoc
// @Summary      Import content from external source
// @Description  Import content from YouTube, and other sources
// @Tags         Import
// @Accept       json
// @Produce      json
// @Param        import  body      v1.ImportRequest  true  "Import request data"
// @Success      204  "No Content"
// @Failure      400     {object}  map[string]string
// @Failure      500     {object}  map[string]string
// @Router       /import [post]
func (h *Handler) postImportContent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req, err := validation.DecodeAndValidate[v1.ImportRequest](r, h.v)
	if err != nil {
		response.RespondWithError(ctx, w, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	seriesID, err := uuid.Parse(req.SeriesID)
	if err != nil {
		response.RespondWithError(ctx, w, http.StatusBadRequest, "Invalid series ID format.")
		return
	}

	_, err = h.s.Queries.GetSeries(ctx, seriesID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			response.RespondWithError(ctx, w, http.StatusNotFound, "Series not found.")
			return
		}
		slog.ErrorContext(ctx, "failed to get series", "err", err, "series_id", seriesID)
		response.RespondWithError(ctx, w, http.StatusInternalServerError, "Could not verify series.")
		return
	}

	p := tasks.ImportContentPayload{
		SourceType: req.SourceType,
		SourceURL:  req.SourceURL,
		SeriesID:   seriesID.String(),
	}

	if err := h.q.EnqueueImportContent(ctx, p); err != nil {
		slog.ErrorContext(ctx, "failed to enqueue import task", "err", err, "series_id", seriesID)
		response.RespondWithError(ctx, w, http.StatusInternalServerError, "Could not queue import task.")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
