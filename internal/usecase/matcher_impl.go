package usecase

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"hotel-matcher/internal/domain"
	"hotel-matcher/internal/pkg/algorithms"
)

type matcherImpl struct {
	repo HotelRepository
}

func NewMatcher(repo HotelRepository) Matcher {
	return &matcherImpl{repo: repo}
}

func (m *matcherImpl) Match(ctx context.Context, hotels []domain.Hotel, cfg domain.Config) (*domain.Result, error) {
	if len(hotels) == 0 {
		return nil, domain.ErrNoHotels
	}
	if cfg.Threshold < 0 || cfg.Threshold > 1 {
		return nil, domain.ErrInvalidConfig
	}

	blocks := buildBlocks(hotels)

	var mu sync.Mutex
	groups := make(map[string][]domain.Hotel)
	used := make(map[string]bool)
	var wg sync.WaitGroup

	for _, block := range blocks {
		wg.Add(1)
		go func(h []domain.Hotel) {
			defer wg.Done()
			m.processBlock(h, cfg, &mu, groups, used)
		}(block)
	}
	wg.Wait()

	// Любой отель, который не попал ни в один кластер (например, единственный
	// в своём блоке, либо ни с кем не совпавший внутри блока), становится
	// группой из одного отеля.
	for _, hotel := range hotels {
		if !used[hotel.ID] {
			groupID := fmt.Sprintf("group-%d", len(groups))
			groups[groupID] = []domain.Hotel{hotel}
			used[hotel.ID] = true
		}
	}

	result := &domain.Result{
		Groups: make([]domain.Group, 0, len(groups)),
	}

	var confidenceSum float64
	var totalDuplicates int
	allProviders := make(map[string]struct{})

	for groupID, hotelsInGroup := range groups {
		matchScore := calculateMatchScoreAvg(hotelsInGroup, cfg)
		providers := distinctSuppliers(hotelsInGroup)
		confidence := calculateConfidenceScore(matchScore, len(hotelsInGroup), len(providers))

		group := domain.Group{
			ID:              groupID,
			PrimaryName:     pickPrimaryName(hotelsInGroup),
			MatchScore:      matchScore,
			ConfidenceScore: confidence,
			ProvidersCount:  len(providers),
			HotelsCount:     len(hotelsInGroup),
			MatchReasons: domain.MatchReasons{
				MatchedSuppliers: providers,
				Total:            len(hotelsInGroup),
			},
			Hotels: hotelsInGroup,
		}
		result.Groups = append(result.Groups, group)

		confidenceSum += confidence
		if len(hotelsInGroup) > 1 {
			totalDuplicates += len(hotelsInGroup) - 1
		}
		for _, h := range hotelsInGroup {
			if h.Source != "" {
				allProviders[h.Source] = struct{}{}
			}
		}
	}

	avgConfidence := 0.0
	if len(result.Groups) > 0 {
		avgConfidence = confidenceSum / float64(len(result.Groups))
	}

	result.Metrics = domain.ResultMetrics{
		TotalHotels:       len(hotels),
		TotalGroups:       len(result.Groups),
		TotalDuplicates:   totalDuplicates,
		TotalProviders:    len(allProviders),
		AverageConfidence: avgConfidence,
	}

	return result, nil
}

func buildBlocks(hotels []domain.Hotel) map[string][]domain.Hotel {
	blocks := make(map[string][]domain.Hotel)
	for _, h := range hotels {
		key := fmt.Sprintf("%s|%s", h.Country, h.City)
		blocks[key] = append(blocks[key], h)
	}
	return blocks
}

// processBlock больше не "развязывает" одиночные отели обратно — если отель
// ни с кем не совпал, он просто остаётся неиспользованным, и потом
// превратится в свою собственную группу в Match().
func (m *matcherImpl) processBlock(hotels []domain.Hotel, cfg domain.Config, mu *sync.Mutex,
	groups map[string][]domain.Hotel, used map[string]bool) {

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

			score := calculateMatchScore(hotels[i], hotels[j], cfg)
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
			groupID := fmt.Sprintf("group-%d", len(groups))
			groups[groupID] = cluster
			mu.Unlock()
		}
		// Если cluster содержит только hotels[i] — ничего не делаем,
		// used[hotels[i].ID] остаётся true, а группу из одного отеля
		// создаст финальный проход в Match().
	}
}

// calculateMatchScore — попарная оценка совпадения двух отелей (как раньше).
func calculateMatchScore(h1, h2 domain.Hotel, cfg domain.Config) float64 {
	alg := cfg.Algorithm

	nameScore := algorithms.CompareNamesWithAlgorithm(h1.Name, h2.Name, alg)
	addrScore := algorithms.CompareAddressesWithAlgorithm(h1.Address, h2.Address, alg)
	geoScore := algorithms.CompareCoordinates(h1.Latitude, h1.Longitude, h2.Latitude, h2.Longitude)
	locScore := algorithms.CompareLocationWithAlgorithm(h1.City, h1.Country, h2.City, h2.Country, alg)

	return cfg.NameWeight*nameScore +
		cfg.AddressWeight*addrScore +
		cfg.GeoWeight*geoScore +
		cfg.LocationWeight*locScore
}

// calculateMatchScoreAvg — средняя попарная оценка совпадения внутри группы.
// Это "сырой" MatchScore, без поправок на размер группы/кол-во поставщиков.
// Для одиночной группы (1 отель) совпадать не с кем — считаем как 1.0
// (отель полностью "совпадает сам с собой").
func calculateMatchScoreAvg(hotels []domain.Hotel, cfg domain.Config) float64 {
	if len(hotels) <= 1 {
		return 1.0
	}
	var total float64
	var count int
	for i := 0; i < len(hotels); i++ {
		for j := i + 1; j < len(hotels); j++ {
			total += calculateMatchScore(hotels[i], hotels[j], cfg)
			count++
		}
	}
	if count == 0 {
		return 1.0
	}
	return total / float64(count)
}

// calculateConfidenceScore — итоговая уверенность алгоритма.
// В отличие от MatchScore (чистое сходство строк/координат), здесь учитывается:
//   - sizeFactor: чем больше отелей в группе подтвердили совпадение, тем выше уверенность
//   - providerFactor: чем больше РАЗНЫХ поставщиков подтвердили этот отель, тем выше уверенность
//     (если 5 записей, но все от одного поставщика — это слабее, чем 3 записи от 3 разных)
//
// Для одиночных групп (отель не нашёл пару) уверенность снижена, т.к. это
// не подтверждённый дубликат, а скорее "нет повторов" — что само по себе не ошибка,
// но и не "уверенное совпадение".
func calculateConfidenceScore(matchScore float64, hotelsCount, providersCount int) float64 {
	if hotelsCount <= 1 {
		return 1.0
	}

	sizeFactor := 0.5 + float64(hotelsCount-1)*0.15
	if sizeFactor > 1.0 {
		sizeFactor = 1.0
	}

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

// pickPrimaryName — самое частое имя в группе (мода), при равенстве — первое встреченное.
func pickPrimaryName(hotels []domain.Hotel) string {
	if len(hotels) == 0 {
		return ""
	}
	counts := make(map[string]int)
	order := make([]string, 0, len(hotels))
	for _, h := range hotels {
		if counts[h.Name] == 0 {
			order = append(order, h.Name)
		}
		counts[h.Name]++
	}
	best := order[0]
	for _, name := range order {
		if counts[name] > counts[best] {
			best = name
		}
	}
	return best
}

// distinctSuppliers — отсортированный список уникальных Source среди отелей группы.
func distinctSuppliers(hotels []domain.Hotel) []string {
	set := make(map[string]struct{})
	for _, h := range hotels {
		if h.Source != "" {
			set[h.Source] = struct{}{}
		}
	}
	result := make([]string, 0, len(set))
	for s := range set {
		result = append(result, s)
	}
	sort.Strings(result)
	return result
}
