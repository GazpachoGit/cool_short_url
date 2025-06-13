package main

import (
	"fmt"
	"log/slog"
	"os"
	"short-url/internal/config"
	"short-url/internal/lib/sl"
	"short-url/internal/storage/sqlite"
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
	fmt.Println(storage.DeleteURL("3"))

	//router chi

	//server

}
