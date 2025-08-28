package cms

import (
	"context"
	"th-application-technical-assignment/pkg/database"
	"th-application-technical-assignment/pkg/storage"
	"th-application-technical-assignment/pkg/tasks"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	mw "th-application-technical-assignment/internal/middleware"
	"github.com/go-chi/jwtauth/v5"
	"github.com/go-playground/validator/v10"
)

type Handler struct {
	s   *database.Store
	v   *validator.Validate
	q   tasks.TaskQueue
	mc  storage.ObjectStorage
	jwt *jwtauth.JWTAuth
}

func Routes(ctx context.Context, h *Handler) chi.Router {
	r := chi.NewRouter()

	r.Route("/v1", func(r chi.Router) {
		r.Use(middleware.AllowContentType("application/json"))
		r.Use(middleware.CleanPath)

		// r.Use(jwtauth.Verifier(jwt))
		// r.Use(jwtauth.Authenticator(jwt))

		r.With(mw.PaginationCtx(h.v)).Get("/series", h.listSeries)
		r.Get("/series/{id}", h.getSeries)
		r.Post("/series", h.postSeries)
		r.Put("/series/{id}", h.putSeries)
		r.Delete("/series/{id}", h.deleteSeries)

		r.With(mw.PaginationCtx(h.v)).Get("/series/episodes", h.listSeriesEpisodes)
		r.Get("/series/episodes/{id}", h.getSeriesEpisode)
		r.Post("/series/episodes", h.postSeriesEpisode)
		r.Put("/series/episodes/{id}", h.putSeriesEpisode)
		r.Delete("/series/episodes/{id}", h.deleteSeriesEpisode)

		r.With(mw.PaginationCtx(h.v)).Get("/categories", h.listCategories)
		r.Get("/categories/{id}", h.getCategory)
		r.Post("/categories", h.postCategory)
		r.Put("/categories/{id}", h.putCategory)
		r.Delete("/categories/{id}", h.deleteCategory)

		r.Post("/import", h.postImportContent)

		r.Post("/series/episodes/{id}/upload-url", h.getEpisodeUploadURL)
		r.Post("/series/episodes/{id}/upload-confirm", h.confirmEpisodeUpload)
	})
	return r
}
