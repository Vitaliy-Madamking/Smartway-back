package main

import (
	"context"          // Для управления временем жизни (graceful shutdown)
	"log/slog"         // Структурированное логирование
	stdhttp "net/http" // HTTP-сервер (алиас чтобы не конфликтовать с нашим пакетом)
	"os"               // Работа с ОС (сигналы, выход)
	"os/signal"        // Перехват системных сигналов
	"syscall"          // Системные вызовы (SIGINT, SIGTERM)
	"time"             // Таймауты и задержки

	deliveryhttp "hotel-matcher/internal/delivery/http" // HTTP-слой (алиас/альтернатива)
	"hotel-matcher/internal/infrastructure/config"      // Конфигурация
	"hotel-matcher/internal/infrastructure/logger"      // Логгер
	"hotel-matcher/internal/repository/memory"          // In-memory хранилище
	"hotel-matcher/internal/usecase"                    // Бизнес-логика
)

func main() {
	//Загрузка конфигурации (порт из .env или переменных окружения)
	cfg := config.Load()

	//Инициализация логгера (структурированные JSON-логи)
	log := logger.NewLogger()

	// Внедрение зависимостей (Dependency Injection)
	//Создаём репозиторий (хранилище в памяти)
	hotelRepo := memory.NewHotelRepository()

	// Вариант 1: Классический матчер 
	// matcher := usecase.NewMatcher(hotelRepo)

	// Вариант 2: Универсальный матчер 
	matcher := usecase.NewUniversalMatcher(hotelRepo)



	//Создаём HTTP-обработчик с матчером и логгером
	handler := deliveryhttp.NewHandler(matcher, log)

	//Настройка маршрутов (роутинг)
	mux := stdhttp.NewServeMux()
	// POST /api/match - JSON-запрос с отелями
	mux.HandleFunc("/api/match", handler.MatchHandler)
	// POST /api/upload - загрузка CSV-файла
	mux.HandleFunc("/api/upload", handler.UploadHandler)
	// GET /health - проверка работоспособности
	mux.HandleFunc("/health", func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(stdhttp.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	//Применение middleware (обёртки)
	// Порядок важен: Recovery → Logging → CORS
	stack := deliveryhttp.RecoveryMiddleware( // Восстановление после паники
		deliveryhttp.LoggingMiddleware( // Логирование запросов
			deliveryhttp.CORSMiddleware(mux), // CORS (для фронта)
			slog.Default(),
		),
		slog.Default(),
	)

	//Создание HTTP-сервера
	server := &stdhttp.Server{
		Addr:         ":" + cfg.Port,   // Порт из конфига
		Handler:      stack,            // Роутер с middleware
		ReadTimeout:  10 * time.Second, // Макс. время чтения запроса
		WriteTimeout: 10 * time.Second, // Макс. время записи ответа
		IdleTimeout:  60 * time.Second, // Макс. время простоя соединения
	}

	//Запуск сервера в отдельной горутине
	go func() {
		log.Info("starting server", "port", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != stdhttp.ErrServerClosed {
			log.Error("server failed", "error", err)
			os.Exit(1) // Критическая ошибка - завершаем программу
		}
	}()

	//Graceful Shutdown (плавное завершение)
	// Ждём сигналы: Ctrl+C (SIGINT) или завершение процесса (SIGTERM)
	quit := make(chan os.Signal, 1) // Канал для сигналов
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit // Блокируемся до получения сигнала

	log.Info("shutting down...")

	// Даём 30 секунд на завершение текущих запросов
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Останавливаем сервер
	if err := server.Shutdown(ctx); err != nil {
		log.Error("shutdown error", "error", err)
	}

	log.Info("server stopped")
}