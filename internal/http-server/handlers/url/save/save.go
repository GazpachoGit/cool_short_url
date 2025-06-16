package save

import (
	"errors"
	"log/slog"
	"net/http"
	responseModel "short-url/internal/http-server/model/response"
	"short-url/internal/lib/random"
	"short-url/internal/lib/sl"
	"short-url/internal/storage"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

const aliasLength = 6

type Request struct {
	URL   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty"`
}

type Response struct {
	responseModel.Response
	Alias string `json:"alias,omitempty"`
}

//go:generate mockery --name=URLSaver
type URLSaver interface {
	SaveURL(urlToSave string, alias string) (int64, error)
}

func New(log *slog.Logger, urlSaver URLSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.save.new"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("can't decode request body", sl.Err(err))
			render.JSON(w, r, responseModel.Error("can't decode request body"))
			return
		}
		log.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			validErrs := err.(validator.ValidationErrors)

			log.Error("invalid request body", sl.Err(err))

			render.JSON(w, r, responseModel.ValidationError(validErrs))
			return
		}

		alias := req.Alias
		if alias == "" {
			alias = random.NewRandomString(aliasLength)
		}

		id, err := urlSaver.SaveURL(req.URL, req.Alias)
		if errors.Is(err, storage.ErrURLExists) {
			log.Info("url already exists", slog.String("url", req.URL))
			render.JSON(w, r, responseModel.Error("url already exists"))
			return
		}
		if err != nil {
			log.Error("failed to add url", sl.Err(err))
			render.JSON(w, r, responseModel.Error("ailed to add url"))
		}
		log.Info("id is added", slog.Int64("id", id))

		ResponseOK(w, r, alias)
	}
}

func ResponseOK(w http.ResponseWriter, r *http.Request, alias string) {
	render.JSON(w, r, Response{
		Response: responseModel.OK(),
		Alias:    alias,
	})
}
