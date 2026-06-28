package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"hotel-matcher/internal/domain"
	"hotel-matcher/internal/infrastructure/logger"
	"hotel-matcher/internal/usecase"
)

type Handler struct {
	matcher   usecase.Matcher
	hotelRepo usecase.HotelReader
	groupRepo usecase.GroupReader
	logger    logger.Logger
}

func NewHandler(
	matcher usecase.Matcher,
	hotelRepo usecase.HotelReader,
	groupRepo usecase.GroupReader,
	log logger.Logger,
) *Handler {
	return &Handler{
		matcher:   matcher,
		hotelRepo: hotelRepo,
		groupRepo: groupRepo,
		logger:    log,
	}
}

// UploadHandler POST /api/upload
// Принимает CSV-файл, сохраняет отели в БД, запускает матчинг, сохраняет группы
func (h *Handler) UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		h.logger.Error("failed to parse multipart form", "error", err)
		http.Error(w, "failed to parse form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		h.logger.Error("failed to get file", "error", err)
		http.Error(w, "file not provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	if !strings.HasSuffix(strings.ToLower(header.Filename), ".csv") {
		http.Error(w, "only CSV files are allowed", http.StatusBadRequest)
		return
	}

	hotels, err := parseCSV(file)
	if err != nil {
		h.logger.Error("failed to parse CSV", "error", err)
		http.Error(w, "invalid CSV format: "+err.Error(), http.StatusBadRequest)
		return
	}
	if len(hotels) == 0 {
		http.Error(w, "no hotels found in CSV", http.StatusBadRequest)
		return
	}

	h.logger.Info("CSV uploaded", "hotels", len(hotels), "filename", header.Filename)

	// Конфиг матчинга из form-параметров
	cfg := domain.DefaultConfig()
	if th := r.FormValue("threshold"); th != "" {
		if val, err := strconv.ParseFloat(th, 64); err == nil && val >= 0 && val <= 1 {
			cfg.Threshold = val
		}
	}
	if alg := r.FormValue("algorithm"); alg != "" {
		cfg.Algorithm = alg
	}

	// Шаг 1: сохраняем отели в БД, получаем обратно с ID
	savedHotels, err := h.hotelRepo.SaveBatch(r.Context(), hotels)
	if err != nil {
		h.logger.Error("failed to save hotels", "error", err)
		http.Error(w, "failed to save hotels", http.StatusInternalServerError)
		return
	}

	// Шаг 2: матчинг
	result, err := h.matcher.Match(r.Context(), savedHotels, cfg)
	if err != nil {
		h.logger.Error("matching failed", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Шаг 3: сохраняем группы в БД
	if err := h.groupRepo.SaveResult(r.Context(), result); err != nil {
		h.logger.Error("failed to save groups", "error", err)
		http.Error(w, "failed to save groups", http.StatusInternalServerError)
		return
	}

	h.logger.Info("matching completed",
		"groups", len(result.Groups),
		"unmatched", len(result.Unmatched),
	)

	writeJSON(w, http.StatusOK, ToDTO(result))
}

// GetHotelsHandler GET /api/hotels
func (h *Handler) GetHotelsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	hotels, err := h.hotelRepo.GetAll(r.Context())
	if err != nil {
		h.logger.Error("failed to get hotels", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, hotels)
}

// GetGroupsHandler GET /api/groups
func (h *Handler) GetGroupsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	groups, err := h.groupRepo.GetAll(r.Context())
	if err != nil {
		h.logger.Error("failed to get groups", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, groups)
}

// GetGroupByIDHandler GET /api/groups/{id}
func (h *Handler) GetGroupByIDHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Вытаскиваем {id} из пути: /api/groups/123
	path := strings.TrimPrefix(r.URL.Path, "/api/groups/")
	if path == "" {
		http.Error(w, "group id is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(path, 10, 64)
	if err != nil {
		http.Error(w, "invalid group id", http.StatusBadRequest)
		return
	}

	group, err := h.groupRepo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrGroupNotFound) {
			http.Error(w, "group not found", http.StatusNotFound)
			return
		}
		h.logger.Error("failed to get group", "id", id, "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, group)
}

// writeJSON — хелпер для отправки JSON-ответа
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		// Заголовки уже отправлены, только логируем
		_ = err
	}
}
