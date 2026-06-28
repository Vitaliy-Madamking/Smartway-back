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
	"hotel-matcher/internal/infrastructure/postgres"
	repopg "hotel-matcher/internal/repository/memory"
	"hotel-matcher/internal/usecase"
)

func main() {
	cfg := config.Load()
	log := logger.NewLogger()

	// Подключение к PostgreSQL
	ctx := context.Background()
	db, err := postgres.New(ctx, postgres.Config{
		Host:     cfg.DBHost,
		Port:     cfg.DBPort,
		User:     cfg.DBUser,
		Password: cfg.DBPassword,
		DBName:   cfg.DBName,
		SSLMode:  cfg.DBSSLMode,
	})
	if err != nil {
		log.Error("failed to connect to postgres", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	log.Info("connected to postgres", "host", cfg.DBHost, "db", cfg.DBName)

	// Репозитории
	hotelRepo := repopg.NewHotelRepository(db)
	groupRepo := repopg.NewGroupRepository(db)

	// Use-case
	matcher := usecase.NewMatcher(hotelRepo)

	// HTTP-хендлер
	handler := deliveryhttp.NewHandler(matcher, hotelRepo, groupRepo, log)

	mux := stdhttp.NewServeMux()

	// --- Маршруты ---
	mux.HandleFunc("/api/upload", handler.UploadHandler)        // POST — загрузка CSV + матчинг
	mux.HandleFunc("/api/hotels", handler.GetHotelsHandler)     // GET  — все отели
	mux.HandleFunc("/api/groups", handler.GetGroupsHandler)     // GET  — все группы
	mux.HandleFunc("/api/groups/", handler.GetGroupByIDHandler) // GET  — группа по id

	mux.HandleFunc("/health", func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(stdhttp.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
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
		WriteTimeout: 30 * time.Second, // увеличили для больших CSV
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
	shutCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := server.Shutdown(shutCtx); err != nil {
		log.Error("shutdown error", "error", err)
	}
	log.Info("server stopped")
}
