package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/m42-labs/alnitak-events/internal/config"
	eventhttp "github.com/m42-labs/alnitak-events/internal/http"
	"github.com/m42-labs/alnitak-events/internal/service"
)

func main() {
	cfg := config.Load()
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("service", "alnitak-events")

	svc, err := service.New(cfg, log)
	if err != nil {
		log.Error("service_init_failed", "error", err.Error())
		os.Exit(1)
	}
	defer svc.Close()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := svc.Start(ctx); err != nil {
		log.Warn("subscriber_start_failed", "error", err.Error())
	}

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           eventhttp.Router(svc),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Info("http_started", "port", cfg.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("http_failed", "error", err.Error())
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = server.Shutdown(shutdownCtx)
}
