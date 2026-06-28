// cmd/loadtest/main.go - Утилита для тестирования API(переедлать с косиками)
// Назначение: загружает CSV с отелями и отправляет на сервер
// Использование: go run cmd/loadtest/main.go -file data/hotels.csv -v

package main

import (
	"bytes"         // Для работы с буферами
	"encoding/csv"  // Парсинг CSV
	"encoding/json" // Сериализация JSON
	"flag"          // Парсинг флагов командной строки
	"fmt"           // Форматированный вывод
	"io"            // Ввод-вывод
	"log"           // Логирование
	"net/http"      // HTTP-клиент
	"os"            // Работа с файлами
	"strconv"       // Конвертация строк в числа
	"strings"       // Работа со строками
)

// DTO (Data Transfer Objects) - структуры для обмена с API

// HotelDTO - структура отеля для отправки на сервер
// Совпадает с той, что использует бэкенд
type HotelDTO struct {
	ID        string  `json:"id"`        // ID отеля от поставщика
	Source    string  `json:"source"`    // Поставщик (expedia, tripcom и т.д.)
	Name      string  `json:"name"`      // Название отеля
	Address   string  `json:"address"`   // Адрес
	City      string  `json:"city"`      // Город
	Country   string  `json:"country"`   // Страна
	Latitude  float64 `json:"latitude"`  // Широта
	Longitude float64 `json:"longitude"` // Долгота
}

// MatchRequest - запрос к API /api/match
type MatchRequest struct {
	Hotels []HotelDTO `json:"hotels"`          // Список отелей
	Config *ConfigDTO `json:"config,omitempty"` // Настройки (опционально)
}

// ConfigDTO - настройки матчинга для запроса
type ConfigDTO struct {
	NameWeight     *float64 `json:"nameWeight,omitempty"`     // Вес названия
	AddressWeight  *float64 `json:"addressWeight,omitempty"`  // Вес адреса
	GeoWeight      *float64 `json:"geoWeight,omitempty"`      // Вес координат
	LocationWeight *float64 `json:"locationWeight,omitempty"` // Вес локации
	Threshold      *float64 `json:"threshold,omitempty"`      // Порог уверенности
}


// Флаги командной строки


var (
	csvFile    = flag.String("file", "data/hotels.csv", "путь к CSV-файлу")
	serverURL  = flag.String("url", "http://localhost:8080/api/match", "URL сервера")
	outputFile = flag.String("out", "", "сохранить ответ в файл")
	verbose    = flag.Bool("v", false, "подробный вывод")
	threshold  = flag.Float64("threshold", 0.75, "порог уверенности (0-1)")
)

func main() {
	flag.Parse() // Парсим флаги командной строки

	// Если включён подробный режим - выводим информацию
	if *verbose {
		log.Printf("чтение файла: %s", *csvFile)
	}

	// Читаем CSV
	hotels, err := readCSV(*csvFile)
	if err != nil {
		log.Fatalf("ошибка CSV: %v", err)
	}
	if len(hotels) == 0 {
		log.Fatal("отелей не найдено")
	}
	log.Printf("прочитано %d отелей", len(hotels))

	// Формируем запрос
	req := MatchRequest{
		Hotels: hotels,
		Config: &ConfigDTO{
			Threshold: threshold, // Передаём порог из флага
		},
	}

	// Сериализуем в JSON
	jsonData, _ := json.Marshal(req)

	// Отправляем POST-запрос на сервер
	resp, err := http.Post(*serverURL, "application/json", bytes.NewReader(jsonData))
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	// Читаем ответ
	body, _ := io.ReadAll(resp.Body)

	// Проверяем статус
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("статус %d: %s", resp.StatusCode, string(body))
	}

	// Сохраняем или выводим результат
	if *outputFile != "" {
		os.WriteFile(*outputFile, body, 0644)
		log.Printf("сохранено в %s", *outputFile)
	} else {
		// Красиво форматируем JSON для вывода
		var pretty bytes.Buffer
		json.Indent(&pretty, body, "", "  ")
		fmt.Println(pretty.String())
	}
}


// readCSV - чтение CSV-файла с отелями

func readCSV(path string) ([]HotelDTO, error) {
	// Открываем файл
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Создаём CSV-ридер
	reader := csv.NewReader(f)
	reader.Comma = ','               // Разделитель - запятая
	reader.TrimLeadingSpace = true   // Убираем пробелы
	reader.FieldsPerRecord = -1      // Переменное число полей

	// Читаем заголовок
	header, err := reader.Read()
	if err != nil {
		return nil, err
	}

	// Строим карту: имя колонки → индекс
	idxMap := make(map[string]int)
	for i, col := range header {
		idxMap[strings.ToLower(strings.TrimSpace(col))] = i
	}

	var hotels []HotelDTO

	// Читаем строки по одной
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break // Конец файла
		}
		if err != nil {
			continue // Пропускаем проблемные строки
		}
		if len(record) < len(header) {
			continue // Строка короче заголовка - пропускаем
		}

		// Вспомогательная функция: получить значение по имени колонки
		get := func(key string) string {
			if idx, ok := idxMap[strings.ToLower(key)]; ok && idx < len(record) {
				return strings.TrimSpace(record[idx])
			}
			return ""
		}

		// Извлекаем поля
		provider := get("provider")                 // Поставщик
		providerID := get("provider_hotel_id")      // ID отеля
		name := get("name")                         // Название
		address := get("address")                   // Адрес
		city := get("city")                         // Город
		country := get("country")                   // Страна
		latStr := get("latitude")                   // Широта (строка)
		lonStr := get("longitude")                  // Долгота (строка)

		// Пропускаем строки без ID или названия
		if providerID == "" || name == "" {
			continue
		}

		// Парсим координаты (если пустые → 0)
		lat, _ := strconv.ParseFloat(latStr, 64)
		lon, _ := strconv.ParseFloat(lonStr, 64)

		// Создаём DTO и добавляем в список
		hotels = append(hotels, HotelDTO{
			ID:        providerID,
			Source:    provider,
			Name:      name,
			Address:   address,
			City:      city,
			Country:   country,
			Latitude:  lat,
			Longitude: lon,
		})
	}

	return hotels, nil
}