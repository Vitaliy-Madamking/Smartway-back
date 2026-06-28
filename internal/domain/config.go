package domain

// Config - настройки алгоритма матчинга
type Config struct {
	NameWeight     float64 // Вес для сравнения названий (0.35 по умолчанию)
	AddressWeight  float64 // Вес для сравнения адресов (0.25 по умолчанию)
	GeoWeight      float64 // Вес для сравнения координат (0.30 по умолчанию)
	LocationWeight float64 // Вес для сравнения города/страны (0.10 по умолчанию)
	Threshold      float64 // Порог уверенности для объединения (0.75 по умолчанию)
	Algorithm      string  // Алгоритм сравнения: "jaro-winkler", "jaro", "levenshtein", "cosine"
}

// DefaultConfig возвращает настройки по умолчанию
func DefaultConfig() Config {
	return Config{
		NameWeight:     0.35,
		AddressWeight:  0.25,
		GeoWeight:      0.30,
		LocationWeight: 0.10,
		Threshold:      0.75,
		Algorithm:      "jaro-winkler", // По умолчанию Jaro-Winkler 
	}
}