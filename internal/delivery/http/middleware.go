package http

import (
	"log/slog"      // структурированное логирование
	"net/http"      // HTTP-сервер
	"runtime/debug" // получение стека вызовов при панике
	"time"          // измерение времени
)

// CORSMiddleware — middleware для CORS (Cross-Origin Resource Sharing)
// Разрешает кросс-доменные запросы с фронтенда
// Возвращает http.Handler с добавленными CORS-заголовками
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Разрешаем запросы с любых источников (для разработки)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		// Разрешённые HTTP-методы
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		// Разрешённые заголовки
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Preflight-запрос (OPTIONS) — возвращаем 200 OK без обработки
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Передаём управление следующему обработчику
		next.ServeHTTP(w, r)
	})
}

// LoggingMiddleware — middleware для логирования HTTP-запросов
// Логирует: метод, путь, длительность выполнения
// Возвращает http.Handler с добавленным логированием
func LoggingMiddleware(next http.Handler, log *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now() // запоминаем время начала

		// Вызываем следующий обработчик
		next.ServeHTTP(w, r)

		// Логируем запрос после его выполнения
		log.Info("request",
			"method", r.Method,                          // HTTP-метод
			"path", r.URL.Path,                          // путь запроса
			"duration_ms", time.Since(start).Milliseconds(), // время выполнения в мс
		)
	})
}

// RecoveryMiddleware — middleware для восстановления после паники
// Перехватывает panic в обработчиках, логирует ошибку и стек
// Возвращает http.Handler с защитой от паники
func RecoveryMiddleware(next http.Handler, log *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Отложенная функция: выполнится при панике или выходе из функции
		defer func() {
			// Если произошла паника — recover() возвращает ошибку
			if err := recover(); err != nil {
				// Логируем ошибку и стек вызовов
				log.Error("panic recovered",
					"error", err,                      // текст паники
					"stack", string(debug.Stack()),    // полный стек вызовов
				)
				// Отвечаем 500 Internal Server Error
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}()

		// Вызываем следующий обработчик (может вызвать panic)
		next.ServeHTTP(w, r)
	})
}