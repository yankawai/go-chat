package http

import (
	"log/slog"
	"net/http"
)

func recoverPanic(next http.Handler, logger *slog.Logger) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				logger.Error(
					"panic recovered",
					"request_id", requestIDFromContext(r.Context()),
					"method", r.Method,
					"path", r.URL.Path,
					"panic", recovered,
				)
				writeError(w, http.StatusInternalServerError, "internal_error", "internal server error")
			}
		}()

		next.ServeHTTP(w, r)
	})
}
