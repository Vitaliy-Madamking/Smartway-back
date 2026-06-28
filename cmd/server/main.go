package main

import (
	"context"
	"log/slog"
	stdhttp "net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	deliveryhttp "hotel-matcher/internal/delivery/http"
	"hotel-matcher/internal/infrastructure/config"
	"hotel-matcher/internal/infrastructure/logger"
	"hotel-matcher/internal/repository/memory"
	"hotel-matcher/internal/usecase"
)

func main() {
	cfg := config.Load()
	log := logger.NewLogger()

	hotelRepo := memory.NewHotelRepository()
	matcher := usecase.NewMatcher(hotelRepo)
	handler := deliveryhttp.NewHandler(matcher, log)

	mux := stdhttp.NewServeMux()
	mux.HandleFunc("/api/match", handler.MatchHandler)
	mux.HandleFunc("/api/upload", handler.UploadHandler)
	mux.HandleFunc("/health", func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(stdhttp.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	stack := deliveryhttp.RecoveryMiddleware(
		deliveryhttp.LoggingMiddleware(
			deliveryhttp.CORSMiddleware(mux),
			slog.Default(),
		),
		slog.Default(),
	)

	server := &stdhttp.Server{
		Addr:         ":" + cfg.Port,
		Handler:      stack,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info("starting server", "port", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != stdhttp.ErrServerClosed {
			log.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Error("shutdown error", "error", err)
	}
	log.Info("server stopped")
}