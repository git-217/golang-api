package redirect

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	resp "psql_crud/internal/lib/api/response"
	"psql_crud/internal/lib/logger/sl"
	"psql_crud/internal/storage"
	db_req "psql_crud/internal/storage/postgres"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type urlGetter interface {
	GetURL(ctx context.Context, alias string) (*db_req.CustomURL, error)
}

func New(logger *slog.Logger, urlRepo urlGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http-server.handlers.redirect.New"

		lg := logger.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		alias := chi.URLParam(r, "alias")
		if alias == "" {
			lg.Info("empty alias")

			render.JSON(w, r, resp.Error("Invalid request"))
			return
		}

		resUrl, err := urlRepo.GetURL(r.Context(), alias)
		if errors.Is(err, storage.ErrURLNotFound) {
			lg.Info("url not found")

			render.JSON(w, r, resp.Error("URL not found"))
			return 
		}

		if err != nil {
			lg.Error("failed to get url", sl.Err(err))

			render.JSON(w, r, resp.Error("internal error"))
			return
		}

		lg.Info("got url correctly", slog.String("url", resUrl.URL))

		http.Redirect(w, r, resUrl.URL, http.StatusFound)
	}
}
