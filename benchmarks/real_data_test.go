package benchmarks

import (
	"encoding/csv"
	"os"
	"testing"

	"hotel-matcher/internal/pkg/algorithms"
)

// HotelRecord — структура для хранения данных из CSV
type HotelRecord struct {
	Provider     string
	ProviderID   string
	Name         string
	Address      string
	City         string
	Country      string
	CountryCode  string
	Latitude     float64
	Longitude    float64
	Stars        string
}

// LoadHotelsFromCSV загружает отели из CSV
func LoadHotelsFromCSV(path string) ([]HotelRecord, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var hotels []HotelRecord
	// Пропускаем заголовок
	for i := 1; i < len(records); i++ {
		if len(records[i]) < 10 {
			continue
		}
		hotels = append(hotels, HotelRecord{
			Provider:    records[i][0],
			ProviderID:  records[i][1],
			Name:        records[i][2],
			Address:     records[i][3],
			City:        records[i][4],
			Country:     records[i][5],
			CountryCode: records[i][6],
		})
	}
	return hotels, nil
}


// hotel_records_for_students


// BenchmarkMyCSVData -  hotel_records_for_students
// Использует все записи из файла
func BenchmarkMyCSVData(b *testing.B) {
	// Путь к вашему файлу в корне проекта
	filePath := "../hotel_records_for_students.csv"
	
	hotels, err := LoadHotelsFromCSV(filePath)
	if err != nil || len(hotels) == 0 {
		b.Skip("No data found in hotel_records_for_students")
	}

	names := make([]string, len(hotels))
	for i, h := range hotels {
		names[i] = h.Name
	}

	b.Logf("Testing with %d hotels from hotel_records_for_students", len(names))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < len(names)-1; j++ {
			algorithms.JaroWinklerSimilarityCustom(names[j], names[j+1])
		}
	}
}

// BenchmarkMyCSVDataLimit - тест на вашем файле с ограничением количества записей
// limit - сколько записей использовать (можно менять в коде)
func BenchmarkMyCSVDataLimit(b *testing.B) {
	// Путь к вашему файлу в корне проекта
	filePath := "../hotel_records_for_students.csv"
	
	hotels, err := LoadHotelsFromCSV(filePath)
	if err != nil || len(hotels) == 0 {
		b.Skip("No data found in hotel_records_for_students")
	}

	// изменить число для тестов можна тут
	// 10  - быстрый тест
	// 50  - средний тест
	// 100 - хороший тест
	// 500 - большой тест
	// 1000 - очень большой тест

	limit := 100
	
	if len(hotels) < limit {
		limit = len(hotels)
	}

	names := make([]string, limit)
	for i := 0; i < limit; i++ {
		names[i] = hotels[i].Name
	}

	b.Logf("Testing with %d hotels from hotel_records_for_students (limited to %d)", len(hotels), limit)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < len(names)-1; j++ {
			algorithms.JaroWinklerSimilarityCustom(names[j], names[j+1])
		}
	}
}

// BenchmarkMyCSVAllAlgorithms - все алгоритмы на вашем файле
func BenchmarkMyCSVAllAlgorithms(b *testing.B) {
	// Путь к вашему файлу в корне проекта
	filePath := "../hotel_records_for_students.csv"
	
	hotels, err := LoadHotelsFromCSV(filePath)
	if err != nil || len(hotels) == 0 {
		b.Skip("No data found in hotel_records_for_students")
	}

	// изменить число что бы тестить разное кол во данных 
	
	limit := 100
	
	if len(hotels) < limit {
		limit = len(hotels)
	}

	names := make([]string, limit)
	for i := 0; i < limit; i++ {
		names[i] = hotels[i].Name
	}

	b.Logf("Testing %d hotels with all algorithms", limit)

	algorithmsList := []string{
		"jaro-winkler",
		"jaro",
		"levenshtein",
		"damerau-levenshtein",
		"soundex",
		"ngram",
	}

	b.ResetTimer()
	for _, alg := range algorithmsList {
		b.Run(alg, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				for j := 0; j < len(names)-1; j++ {
					algorithms.CompareWithCustomAlgorithm(names[j], names[j+1], alg)
				}
			}
		})
	}
}


// бенчмарк на реальных (сравнение всех попарно)


func BenchmarkRealDataJaroWinkler(b *testing.B) {
	hotels, err := LoadHotelsFromCSV("../data/hotels.csv")
	if err != nil || len(hotels) == 0 {
		b.Skip("No real data found, skipping benchmark")
	}

	names := make([]string, len(hotels))
	for i, h := range hotels {
		names[i] = h.Name
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < len(names)-1; j++ {
			algorithms.JaroWinklerSimilarityCustom(names[j], names[j+1])
		}
	}
}

//бенч только похожие отели


func BenchmarkRealDataSimilarHotels(b *testing.B) {
	hotels, err := LoadHotelsFromCSV("../data/hotels.csv")
	if err != nil || len(hotels) == 0 {
		b.Skip("No real data found, skipping benchmark")
	}

	// Берём только отели, которые должны совпадать
	similarPairs := [][]string{
		{"Rodeway Inn Union", "Rodeway Inn", "Palmetto Inn Union"},
		{"B&B at Bloem", "B&B at Bloem"},
		{"Hotel Taisetsu ONSEN＆CANYON RESORT", "Taisetsu", "Hotel Taisetsu ONSEN&CANYON RESORT"},
		{"Rotary Lodge Port Macquarie", "Rotary Lodge Port Macquarie"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, pair := range similarPairs {
			if len(pair) >= 2 {
				algorithms.JaroWinklerSimilarityCustom(pair[0], pair[1])
			}
		}
	}
}

// Бенч на реальных данных (все алгоритмы)
func BenchmarkRealDataAllAlgorithms(b *testing.B) {
	hotels, err := LoadHotelsFromCSV("../data/hotels.csv")
	if err != nil || len(hotels) == 0 {
		b.Skip("No real data found, skipping benchmark")
	}

	// Берём первые 100 отелей для сравнения
	limit := 100
	if len(hotels) < limit {
		limit = len(hotels)
	}

	names := make([]string, limit)
	for i := 0; i < limit; i++ {
		names[i] = hotels[i].Name
	}

	algorithmsList := []string{
		"jaro-winkler",
		"jaro",
		"levenshtein",
		"damerau-levenshtein",
		"soundex",
		"ngram",
	}

	b.ResetTimer()
	for _, alg := range algorithmsList {
		b.Run(alg, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				for j := 0; j < len(names)-1; j++ {
					algorithms.CompareWithCustomAlgorithm(names[j], names[j+1], alg)
				}
			}
		})
	}
}