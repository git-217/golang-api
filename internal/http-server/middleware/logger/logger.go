package logger

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

func New(log *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		lg := log.With(
			slog.String("component", "middleware/logger"),
		)
		lg.Info("middleware logger has been enabled")

		fn := func(w http.ResponseWriter, r *http.Request) {
			entry := lg.With(
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("remote_addr", r.RemoteAddr),
				slog.String("user_agent", r.UserAgent()),
				slog.String("request_id", middleware.GetReqID(r.Context())),
			)
			wrw := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			t1 := time.Now()

			defer func() {
				entry.Info("request completed",
					slog.Int("status", wrw.Status()),
					slog.Int("bytes", wrw.BytesWritten()),
					slog.Duration("duration", time.Since(t1)),
				)
			}()

			next.ServeHTTP(wrw, r)
		}

		return http.HandlerFunc(fn)
	}

}
