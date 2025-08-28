package middleware

import (
	"context"
	"net/http"
	"strconv"
	"th-application-technical-assignment/pkg/util"

	"github.com/go-playground/validator/v10"
)

type contextKey string

const (
	PaginationContextKey = contextKey("pagination")
)

func PaginationCtx(v *validator.Validate) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			page := util.DefaultPage
			pageSize := util.DefaultPageSize

			if pageStr := r.URL.Query().Get("page"); pageStr != "" {
				if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
					page = p
				}
			}

			if pageSizeStr := r.URL.Query().Get("page_size"); pageSizeStr != "" {
				if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= util.MaxPageSize {
					pageSize = ps
				}
			}

			pagination := util.PaginationRequest{
				Page:     page,
				PageSize: pageSize,
			}

			if err := v.Struct(&pagination); err != nil {
				http.Error(w, "Invalid pagination parameters: "+err.Error(), http.StatusBadRequest)
				return
			}

			ctx := context.WithValue(r.Context(), PaginationContextKey, pagination)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetPagination(ctx context.Context) util.PaginationRequest {
	if pagination, ok := ctx.Value(PaginationContextKey).(util.PaginationRequest); ok {
		return pagination
	}
	return util.PaginationRequest{
		Page:     util.DefaultPage,
		PageSize: util.DefaultPageSize,
	}
}
