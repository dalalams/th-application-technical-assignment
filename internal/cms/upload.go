package cms

import (
	"log/slog"
	"net/http"
	"th-application-technical-assignment/internal/response"
	v1 "th-application-technical-assignment/pkg/api/cms/v1"
	"th-application-technical-assignment/pkg/mapping"
	"th-application-technical-assignment/pkg/validation"
	"th-application-technical-assignment/sqlc"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// getEpisodeUploadURL godoc
// @Summary      Get a pre-signed URL for an episode media upload
// @Description  Validates the episode ID and returns a temporary URL for the client to upload a file directly to S3.
// @Tags         Episodes
// @Produce      json
// @Param        id        path      string  true  "Episode ID"
// @Param        request   body      v1.UploadURLRequest       true  "Upload request details"
// @Success      200       {object}  v1.UploadURLResponse "Successfully generated the pre-signed URL"
// @Failure      400       {object}  map[string]string "Bad Request (e.g., invalid ID, missing filename)"
// @Failure      404       {object}  map[string]string "Episode not found"
// @Failure      500       {object}  map[string]string "Internal Server Error"
// @Router       /series/episodes/{id}/upload-url [post]
func (h *Handler) getEpisodeUploadURL(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idParam := chi.URLParam(r, "id")
	if idParam == "" {
		response.RespondWithError(ctx, w, http.StatusBadRequest, "Episode Id required")
	}

	episodeID, err := uuid.Parse(idParam)
	if err != nil {
		response.RespondWithError(ctx, w, http.StatusBadRequest, "Invalid episode ID format.")
		return
	}

	req, err := validation.DecodeAndValidate[v1.UploadURLRequest](r, h.v)
	if err != nil {
		response.RespondWithError(ctx, w, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	ep, err := h.s.Queries.GetEpisode(ctx, episodeID)
	if err != nil {
		response.HandleDBError(ctx, w, err, "Episode not found.")
		return
	}

	key := h.mc.GenerateKey(ep.SeriesID, ep.ID, req.Filename)

	presignedURL, err := h.mc.GeneratePresignedPutURL(ctx, key, 20*time.Minute)
	if err != nil {
		slog.ErrorContext(ctx, "failed to generate presigned URL", "err", err)
		response.RespondWithError(ctx, w, http.StatusInternalServerError, "Could not generate upload URL.")
		return
	}

	res := v1.UploadURLResponse{
		UploadURL: presignedURL.String(),
		S3Key:     key,
		S3Bucket:  h.mc.GetBucketName(),
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}

	slog.InfoContext(ctx, "generated upload URL for episode",
		"episode_id", episodeID,
		"s3_key", key,
	)

	response.RespondWithJSON(ctx, w, http.StatusOK, res)
}

// confirmEpisodeUpload godoc
// @Summary      Confirm episode file upload
// @Description  Confirm that the file was successfully uploaded and update episode metadata
// @Tags         Episodes
// @Accept       json
// @Produce      json
// @Param        id        path      string                      true  "Episode ID"
// @Param        request   body      v1.ConfirmUploadRequest     true  "Upload confirmation details"
// @Success      200       {object}  v1.EpisodeResponse
// @Failure      400       {object}  map[string]string
// @Failure      404       {object}  map[string]string
// @Failure      500       {object}  map[string]string
// @Router       /series/episodes/{id}/upload-confirm [post]
func (h *Handler) confirmEpisodeUpload(w http.ResponseWriter, r *http.Request) {
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

	req, err := validation.DecodeAndValidate[v1.ConfirmUploadRequest](r, h.v)
	if err != nil {
		response.RespondWithError(ctx, w, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	assetParams := sqlc.CreateAssetParams{
		EpisodeID: episodeID,
		AssetType: req.AssetType,
		MimeType:  req.MimeType,
		SizeBytes: &req.Size,
		Url:       &req.S3Key,
	}

	_, err = h.s.Queries.CreateAsset(ctx, assetParams)
	if err != nil {
		response.HandleDBError(ctx, w, err, "Failed to confirm upload.")
		return
	}

	episode, err := h.s.Queries.GetEpisode(ctx, episodeID)
	if err != nil {
		response.HandleDBError(ctx, w, err, "Episode not found.")
		return
	}

	assets, err := h.s.Queries.ListAssetsByEpisode(ctx, episodeID)
	if err != nil {
		response.HandleDBError(ctx, w, err, "We couldn't retrieve the episode assets.")
		return
	}

	res := mapping.Episode(episode, assets)

	slog.InfoContext(ctx, "episode upload confirmed",
		"episode_id", episodeID,
		"s3_key", req.S3Key,
		"size", req.Size,
	)

	response.RespondWithJSON(ctx, w, http.StatusOK, res)
}
