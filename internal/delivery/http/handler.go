package http

import (
	"encoding/json" // для работы с JSON
	"net/http"      // HTTP-сервер и клиент
	"strconv"       // конвертация строк в числа
	"strings"       // работа со строками

	"hotel-matcher/internal/domain"                // бизнес-сущности
	"hotel-matcher/internal/infrastructure/logger" // логгер
	"hotel-matcher/internal/usecase"               // бизнес-логика
)

// Handler — HTTP-обработчик, содержит зависимости для обработки запросов
type Handler struct {
	matcher usecase.Matcher // интерфейс матчинга (бизнес-логика)
	logger  logger.Logger   // интерфейс логгера
}

// NewHandler — конструктор, внедряет зависимости
func NewHandler(matcher usecase.Matcher, log logger.Logger) *Handler {
	return &Handler{matcher: matcher, logger: log}
}

// MatchHandler — обрабатывает POST /api/match (JSON-запрос)
// 1. Проверяет метод
// 2. Декодирует JSON в MatchRequest
// 3. Преобразует DTO → Domain
// 4. Вызывает матчинг
// 5. Преобразует результат → DTO
// 6. Отправляет JSON-ответ
func (h *Handler) MatchHandler(w http.ResponseWriter, r *http.Request) {
	// Проверка метода: только POST
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Декодируем JSON-запрос
	var req MatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode request", "error", err)
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	// Преобразуем DTO в доменные модели
	hotels, cfg := req.ToDomain()

	// Вызываем бизнес-логику (матчинг)
	result, err := h.matcher.Match(r.Context(), hotels, cfg)
	if err != nil {
		h.logger.Error("matching failed", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Преобразуем результат в DTO для ответа
	response := ToDTO(result)

	// Отправляем JSON-ответ
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}

// UploadHandler — обрабатывает POST /api/upload (CSV-загрузка)
// 1. Проверяет метод
// 2. Парсит multipart/form-data (макс. 10 MB)
// 3. Извлекает файл из поля "file"
// 4. Проверяет расширение .csv
// 5. Парсит CSV → []domain.Hotel
// 6. Извлекает параметры threshold и algorithm из формы
// 7. Вызывает матчинг
// 8. Отправляет JSON-ответ
func (h *Handler) UploadHandler(w http.ResponseWriter, r *http.Request) {
	// Проверка метода: только POST
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Парсим multipart/form-data (ограничение 10 MB)
	// 10 << 20 = 10 * 1024 * 1024 = 10 MB
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		h.logger.Error("failed to parse multipart form", "error", err)
		http.Error(w, "failed to parse form", http.StatusBadRequest)
		return
	}

	// Извлекаем файл из поля "file"
	file, header, err := r.FormFile("file")
	if err != nil {
		h.logger.Error("failed to get file", "error", err)
		http.Error(w, "file not provided", http.StatusBadRequest)
		return
	}
	defer file.Close() // обязательно закрываем файл

	// Проверка расширения: только .csv
	if !strings.HasSuffix(header.Filename, ".csv") && !strings.HasSuffix(header.Filename, ".CSV") {
		http.Error(w, "only CSV files are allowed", http.StatusBadRequest)
		return
	}

	// Парсим CSV-файл в []domain.Hotel
	hotels, err := parseCSV(file)
	if err != nil {
		h.logger.Error("failed to parse CSV", "error", err)
		http.Error(w, "invalid CSV format", http.StatusBadRequest)
		return
	}

	// Проверка: есть ли отели в CSV
	if len(hotels) == 0 {
		http.Error(w, "no hotels found in CSV", http.StatusBadRequest)
		return
	}

	// Логируем успешную загрузку
	h.logger.Info("CSV uploaded", "hotels", len(hotels), "filename", header.Filename)

	// Настройки по умолчанию
	cfg := domain.DefaultConfig()

	// Читаем threshold из формы (опционально)
	if thresholdStr := r.FormValue("threshold"); thresholdStr != "" {
		if th, err := strconv.ParseFloat(thresholdStr, 64); err == nil && th >= 0 && th <= 1 {
			cfg.Threshold = th
		}
	}

	// Читаем algorithm из формы (опционально)
	if alg := r.FormValue("algorithm"); alg != "" {
		cfg.Algorithm = alg
	}

	// Вызываем бизнес-логику (матчинг)
	result, err := h.matcher.Match(r.Context(), hotels, cfg)
	if err != nil {
		h.logger.Error("matching failed", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Преобразуем результат в DTO для ответа
	response := ToDTO(result)

	// Отправляем JSON-ответ
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}