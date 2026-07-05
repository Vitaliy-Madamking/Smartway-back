package usecase

import (
	"context"
	"fmt"
	"sync"

	"hotel-matcher/internal/domain"
	"hotel-matcher/internal/pkg/algorithms"
)

// UniversalMatcher — универсальный матчер с адаптивными весами
type UniversalMatcher struct {
	repo HotelRepository
}

// NewUniversalMatcher — конструктор
func NewUniversalMatcher(repo HotelRepository) *UniversalMatcher {
	return &UniversalMatcher{repo: repo}
}

// Match — выполняет матчинг с универсальной комбинацией алгоритмов
func (m *UniversalMatcher) Match(ctx context.Context, hotels []domain.Hotel, cfg domain.Config) (*domain.Result, error) {
	if len(hotels) == 0 {
		return nil, domain.ErrNoHotels
	}
	if cfg.Threshold < 0 || cfg.Threshold > 1 {
		return nil, domain.ErrInvalidConfig
	}

	// 1. БЛОКИНГ: улучшенная группировка по стране+городу
	blocks := buildUniversalBlocks(hotels)

	var mu sync.Mutex
	groups := make(map[string][]domain.Hotel)
	used := make(map[string]bool)
	var wg sync.WaitGroup

	// 2. Параллельная обработка блоков
	for _, block := range blocks {
		wg.Add(1)
		go func(h []domain.Hotel) {
			defer wg.Done()
			m.processUniversalBlock(h, cfg, &mu, groups, used)
		}(block)
	}
	wg.Wait()

	// 3. Формируем результат
	result := &domain.Result{
		Groups:    make([]domain.Group, 0),
		Unmatched: make([]domain.Hotel, 0),
	}

	for groupID, hotelsInGroup := range groups {
		// Вычисляем оценки для группы
		matchScore, reasons := calculateUniversalGroupConfidence(hotelsInGroup, cfg)

		// Считаем поставщиков в группе
		providersInGroup := make(map[string]bool)
		for _, h := range hotelsInGroup {
			if h.Source != "" {
				providersInGroup[h.Source] = true
			}
		}

		// Итоговая уверенность
		confidence := calculateUniversalConfidenceScore(matchScore, len(hotelsInGroup), len(providersInGroup))

		var pairwiseMatrix []domain.PairwiseSimilarity
		var featureContribution domain.FeatureContribution

		pairwiseMatrix = calculatePairwiseMatrix(hotelsInGroup, cfg)
		featureContribution = calculateFeatureContribution(hotelsInGroup, cfg)

		result.Groups = append(result.Groups, domain.Group{
			ID:                  groupID,
			ConfidenceScore:     confidence,
			MatchScore:          matchScore,
			Hotels:              hotelsInGroup,
			MatchReasons:        reasons,
			PairwiseMatrix:      pairwiseMatrix,
			FeatureContribution: featureContribution,
		})
	}

	// Собираем несоответствующие отели
	for _, hotel := range hotels {
		if !used[hotel.ID] {
			result.Unmatched = append(result.Unmatched, hotel)
		}
	}

	return result, nil
}

// buildUniversalBlocks — улучшенная блокировка
func buildUniversalBlocks(hotels []domain.Hotel) map[string][]domain.Hotel {
	blocks := make(map[string][]domain.Hotel)

	for _, h := range hotels {
		// Основной ключ: страна|город
		key := fmt.Sprintf("%s|%s", h.Country, h.City)

		// Если город пустой — используем только страну
		if h.City == "" {
			key = fmt.Sprintf("%s|unknown", h.Country)
		}

		// Если страна пустая — используем координатную сетку
		if h.Country == "" && h.Latitude != 0 && h.Longitude != 0 {
			latBlock := int(h.Latitude / 5) // Блоки по 5 градусов
			lonBlock := int(h.Longitude / 5)
			key = fmt.Sprintf("geo|%d|%d", latBlock, lonBlock)
		}

		blocks[key] = append(blocks[key], h)
	}

	return blocks
}

// processUniversalBlock — обработка блока с универсальной комбинацией
func (m *UniversalMatcher) processUniversalBlock(
	hotels []domain.Hotel,
	cfg domain.Config,
	mu *sync.Mutex,
	groups map[string][]domain.Hotel,
	used map[string]bool,
) {
	if len(hotels) <= 1 {
		return
	}

	for i := 0; i < len(hotels); i++ {
		mu.Lock()
		if used[hotels[i].ID] {
			mu.Unlock()
			continue
		}
		used[hotels[i].ID] = true
		mu.Unlock()

		cluster := []domain.Hotel{hotels[i]}

		for j := i + 1; j < len(hotels); j++ {
			mu.Lock()
			if used[hotels[j].ID] {
				mu.Unlock()
				continue
			}
			mu.Unlock()

			// Универсальная оценка совпадения
			score, _ := calculateUniversalScore(hotels[i], hotels[j], cfg)

			if score >= cfg.Threshold {
				mu.Lock()
				if !used[hotels[j].ID] {
					used[hotels[j].ID] = true
					cluster = append(cluster, hotels[j])
				}
				mu.Unlock()
			}
		}

		if len(cluster) > 1 {
			mu.Lock()
			groupID := fmt.Sprintf("universal-group-%d", len(groups))
			groups[groupID] = cluster
			mu.Unlock()
		} else {
			mu.Lock()
			used[hotels[i].ID] = false
			mu.Unlock()
		}
	}
}

// calculateUniversalScore — универсальная оценка с адаптивными весами
func calculateUniversalScore(h1, h2 domain.Hotel, cfg domain.Config) (float64, []string) {
	// 1. НАЗВАНИЕ: Jaro-Winkler (лучший для названий)
	nameScore := algorithms.CompareNamesWithAlgorithm(h1.Name, h2.Name, "jaro-winkler")

	// 2. АДРЕС: Levenshtein + нормализация
	addrScore := algorithms.CompareAddressesWithAlgorithm(h1.Address, h2.Address, "levenshtein")

	// 3. КООРДИНАТЫ: Haversine (гео-расстояние)
	geoScore := algorithms.CompareCoordinates(h1.Latitude, h1.Longitude, h2.Latitude, h2.Longitude)

	// 4. ГОРОД/СТРАНА: Jaro для города, точное совпадение для страны
	locScore := calculateUniversalLocationScore(h1.City, h1.Country, h2.City, h2.Country)

	// 5. Собираем причины совпадения
	reasons := findUniversalMatchReasons(nameScore, addrScore, geoScore, locScore)

	// 6. АДАПТИВНЫЕ ВЕСА (подстраиваются под данные)
	weights := calculateAdaptiveWeights(h1, h2, cfg)

	// 7. ИТОГОВАЯ ОЦЕНКА
	score := weights.NameWeight*nameScore +
		weights.AddressWeight*addrScore +
		weights.GeoWeight*geoScore +
		weights.LocationWeight*locScore

	return score, reasons
}

// calculateUniversalLocationScore — оценка города и страны
func calculateUniversalLocationScore(city1, country1, city2, country2 string) float64 {
	// Страна — точное совпадение (без исключений)
	if country1 != country2 {
		return 0.0
	}

	// Город — Jaro (учитывает перестановки)
	if city1 == "" || city2 == "" {
		return 0.5 // Нейтральная оценка при отсутствии города
	}

	return algorithms.CompareWithAlgorithm(city1, city2, "jaro")
}

// findUniversalMatchReasons — находит причины совпадения
func findUniversalMatchReasons(nameScore, addrScore, geoScore, locScore float64) []string {
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

// AdaptiveWeights — адаптивные веса
type AdaptiveWeights struct {
	NameWeight     float64
	AddressWeight  float64
	GeoWeight      float64
	LocationWeight float64
}

// calculateAdaptiveWeights — вычисляет веса в зависимости от качества данных
func calculateAdaptiveWeights(h1, h2 domain.Hotel, cfg domain.Config) AdaptiveWeights {
	// Базовые веса из конфига
	w := AdaptiveWeights{
		NameWeight:     cfg.NameWeight,
		AddressWeight:  cfg.AddressWeight,
		GeoWeight:      cfg.GeoWeight,
		LocationWeight: cfg.LocationWeight,
	}

	// Проверяем качество данных
	hasName := h1.Name != "" && h2.Name != ""
	hasAddress := h1.Address != "" && h2.Address != ""
	hasGeo := h1.Latitude != 0 && h1.Longitude != 0 &&
		h2.Latitude != 0 && h2.Longitude != 0
	hasLocation := h1.City != "" && h2.City != "" &&
		h1.Country != "" && h2.Country != ""

	// АДАПТАЦИЯ: если поле пустое — уменьшаем его вес, перераспределяем на другие
	if !hasName {
		w.NameWeight = 0
		w.AddressWeight += 0.15
		w.GeoWeight += 0.15
		w.LocationWeight += 0.10
	}

	if !hasAddress {
		w.AddressWeight = 0
		w.NameWeight += 0.15
		w.GeoWeight += 0.15
		w.LocationWeight += 0.10
	}

	if !hasGeo {
		w.GeoWeight = 0
		w.NameWeight += 0.20
		w.AddressWeight += 0.15
		w.LocationWeight += 0.10
	}

	if !hasLocation {
		w.LocationWeight = 0
		w.NameWeight += 0.15
		w.AddressWeight += 0.15
		w.GeoWeight += 0.15
	}

	// Нормализация: сумма весов должна быть ≈ 1.0
	total := w.NameWeight + w.AddressWeight + w.GeoWeight + w.LocationWeight
	if total > 0 {
		w.NameWeight /= total
		w.AddressWeight /= total
		w.GeoWeight /= total
		w.LocationWeight /= total
	}

	return w
}

// calculateUniversalGroupConfidence — средняя попарная оценка в группе
func calculateUniversalGroupConfidence(hotels []domain.Hotel, cfg domain.Config) (float64, []string) {
	if len(hotels) <= 1 {
		return 1.0, nil
	}

	var total float64
	var count int
	reasonsSet := make(map[string]struct{})

	for i := 0; i < len(hotels); i++ {
		for j := i + 1; j < len(hotels); j++ {
			score, reasons := calculateUniversalScore(hotels[i], hotels[j], cfg)
			total += score
			count++

			for _, reason := range reasons {
				reasonsSet[reason] = struct{}{}
			}
		}
	}

	if count == 0 {
		return 1.0, nil
	}

	uniqueReasons := make([]string, 0, len(reasonsSet))
	for reason := range reasonsSet {
		uniqueReasons = append(uniqueReasons, reason)
	}

	return total / float64(count), uniqueReasons
}

// calculateUniversalConfidenceScore — итоговая уверенность
func calculateUniversalConfidenceScore(matchScore float64, hotelsCount, providersCount int) float64 {
	if hotelsCount <= 1 {
		return 1.0
	}

	// Чем больше отелей в группе, тем выше уверенность
	sizeFactor := 0.5 + float64(hotelsCount-1)*0.15
	if sizeFactor > 1.0 {
		sizeFactor = 1.0
	}

	// Чем больше разных поставщиков, тем выше уверенность
	providerFactor := 0.5 + float64(providersCount-1)*0.2
	if providerFactor > 1.0 {
		providerFactor = 1.0
	}
	if providerFactor < 0.5 {
		providerFactor = 0.5
	}

	confidence := matchScore * (0.5 + 0.25*sizeFactor + 0.25*providerFactor)
	if confidence > 1.0 {
		confidence = 1.0
	}
	if confidence < 0 {
		confidence = 0
	}
	return confidence
}
