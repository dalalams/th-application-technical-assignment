package response

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
)

func RespondWithError(ctx context.Context, w http.ResponseWriter, code int, message string) {
	payload := map[string]string{"error": message}
	RespondWithJSON(ctx, w, code, payload)
}

func RespondWithJSON(ctx context.Context, w http.ResponseWriter, code int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		slog.ErrorContext(ctx, "failed to encode response", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
		return
	}
}

