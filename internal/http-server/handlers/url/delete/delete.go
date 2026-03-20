package del

import (
	"context"
	"log/slog"
	"net/http"
	"psql_crud/internal/lib/api/response"
	"psql_crud/internal/lib/logger/sl"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type URLRepo interface {
	DeleteURL(ctx context.Context, alias string) error
}

// delete only one url by its alias
func One(log *slog.Logger, repo URLRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.delete.One"

		lg := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		alias := chi.URLParam(r, "alias")

		if alias == "" {
			lg.Info("empty alias")

			render.JSON(w, r, response.Error("url cannot be empty"))
			return
		}

		err := repo.DeleteURL(r.Context(), alias)
		if err != nil {
			lg.Error("cannot delete url", sl.Err(err))

			render.JSON(w, r, response.Error("internal error"))
		}

		lg.Info("url deleted successfully")

		render.JSON(w, r, response.OK())
	}
}
