package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"hotel-matcher/internal/domain"
	"hotel-matcher/internal/infrastructure/logger"
	"hotel-matcher/internal/usecase"
)

type Handler struct {
	matcher usecase.Matcher
	logger  logger.Logger
}

func NewHandler(matcher usecase.Matcher, log logger.Logger) *Handler {
	return &Handler{matcher: matcher, logger: log}
}

func (h *Handler) MatchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req MatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode request", "error", err)
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	hotels, cfg := req.ToDomain()
	result, err := h.matcher.Match(r.Context(), hotels, cfg)
	if err != nil {
		h.logger.Error("matching failed", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	response := ToDTO(result)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}

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

	if !strings.HasSuffix(header.Filename, ".csv") && !strings.HasSuffix(header.Filename, ".CSV") {
		http.Error(w, "only CSV files are allowed", http.StatusBadRequest)
		return
	}

	hotels, err := parseCSV(file)
	if err != nil {
		h.logger.Error("failed to parse CSV", "error", err)
		http.Error(w, "invalid CSV format", http.StatusBadRequest)
		return
	}
	if len(hotels) == 0 {
		http.Error(w, "no hotels found in CSV", http.StatusBadRequest)
		return
	}
	h.logger.Info("CSV uploaded", "hotels", len(hotels), "filename", header.Filename)

	cfg := domain.DefaultConfig()
	if thresholdStr := r.FormValue("threshold"); thresholdStr != "" {
		if th, err := strconv.ParseFloat(thresholdStr, 64); err == nil && th >= 0 && th <= 1 {
			cfg.Threshold = th
		}
	}
	// Поддержка выбора алгоритма через form-data
	if alg := r.FormValue("algorithm"); alg != "" {
		cfg.Algorithm = alg
	}

	result, err := h.matcher.Match(r.Context(), hotels, cfg)
	if err != nil {
		h.logger.Error("matching failed", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := ToDTO(result)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}