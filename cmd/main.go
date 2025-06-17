package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"short-url/internal/config"
	"short-url/internal/http-server/handlers/url/redirect"
	"short-url/internal/http-server/handlers/url/save"
	mwLogger "short-url/internal/http-server/middleware"
	"short-url/internal/lib/sl"
	"short-url/internal/storage/sqlite"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

const (
	envLocal = "local"
	envDev   = "dev"
)

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger
	switch env {
	case envLocal:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envDev:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	}
	return log
}

func main() {
	//config cleanenv
	cfg := config.MustLoad()

	//TODO: remove
	fmt.Println(cfg)

	//logger slog
	log := setupLogger(cfg.Env)
	//TODO: remove
	log.Debug("Logger ready")
	log.Info("env is", slog.String("env", cfg.Env))

	//storage sqllite
	storage, err := sqlite.New(cfg.StoragePath)
	if err != nil {
		log.Error("can't connect to storage", sl.Err(err))
		os.Exit(1)
	}

	//router chi
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(mwLogger.New(log))
	router.Use(middleware.Recoverer)
	//to get params from url
	router.Use(middleware.URLFormat)

	// router.Route("/url", func(r chi.Router) {
	// 	//'short-url' - title in browser
	// 	r.Use(middleware.BasicAuth("short-url", map[string]string{
	// 		cfg.HTTPServer.User: cfg.HTTPServer.Password,
	// 	}))

	// 	r.Post("/", save.New(log, storage))
	// })

	router.Post("/url", save.New(log, storage))
	router.Get("/{alias}", redirect.New(log, storage))

	//server
	log.Info("starting server", slog.String("address", cfg.Address))
	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	if err = srv.ListenAndServe(); err != nil {
		log.Error("failed to start server")
	}
}
