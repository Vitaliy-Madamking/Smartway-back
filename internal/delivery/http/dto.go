package http

import "hotel-matcher/internal/domain"

// MatchRequest — структура входящего запроса к эндпоинту /api/match
// Содержит список отелей и опциональные настройки матчинга
type MatchRequest struct {
	Hotels []HotelDTO `json:"hotels"`          // список отелей от поставщиков
	Config ConfigDTO  `json:"config,omitempty"` // настройки алгоритма (необязательно)
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
// Содержит найденные группы и отели без совпадений
type MatchResponse struct {
	Groups    []GroupDTO `json:"groups"`    // группы совпавших отелей
	Unmatched []HotelDTO `json:"unmatched"` // отели без совпадений
}

// GroupDTO — группа совпавших отелей в ответе
type GroupDTO struct {
	GroupID         string     `json:"groupId"`         // уникальный ID группы
	ConfidenceScore float64    `json:"confidenceScore"` // степень уверенности (0..1)
	Hotels          []HotelDTO `json:"hotels"`          // список отелей в группе
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
func ToDTO(result *domain.Result) MatchResponse {
	// Инициализируем ответ с нужной ёмкостью
	resp := MatchResponse{
		Groups:    make([]GroupDTO, len(result.Groups)),
		Unmatched: make([]HotelDTO, len(result.Unmatched)),
	}

	// Преобразуем группы
	for i, g := range result.Groups {
		// Конвертируем отели внутри группы
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
		resp.Groups[i] = GroupDTO{
			GroupID:         g.ID,
			ConfidenceScore: g.ConfidenceScore,
			Hotels:          hotelsDTO,
		}
	}

	// Преобразуем несоответствующие отели
	for i, h := range result.Unmatched {
		resp.Unmatched[i] = HotelDTO{
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

	return resp
}