package http

import (
	"encoding/csv"                  // для парсинга CSV-файлов
	"hotel-matcher/internal/domain" // бизнес-сущности
	"io"                            // для работы с потоками (io.Reader)
	"strconv"                       // конвертация строк в числа
	"strings"                       // работа со строками
)

// parseCSV — парсит CSV-файл в список отелей (domain.Hotel)
// Принимает io.Reader (можно передать *os.File или HTTP-тело)
// Возвращает []domain.Hotel и ошибку
func parseCSV(r io.Reader) ([]domain.Hotel, error) {
	// Создаём CSV-ридер
	reader := csv.NewReader(r)
	reader.Comma = ','               // разделитель — запятая
	reader.TrimLeadingSpace = true   // удаляем пробелы в начале полей
	reader.FieldsPerRecord = -1      // разрешаем переменное количество полей

	// ШАГ 1: Читаем заголовок
	header, err := reader.Read()
	if err != nil {
		return nil, err // ошибка чтения заголовка
	}

	// Строим карту: имя колонки → её индекс
	idxMap := make(map[string]int)
	for i, col := range header {
		idxMap[strings.ToLower(strings.TrimSpace(col))] = i
	}

	// ШАГ 2: Читаем строки с данными
	var hotels []domain.Hotel
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break // достигнут конец файла
		}
		if err != nil {
			continue // пропускаем проблемные строки
		}
		if len(record) < len(header) {
			continue // строка короче заголовка — пропускаем
		}

		// get — вспомогательная функция для получения значения по имени колонки
		get := func(key string) string {
			if idx, ok := idxMap[strings.ToLower(key)]; ok && idx < len(record) {
				return strings.TrimSpace(record[idx])
			}
			return ""
		}

		// Извлекаем поля по именам колонок
		provider := get("provider")            // поставщик (expedia, tripcom и т.д.)
		providerID := get("provider_hotel_id") // ID отеля от поставщика
		name := get("name")                    // название отеля
		address := get("address")              // адрес
		city := get("city")                    // город
		country := get("country")              // страна
		latStr := get("latitude")              // широта (строка)
		lonStr := get("longitude")             // долгота (строка)

		// Пропускаем строки без ID или названия (обязательные поля)
		if providerID == "" || name == "" {
			continue
		}

		// Парсим координаты (если пустые или кривые — будет 0)
		lat, _ := strconv.ParseFloat(latStr, 64)
		lon, _ := strconv.ParseFloat(lonStr, 64)

		// Создаём доменный объект отеля
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

		// Добавляем в результат
		hotels = append(hotels, hotel)
	}

	return hotels, nil
}