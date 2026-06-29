package domain

// Config - настройки алгоритма матчинга
type Config struct {
	Threshold      float64 `json:"threshold"`
	NameWeight     float64 `json:"name_weight"`
	AddressWeight  float64 `json:"address_weight"`
	GeoWeight      float64 `json:"geo_weight"`
	LocationWeight float64 `json:"location_weight"`
	Algorithm      string  `json:"algorithm"`
}

// DefaultConfig возвращает настройки по умолчанию
func DefaultConfig() Config {
	return Config{
		Threshold:      0.85,
		NameWeight:     0.4,
		AddressWeight:  0.3,
		GeoWeight:      0.2,
		LocationWeight: 0.1,
		Algorithm:      "jaro-winkler",
	}
}
