package response

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
)

func HandleDBError(ctx context.Context, w http.ResponseWriter, err error, message string) {
	if errors.Is(err, sql.ErrNoRows) {
		RespondWithError(ctx, w, http.StatusNotFound, message)
		return
	}
	slog.ErrorContext(ctx, "database error", "err", err)
	RespondWithError(ctx, w, http.StatusInternalServerError, "A database error occurred.")
}
