package usecase

import (
	"context" // Для работы с контекстом (таймауты, отмена)
	"fmt"
	"sync" // Для синхронизации горутин (мьютексы, WaitGroup)

	"hotel-matcher/internal/domain"         // Бизнес-сущности
	"hotel-matcher/internal/pkg/algorithms" // Алгоритмы сравнения
)

const (
	ReasonName     = "Similar names"
	ReasonAddress  = "Similar addresses"
	ReasonGeo      = "Similar geo"
	ReasonLocation = "Similar locations"
)

// matcherImpl - структура, реализующая интерфейс Matcher
type matcherImpl struct {
	repo HotelRepository // Репозиторий(переделать под постгру) для работы с отелями (пока не используется)
}

func NewMatcher(repo HotelRepository) Matcher {
	return &matcherImpl{repo: repo}
}

// Match - ОСНОВНАЯ ФУНКЦИЯ: выполняет сопоставление отелей
// Принимает:
//   - ctx: контекст для управления временем жизни запроса
//   - hotels: список отелей для сопоставления
//   - cfg: конфигурация (веса, порог, алгоритм)
//
// Возвращает:
//   - *domain.Result: группы отелей и несоответствующие отели
//   - error: ошибка, если что-то пошло не так
func (m *matcherImpl) Match(ctx context.Context, hotels []domain.Hotel, cfg domain.Config) (*domain.Result, error) {
	if len(hotels) == 0 { // Проверка: если список отелей пуст - возвращаем ошибку
		return nil, domain.ErrNoHotels
	}
	if cfg.Threshold < 0 || cfg.Threshold > 1 { // Проверка: порог должен быть в диапазоне [0, 1]
		return nil, domain.ErrInvalidConfig
	}

	// БЛОКИРОВКА (Blocking) - группируем по стране+городу для уменьшения сложности
	// Группируем отели по стране+городу
	// Это уменьшает сложность с O(n²) до O(n²/m), где m - количество блоков
	// Вместо сравнения ВСЕХ отелей со ВСЕМИ, мы сравниваем только внутри одного блока
	blocks := buildBlocks(hotels)
	// Создаём структуры для хранения результатов
	var mu sync.Mutex // Мьютекс для защиты общих данных от конкурентного доступа

	groups := make(map[string][]domain.Hotel) // Карта: ID группы -> список отелей в группе
	used := make(map[string]bool)             // Карта: ID отеля -> использован ли он уже
	var wg sync.WaitGroup                     // WaitGroup для ожидания завершения всех горутин

	// Параллельная обработка блоков (ускоряем в 2 раза)
	// Для каждого блока запускаем отдельную горутину
	for _, block := range blocks {
		wg.Add(1) // Увеличиваем счётчик горутин на 1
		// Запускаем горутину для обработки блока
		go func(h []domain.Hotel) {
			defer wg.Done()                           // При завершении горутины уменьшаем счётчикИ
			m.processBlock(h, cfg, &mu, groups, used) // Обрабатываем блок (сравниваем отели внутри него)
		}(block)
	}
	wg.Wait() // Ожидаем завершения ВСЕХ горутин

	// Формируем результат
	result := &domain.Result{
		Groups:    make([]domain.Group, 0), // Инициализируем слайс для групп
		Unmatched: make([]domain.Hotel, 0), // Инициализируем слайс для несоответствующих отелей
	}
	for groupID, hotelsInGroup := range groups { // Проходим по всем найденным группам
		score, reasons := calculateGroupConfidence(hotelsInGroup, cfg) // Вычисляем степень уверенности для группы (средняя оценка)
		result.Groups = append(result.Groups, domain.Group{            // Добавляем группу в результат
			ID:              groupID,       // Уникальный ID группы
			ConfidenceScore: score,         // Степень уверенности
			Hotels:          hotelsInGroup, // Список отелей в группе
			MatchReasons:    reasons,       // Причины совпадения
		})
	}
	// Собираем несоответствующие отели (которые не попали ни в одну группу)
	for _, hotel := range hotels {
		// Если отель не был использован (не попал в группу)
		if !used[hotel.ID] {
			// Добавляем в список несоответствующих
			result.Unmatched = append(result.Unmatched, hotel)
		}
	}
	return result, nil
}

// buildBlocks - группировка по стране+городу (блокировка)
// Уменьшает количество попарных сравнений
func buildBlocks(hotels []domain.Hotel) map[string][]domain.Hotel { // Принимает список отелей, возвращает карту: ключ = страна|город, значение = список отелей
	blocks := make(map[string][]domain.Hotel)
	// Создаём карту для блоков
	for _, h := range hotels { // Проходим по всем отелям
		key := fmt.Sprintf("%s|%s", h.Country, h.City) // Формируем ключ из страны и города (разделитель "|")
		blocks[key] = append(blocks[key], h)           // Добавляем отель в соответствующий блок
	}
	// Возвращаем карту блоков
	return blocks
}

// processBlock - обработка одного блока (кластеризация)
// Сравнивает все отели внутри блока и формирует группы
// Принимает:
//   - hotels: список отелей в блоке
//   - cfg: конфигурация
//   - mu: мьютекс для синхронизации
//   - groups: карта для записи найденных групп
//   - used: карта для отслеживания использованных отелей
//
// processBlock - обработка одного блока (кластеризация)
func (m *matcherImpl) processBlock(hotels []domain.Hotel, cfg domain.Config, mu *sync.Mutex,
	groups map[string][]domain.Hotel, used map[string]bool) {

	if len(hotels) <= 1 { // Если в блоке 1 или 0 отелей - группировка не нужна
		return
	}
	// Проходим по всем отелям в блоке
	for i := 0; i < len(hotels); i++ {
		// Проверяем used с мьютексом (безопасно для горутин)
		mu.Lock()               //Блокирвка  для безопасного доступа к used
		if used[hotels[i].ID] { // Если отель уже использован (попал в группу) - пропускаем
			mu.Unlock() // Разблокируем
			continue    // Переходим к следующему отелю
		}
		// Помечаем отель как использованный
		used[hotels[i].ID] = true
		mu.Unlock() // Разблокируем мьютекс

		cluster := []domain.Hotel{hotels[i]} // Начинаем новую группу с текущего отеля

		// Сравниваем каждый с каждым внутри блока
		for j := i + 1; j < len(hotels); j++ {
			mu.Lock()
			if used[hotels[j].ID] {
				mu.Unlock()
				continue
			}
			mu.Unlock()

			// Вычисляем общую оценку совпадения
			score, _ := calculateMatchScore(hotels[i], hotels[j], cfg)
			if score >= cfg.Threshold {
				mu.Lock()
				if !used[hotels[j].ID] {
					used[hotels[j].ID] = true
					cluster = append(cluster, hotels[j])
				}
				mu.Unlock()
			}
		}
		// Если в группе больше 1 отеля - сохраняем группу
		if len(cluster) > 1 {
			mu.Lock()
			// Создаём уникальный ID группы (group-0, group-1, и тд ...)
			groupID := fmt.Sprintf("group-%d", len(groups))
			groups[groupID] = cluster
			mu.Unlock()
		} else {
			// Если отель не совпал ни с кем - снимаем метку used
			// Он будет добавлен в Unmatched позже
			mu.Lock()
			used[hotels[i].ID] = false
			mu.Unlock()
		}
	}
}

// calculateMatchScore - вычисляет общую оценку совпадения (взвешенная сумма)
// Сравнивает два отеля по 4 критериям:
//  1. Названия (Jaro-Winkler)
//  2. Адреса (Levenshtein)
//  3. Координаты (Haversine)
//  4. Город/страна (Jaro)
//
// Каждый критерий умножается на свой вес из конфига
// Возвращает оценку в диапазоне [0, 1]
func calculateMatchScore(h1, h2 domain.Hotel, cfg domain.Config) (float64, []string) {
	alg := cfg.Algorithm

	// Каждый критерий сравнивается своим алгоритмом
	nameScore := algorithms.CompareNamesWithAlgorithm(h1.Name, h2.Name, alg)                           // Учитывает префиксный бонус, хорошо для коротких строк с опечатками(Jaro-Winkler)
	addrScore := algorithms.CompareAddressesWithAlgorithm(h1.Address, h2.Address, alg)                 // Учитывает вставки/удаления/замены, хорошо для длинных строк((Levenshtein)
	geoScore := algorithms.CompareCoordinates(h1.Latitude, h1.Longitude, h2.Latitude, h2.Longitude)    // Вычисляет географическое расстояние в км и преобразует в оценку
	locScore := algorithms.CompareLocationWithAlgorithm(h1.City, h1.Country, h2.City, h2.Country, alg) // Страна - точное совпадение, город - нечёткое сравнен

	reasons := findMatchReasons(nameScore, addrScore, geoScore, locScore)

	// Взвешенная сумма
	return cfg.NameWeight*nameScore +
			cfg.AddressWeight*addrScore +
			cfg.GeoWeight*geoScore +
			cfg.LocationWeight*locScore,
		reasons
}

// calculateGroupConfidence - средняя попарная оценка внутри группы
// Принимает список отелей в группе и конфигурацию
// Возвращает среднюю попарную оценку внутри группы и причины сходства
// Если в группе 1 отель - уверенность = 1.0 (100%)
func calculateGroupConfidence(hotels []domain.Hotel, cfg domain.Config) (float64, []string) {
	if len(hotels) <= 1 {
		return 1.0, nil
	}

	var total float64
	var count int

	// Множество уникальных причин
	reasonsSet := make(map[string]struct{})

	for i := 0; i < len(hotels); i++ {
		for j := i + 1; j < len(hotels); j++ {
			rating, reasons := calculateMatchScore(hotels[i], hotels[j], cfg)

			total += rating
			count++

			// Добавляем причины в множество
			for _, reason := range reasons {
				reasonsSet[reason] = struct{}{}
			}
		}
	}

	if count == 0 {
		return 1.0, nil
	}

	// Преобразуем множество в слайс
	uniqueReasons := make([]string, 0, len(reasonsSet))
	for reason := range reasonsSet {
		uniqueReasons = append(uniqueReasons, reason)
	}

	return total / float64(count), uniqueReasons
}

func findMatchReasons(nameScore, addrScore, geoScore, locScore float64) []string {
	var reasons []string

	if nameScore >= 0.7 {
		reasons = append(reasons, ReasonName)
	}
	if addrScore >= 0.7 {
		reasons = append(reasons, ReasonAddress)
	}
	if geoScore >= 0.7 {
		reasons = append(reasons, ReasonGeo)
	}
	if locScore >= 0.7 {
		reasons = append(reasons, ReasonLocation)
	}

	return reasons
}
