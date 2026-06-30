package http

import "hotel-matcher/internal/domain"

// ErrorResponse — JSON-ответ при ошибке
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// MatchRequest — структура входящего запроса к эндпоинту /api/match
// Содержит список отелей и опциональные настройки матчинга
type MatchRequest struct {
	Hotels []HotelDTO `json:"hotels"`              // список отелей от поставщиков
	Config ConfigDTO  `json:"config,omitempty"`    // настройки алгоритма (необязательно)
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
	Groups    []GroupDTO `json:"groups"`    // группы совпавших отелей
	Unmatched []HotelDTO `json:"unmatched"` // отели без совпадений (всегда пустой)
	Metrics   MetricsDTO `json:"metrics"`   // метрики результата
}

// MetricsDTO — метрики результата матчинга
type MetricsDTO struct {
	TotalHotels       int     `json:"totalHotels"`       // всего отелей
	TotalGroups       int     `json:"totalGroups"`       // всего групп (включая одиночные)
	UniqueHotels      int     `json:"uniqueHotels"`      // уникальных отелей
	GroupsWithMatches int     `json:"groupsWithMatches"` // группы с >=2 отелями
	SingleHotels      int     `json:"singleHotels"`      // одиночные отели
	ProvidersCount    int     `json:"providersCount"`    // количество поставщиков
	AvgConfidence     float64 `json:"avgConfidence"`     // средняя уверенность
	MinConfidence     float64 `json:"minConfidence"`     // минимальная уверенность
	MaxConfidence     float64 `json:"maxConfidence"`     // максимальная уверенность
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

	// Считаем уверенность
	var totalConfidence float64
	var minConfidence float64 = 1.0
	var maxConfidence float64
	confidenceCount := len(result.Groups)

	for _, g := range result.Groups {
		totalConfidence += g.ConfidenceScore
		if g.ConfidenceScore < minConfidence {
			minConfidence = g.ConfidenceScore
		}
		if g.ConfidenceScore > maxConfidence {
			maxConfidence = g.ConfidenceScore
		}
	}

	// Формируем метрики
	metrics := MetricsDTO{
		TotalHotels:       len(allHotels),
		TotalGroups:       len(allGroups),
		UniqueHotels:      len(allGroups),
		GroupsWithMatches: len(result.Groups),
		SingleHotels:      len(result.Unmatched),
		ProvidersCount:    len(providers),
		AvgConfidence:     0,
		MinConfidence:     0,
		MaxConfidence:     0,
	}

	if confidenceCount > 0 {
		metrics.AvgConfidence = totalConfidence / float64(confidenceCount)
		metrics.MinConfidence = minConfidence
		metrics.MaxConfidence = maxConfidence
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
	}
}