package http

import (
	"hotel-matcher/internal/domain"
	"strings"
)

// ErrorResponse — JSON-ответ при ошибке
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// MatchRequest — структура входящего запроса к эндпоинту /api/match
// Содержит список отелей и опциональные настройки матчинга
type MatchRequest struct {
	Hotels   []HotelDTO `json:"hotels"`              // список отелей от поставщиков
	Config   ConfigDTO  `json:"config,omitempty"`    // настройки алгоритма (необязательно)
	Page     int        `json:"page,omitempty"`      // номер страницы (по умолчанию 1)
	Limit    int        `json:"limit,omitempty"`     // количество групп на странице (по умолчанию 50)
	Search   string     `json:"search,omitempty"`    // поиск по названию группы
	SortBy   string     `json:"sortBy,omitempty"`    // поле для сортировки: "name", "confidence", "hotelsCount"
	SortDir  string     `json:"sortDir,omitempty"`   // направление сортировки: "asc" или "desc" (по умолчанию "desc")
}

// HotelDTO — DTO для отеля в JSON-запросах и ответах
// Полностью соответствует domain.Hotel, но с JSON-тегами для сериализации
type HotelDTO struct {
	ID        string  `json:"id"`        // ID отеля от поставщика
	Source    string  `json:"source"`    // название поставщика (expedia, tripcom и т.д.)
	Name      string  `json:"name"`      // название отеля
	Address   string  `json:"address"`   // адрес
	City      string  `json:"city"`      // город
	Country   string  `json:"country"`   // страна
	Latitude  float64 `json:"latitude"`  // широта
	Longitude float64 `json:"longitude"` // долгота
}

// ConfigDTO — настройки алгоритма сравнения
// Все поля — указатели, чтобы отличать "не передано" от "передано явно"
type ConfigDTO struct {
	NameWeight     *float64 `json:"nameWeight,omitempty"`     // вес сравнения названий
	AddressWeight  *float64 `json:"addressWeight,omitempty"`  // вес сравнения адресов
	GeoWeight      *float64 `json:"geoWeight,omitempty"`      // вес сравнения координат
	LocationWeight *float64 `json:"locationWeight,omitempty"` // вес сравнения города/страны
	Threshold      *float64 `json:"threshold,omitempty"`      // порог уверенности (0..1)
	Algorithm      *string  `json:"algorithm,omitempty"`      // алгоритм: jaro-winkler, jaro, levenshtein, cosine, и др.
}

// MatchResponse — структура ответа сервера
// Содержит найденные группы, метрики и отели без совпадений
type MatchResponse struct {
	Groups     []GroupDTO     `json:"groups"`     // группы совпавших отелей
	Unmatched  []HotelDTO     `json:"unmatched"`  // отели без совпадений (всегда пустой)
	Metrics    MetricsDTO     `json:"metrics"`    // метрики результата
	Pagination PaginationDTO  `json:"pagination"` // информация о пагинации
}

// PaginationDTO — информация о пагинации
type PaginationDTO struct {
	Page       int  `json:"page"`        // текущая страница
	Limit      int  `json:"limit"`       // элементов на странице
	TotalItems int  `json:"totalItems"`  // всего элементов
	TotalPages int  `json:"totalPages"`  // всего страниц
	HasNext    bool `json:"hasNext"`     // есть ли следующая страница
	HasPrev    bool `json:"hasPrev"`     // есть ли предыдущая страница
}

// MetricsDTO — метрики результата матчинга (формат для фронтенда)
type MetricsDTO struct {
	TotalHotels        int     `json:"totalHotels"`        // всего отелей
	TotalGroups        int     `json:"totalGroups"`        // всего групп (включая одиночные)
	TotalDuplicates    int     `json:"totalDuplicates"`    // всего дубликатов (отели в группах с >=2)
	TotalProviders     int     `json:"totalProviders"`     // количество поставщиков
	AverageConfidence  float64 `json:"averageConfidence"`  // средняя уверенность
}

// GroupDTO — группа совпавших отелей в ответе
type GroupDTO struct {
	GroupID         string         `json:"groupId"`         // уникальный ID группы
	PrimaryName     string         `json:"primaryName"`     // основное название (первый отель)
	ConfidenceScore float64        `json:"confidenceScore"` // степень уверенности (0..1)
	MatchScore      float64        `json:"matchScore"`      // оценка совпадения
	ProvidersCount  int            `json:"providersCount"`  // количество поставщиков в группе
	HotelsCount     int            `json:"hotelsCount"`     // количество отелей в группе
	MatchReasons    map[string]any `json:"matchReasons"`    // причины объединения
	Hotels          []HotelDTO     `json:"hotels"`          // список отелей в группе
}

// ToDomain — преобразует MatchRequest в доменные модели
// Конвертирует DTO → domain.Hotel и domain.Config
// Если настройки не переданы — используются значения по умолчанию
func (r MatchRequest) ToDomain() ([]domain.Hotel, domain.Config) {
	// Преобразуем список отелей
	hotels := make([]domain.Hotel, len(r.Hotels))
	for i, h := range r.Hotels {
		hotels[i] = domain.Hotel{
			ID:        h.ID,
			Source:    h.Source,
			Name:      h.Name,
			Address:   h.Address,
			City:      h.City,
			Country:   h.Country,
			Latitude:  h.Latitude,
			Longitude: h.Longitude,
		}
	}

	// Начинаем с настроек по умолчанию
	cfg := domain.DefaultConfig()

	// Переопределяем только те поля, которые были переданы
	if r.Config.NameWeight != nil {
		cfg.NameWeight = *r.Config.NameWeight
	}
	if r.Config.AddressWeight != nil {
		cfg.AddressWeight = *r.Config.AddressWeight
	}
	if r.Config.GeoWeight != nil {
		cfg.GeoWeight = *r.Config.GeoWeight
	}
	if r.Config.LocationWeight != nil {
		cfg.LocationWeight = *r.Config.LocationWeight
	}
	if r.Config.Threshold != nil {
		cfg.Threshold = *r.Config.Threshold
	}
	if r.Config.Algorithm != nil {
		cfg.Algorithm = *r.Config.Algorithm
	}

	return hotels, cfg
}

// ToDTO — преобразует доменный результат в ответ для клиента
// Конвертирует domain.Result → MatchResponse (DTO)
// Добавлены: метрики, primaryName, matchScore, providersCount, hotelsCount, matchReasons
func ToDTO(result *domain.Result) MatchResponse {
	if result == nil {
		return MatchResponse{
			Groups:    []GroupDTO{},
			Unmatched: []HotelDTO{},
			Metrics:   MetricsDTO{},
			Pagination: PaginationDTO{},
		}
	}

	// Собираем все отели
	allHotels := make([]domain.Hotel, 0)
	for _, g := range result.Groups {
		allHotels = append(allHotels, g.Hotels...)
	}
	allHotels = append(allHotels, result.Unmatched...)

	// Считаем поставщиков
	providers := make(map[string]bool)
	for _, h := range allHotels {
		if h.Source != "" {
			providers[h.Source] = true
		}
	}

	// Все группы (включая одиночные)
	allGroups := make([]domain.Group, 0)
	allGroups = append(allGroups, result.Groups...)

	// Каждый unmatched отель — группа из 1
	for _, hotel := range result.Unmatched {
		allGroups = append(allGroups, domain.Group{
			ID:              hotel.ID + "-single",
			ConfidenceScore: 1.0,
			Hotels:          []domain.Hotel{hotel},
		})
	}

	// Считаем уверенность и дубликаты
	var totalConfidence float64
	var minConfidence float64 = 1.0
	var maxConfidence float64
	confidenceCount := len(result.Groups)
	totalDuplicates := 0

	for _, g := range result.Groups {
		totalConfidence += g.ConfidenceScore
		confidenceCount++
		if g.ConfidenceScore < minConfidence {
			minConfidence = g.ConfidenceScore
		}
		if g.ConfidenceScore > maxConfidence {
			maxConfidence = g.ConfidenceScore
		}

		// Если в группе >=2 отеля — это дубликаты
		if len(g.Hotels) >= 2 {
			totalDuplicates += len(g.Hotels)
		}
	}

	// Формируем метрики для фронтенда
	metrics := MetricsDTO{
		TotalHotels:        len(allHotels),
		TotalGroups:        len(allGroups),
		TotalDuplicates:    totalDuplicates,
		TotalProviders:     len(providers),
		AverageConfidence:  0,
	}

	if confidenceCount > 0 {
		metrics.AverageConfidence = totalConfidence / float64(confidenceCount)
	}

	// Преобразуем группы в DTO
	groupsDTO := make([]GroupDTO, 0, len(allGroups))
	for _, g := range allGroups {
		// Определяем primary name
		primaryName := ""
		if len(g.Hotels) > 0 {
			primaryName = g.Hotels[0].Name
		}

		// Собираем поставщиков в группе
		providersInGroup := make(map[string]bool)
		for _, h := range g.Hotels {
			if h.Source != "" {
				providersInGroup[h.Source] = true
			}
		}
		providersList := make([]string, 0, len(providersInGroup))
		for p := range providersInGroup {
			providersList = append(providersList, p)
		}

		// Причины матчинга
		matchReasons := map[string]any{
			"matched_suppliers": providersList,
			"total":             len(g.Hotels),
			"confidence":        g.ConfidenceScore,
		}

		// Преобразуем отели в DTO
		hotelsDTO := make([]HotelDTO, len(g.Hotels))
		for j, h := range g.Hotels {
			hotelsDTO[j] = HotelDTO{
				ID:        h.ID,
				Source:    h.Source,
				Name:      h.Name,
				Address:   h.Address,
				City:      h.City,
				Country:   h.Country,
				Latitude:  h.Latitude,
				Longitude: h.Longitude,
			}
		}

		groupsDTO = append(groupsDTO, GroupDTO{
			GroupID:         g.ID,
			PrimaryName:     primaryName,
			ConfidenceScore: g.ConfidenceScore,
			MatchScore:      g.ConfidenceScore,
			ProvidersCount:  len(providersInGroup),
			HotelsCount:     len(g.Hotels),
			MatchReasons:    matchReasons,
			Hotels:          hotelsDTO,
		})
	}

	return MatchResponse{
		Groups:    groupsDTO,
		Unmatched: []HotelDTO{}, // всегда пустой, все отели теперь в группах
		Metrics:   metrics,
		Pagination: PaginationDTO{
			Page:       1,
			Limit:      len(groupsDTO),
			TotalItems: len(groupsDTO),
			TotalPages: 1,
			HasNext:    false,
			HasPrev:    false,
		},
	}
}

// ToDTOWithPagination — преобразует доменный результат в ответ с пагинацией
// Конвертирует domain.Result → MatchResponse с учетом пагинации, поиска и сортировки
func ToDTOWithPagination(result *domain.Result, req MatchRequest) MatchResponse {
	// Получаем все группы через стандартный ToDTO
	fullResponse := ToDTO(result)

	// Если групп нет, возвращаем пустой ответ
	if len(fullResponse.Groups) == 0 {
		return MatchResponse{
			Groups:    []GroupDTO{},
			Unmatched: []HotelDTO{},
			Metrics:   fullResponse.Metrics,
			Pagination: PaginationDTO{
				Page:       req.Page,
				Limit:      req.Limit,
				TotalItems: 0,
				TotalPages: 0,
				HasNext:    false,
				HasPrev:    false,
			},
		}
	}

	// Устанавливаем значения по умолчанию
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 {
		req.Limit = 50
	}
	if req.Limit > 500 {
		req.Limit = 500 // Максимальный лимит для защиты
	}
	if req.SortBy == "" {
		req.SortBy = "confidence"
	}
	if req.SortDir == "" {
		req.SortDir = "desc"
	}

	// Копируем группы для фильтрации и сортировки
	allGroups := fullResponse.Groups

	// Применяем поиск (фильтрация по названию)
	if req.Search != "" {
		allGroups = filterGroups(allGroups, req.Search)
	}

	// Применяем сортировку
	allGroups = sortGroups(allGroups, req.SortBy, req.SortDir)

	// Вычисляем пагинацию
	totalItems := len(allGroups)
	totalPages := (totalItems + req.Limit - 1) / req.Limit
	if totalPages == 0 {
		totalPages = 1
	}

	// Корректируем страницу
	if req.Page > totalPages {
		req.Page = totalPages
	}
	if req.Page < 1 {
		req.Page = 1
	}

	// Вычисляем индексы для среза
	start := (req.Page - 1) * req.Limit
	end := start + req.Limit
	if end > totalItems {
		end = totalItems
	}

	// Получаем пагинированные группы
	paginatedGroups := allGroups[start:end]

	return MatchResponse{
		Groups:    paginatedGroups,
		Unmatched: []HotelDTO{},
		Metrics:   fullResponse.Metrics,
		Pagination: PaginationDTO{
			Page:       req.Page,
			Limit:      req.Limit,
			TotalItems: totalItems,
			TotalPages: totalPages,
			HasNext:    req.Page < totalPages,
			HasPrev:    req.Page > 1,
		},
	}
}

// filterGroups — фильтрует группы по поисковому запросу
func filterGroups(groups []GroupDTO, search string) []GroupDTO {
	search = strings.ToLower(strings.TrimSpace(search))
	if search == "" {
		return groups
	}

	result := make([]GroupDTO, 0, len(groups))
	for _, g := range groups {
		if strings.Contains(strings.ToLower(g.PrimaryName), search) {
			result = append(result, g)
		}
	}
	return result
}

// sortGroups — сортирует группы по указанному полю
func sortGroups(groups []GroupDTO, sortBy, sortDir string) []GroupDTO {
	if len(groups) <= 1 {
		return groups
	}

	result := make([]GroupDTO, len(groups))
	copy(result, groups)

	// Определяем функцию сравнения
	less := func(i, j int) bool {
		switch sortBy {
		case "name":
			return result[i].PrimaryName < result[j].PrimaryName
		case "confidence":
			return result[i].ConfidenceScore < result[j].ConfidenceScore
		case "hotelsCount":
			return result[i].HotelsCount < result[j].HotelsCount
		case "providersCount":
			return result[i].ProvidersCount < result[j].ProvidersCount
		default:
			return result[i].GroupID < result[j].GroupID
		}
	}

	// Сортируем
	for i := 0; i < len(result); i++ {
		for j := i + 1; j < len(result); j++ {
			if (sortDir == "asc" && less(i, j)) || (sortDir == "desc" && less(j, i)) {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	return result
}