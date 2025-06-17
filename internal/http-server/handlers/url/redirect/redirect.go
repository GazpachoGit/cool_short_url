package redirect

import (
	"errors"
	"log/slog"
	"net/http"

	responseModel "short-url/internal/http-server/model/response"
	"short-url/internal/lib/sl"
	"short-url/internal/storage"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type URLGetter interface {
	GetURL(alias string) (string, error)
}

func New(log *slog.Logger, urlGetter URLGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.redirect.new"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		alias := chi.URLParam(r, "alias")
		if alias == "" {
			log.Info("alias url param is empty")
			render.JSON(w, r, responseModel.Error("invalid request"))
			return
		}

		url, err := urlGetter.GetURL(alias)
		if err != nil {
			if errors.Is(err, storage.ErrURLNotFound) {
				log.Info("url not found", slog.String("alias", alias))
				render.JSON(w, r, responseModel.Error("url not found"))
			} else {
				log.Error("failed to get url", sl.Err(err))
				render.JSON(w, r, responseModel.Error("failed to get url"))
			}
			return
		}

		log.Info("url found", slog.String("url", url))
		http.Redirect(w, r, url, http.StatusFound)
	}

}
