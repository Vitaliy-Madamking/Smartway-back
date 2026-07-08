package http

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"

	"hotel-matcher/internal/domain"
)

// parseCSV парсит CSV с отелями
func parseCSV(r io.Reader) ([]domain.Hotel, error) {
	// 1. Удаляем BOM (если есть)
	reader := bufio.NewReader(r)
	bom := []byte{0xEF, 0xBB, 0xBF}
	peek, err := reader.Peek(3)
	if err == nil && bytes.Equal(peek, bom) {
		reader.Discard(3)
	}

	// 2. Создаём CSV-ридер с ПРАВИЛЬНЫМИ настройками
	csvReader := csv.NewReader(reader)
	csvReader.Comma = ','               // разделитель — запятая
	csvReader.TrimLeadingSpace = true   // убираем пробелы в начале полей
	csvReader.FieldsPerRecord = -1      // разрешаем разное количество полей
	csvReader.LazyQuotes = true         // ✅ ГЛАВНОЕ: игнорируем ошибки с кавычками!
	csvReader.ReuseRecord = true        // оптимизация

	// 3. Читаем заголовки
	header, err := csvReader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}

	// 4. Строим карту колонок (без учёта регистра)
	idxMap := make(map[string]int)
	for i, col := range header {
		idxMap[strings.ToLower(strings.TrimSpace(col))] = i
	}

	// 5. Добавляем синонимы для колонок (если в CSV названия другие)
	// Если есть "source" — считаем её "provider"
	if idx, ok := idxMap["source"]; ok {
		idxMap["provider"] = idx
	}
	// Если есть "hotel_id" — считаем её "provider_hotel_id"
	if idx, ok := idxMap["hotel_id"]; ok {
		idxMap["provider_hotel_id"] = idx
	}
	// Если есть "hotel_name" или "title" — считаем их "name"
	if idx, ok := idxMap["hotel_name"]; ok {
		idxMap["name"] = idx
	}
	if idx, ok := idxMap["title"]; ok {
		idxMap["name"] = idx
	}
	// Если есть "lat" — считаем её "latitude"
	if idx, ok := idxMap["lat"]; ok {
		idxMap["latitude"] = idx
	}
	// Если есть "lon" или "lng" — считаем их "longitude"
	if idx, ok := idxMap["lon"]; ok {
		idxMap["longitude"] = idx
	}
	if idx, ok := idxMap["lng"]; ok {
		idxMap["longitude"] = idx
	}

	// 6. Проверяем обязательные колонки
	requiredCols := []string{"provider", "provider_hotel_id", "name"}
	for _, col := range requiredCols {
		if _, ok := idxMap[col]; !ok {
			return nil, fmt.Errorf("missing required column: %s (available: %v)", col, header)
		}
	}

	log.Printf("✅ CSV headers: %v", header)
	log.Printf("✅ Column mapping: %v", idxMap)

	// 7. Читаем данные
	var hotels []domain.Hotel
	lineNum := 1

	for {
		record, err := csvReader.Read()
		lineNum++

		if err == io.EOF {
			break
		}
		if err != nil {
			// Пропускаем проблемные строки, но логируем их
			log.Printf("⚠️ Warning: skipping line %d: %v", lineNum, err)
			continue
		}

		if len(record) < len(header) {
			continue
		}

		// Функция для получения значения по имени колонки
		get := func(key string) string {
			if idx, ok := idxMap[strings.ToLower(key)]; ok && idx < len(record) {
				return strings.TrimSpace(record[idx])
			}
			return ""
		}

		provider := get("provider")
		providerID := get("provider_hotel_id")
		name := get("name")
		address := get("address")
		city := get("city")
		country := get("country")
		latStr := get("latitude")
		lonStr := get("longitude")

		// Пропускаем строки без ID или названия
		if providerID == "" || name == "" {
			continue
		}

		lat, _ := strconv.ParseFloat(latStr, 64)
		lon, _ := strconv.ParseFloat(lonStr, 64)

		hotel := domain.Hotel{
			ID:        providerID,
			Source:    provider,
			Name:      name,
			Address:   address,
			City:      city,
			Country:   country,
			Latitude:  lat,
			Longitude: lon,
		}

		hotels = append(hotels, hotel)
	}

	if len(hotels) == 0 {
		return nil, fmt.Errorf("no hotels found in CSV")
	}

	log.Printf("✅ Parsed %d hotels from CSV", len(hotels))
	return hotels, nil
}