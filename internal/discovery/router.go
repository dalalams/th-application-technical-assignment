package discovery

import (
	"context"
	"th-application-technical-assignment/pkg/search"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

type Handler struct {
	v            *validator.Validate
	searchClient search.Searcher
}

func Routes(ctx context.Context, h *Handler) chi.Router {
	r := chi.NewRouter()

	r.Route("/v1", func(r chi.Router) {
		r.Get("/search/series", h.searchSeries)
		r.Get("/search/episodes", h.searchEpisodes)
	})
	return r
}
