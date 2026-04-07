package save

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	resp "github.com/git-217/golang-api/internal/lib/api/response"
	"github.com/git-217/golang-api/internal/lib/logger/sl"
	"github.com/git-217/golang-api/internal/lib/random"
	"github.com/git-217/golang-api/internal/storage"
	"strings"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator"
)

type Request struct {
	URL   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty"`
}

type Response struct {
	resp.Response
	Alias string `json:"alias,omitempty"`
}

const (
	aliasLength    = 5
	maxGenAttempts = 5
)

type URLSaver interface {
	SaveURL(ctx context.Context, urlToSave string, alias string) (int, error)
}

func New(log *slog.Logger, urlSaver URLSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.save.New"
		lg := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			lg.Error("failed to decode request body", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to decode request"))

			return
		}

		if !strings.HasPrefix(req.URL, "https://") {
			lg.Info("URL without protocol")

			render.JSON(w, r, resp.Error("URL must starts with 'https://'"))

			return
		}

		lg.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)
			lg.Error("Invalid request", sl.Err(err))

			render.JSON(w, r, resp.ValidationError(validateErr))
			return
		}

		alias := req.Alias

		for range maxGenAttempts {
			if alias == "" {
				alias = random.NewRandomString(aliasLength)
			}

			id, err := urlSaver.SaveURL(r.Context(), req.URL, alias)

			if err == nil {
				lg.Info("url added", slog.Int("id", id))
				responseOK(w, r, alias)
				return
			}

			if errors.Is(err, storage.ErrAliasExists) {
				if req.Alias != "" {
					lg.Info("alias already exists", sl.Err(err))

					render.JSON(w, r, resp.Error("alias already exists"))
					return
				}

				alias = ""
				continue
			}

			lg.Error("failed to save URL", sl.Err(err))
			render.JSON(w, r, resp.Error("failed to save URL"))
			return
		}
		lg.Error("max attempts reached for generating unique alias")
		render.JSON(w, r, resp.Error("failed to generate unique alias"))

	}
}

func responseOK(w http.ResponseWriter, r *http.Request, alias string) {
	render.JSON(w, r, Response{
		Response: resp.OK(),
		Alias:    alias,
	})
}
