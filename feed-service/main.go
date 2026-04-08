package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	dbpkg "bsky-schwartz/db"
)

func main() {
	// Logger strutturato
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	// Inizializza database
	if err := dbpkg.Init("../data.db"); err != nil {
		logger.Error("failed to initialize database", "err", err)
		os.Exit(1)
	}
	defer dbpkg.Close()

	// Carica configurazione
	cfg, err := LoadConfig()
	if err != nil {
		logger.Error("failed to load config", "err", err)
		os.Exit(1)
	}

	// Contesto per shutdown graceful
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	handler := NewServer(cfg, logger)
	addr := fmt.Sprintf("%s:%d", cfg.ListenHost, cfg.Port)
	srv := &http.Server{Addr: addr, Handler: handler}

	// Goroutine per shutdown graceful
	go func() {
		<-ctx.Done()
		logger.Info("shutting down")

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()

		srv.SetKeepAlivesEnabled(false)
		srv.Shutdown(shutdownCtx)
	}()

	// Avvia server
	logger.Info("running feed generator server", "addr", fmt.Sprintf("http://%s", addr))
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		logger.Error("server error", "err", err)
		os.Exit(1)
	}
}
