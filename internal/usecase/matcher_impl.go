package usecase

import (
	"context"
	"fmt"
	"sync"

	"hotel-matcher/internal/domain"
	"hotel-matcher/internal/pkg/algorithms"
)

const (
	ReasonName     = "Similar names"
	ReasonAddress  = "Similar addresses"
	ReasonGeo      = "Similar geo"
	ReasonLocation = "Similar locations"
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

	result := &domain.Result{
		Groups:    make([]domain.Group, 0),
		Unmatched: make([]domain.Hotel, 0),
	}

	for groupID, hotelsInGroup := range groups {
		providersInGroup := make(map[string]bool)
		for _, h := range hotelsInGroup {
			if h.Source != "" {
				providersInGroup[h.Source] = true
			}
		}

		// Единственный проход по попарным сравнениям группы:
		// здесь же считаются confidence, reasons, PairwiseMatrix и FeatureContribution —
		// раньше это же самое (Compare*) пересчитывалось второй раз в http.buildGroupStats
		matchScore, reasons, pairwiseMatrix, featureContribution := calculateGroupConfidence(hotelsInGroup, cfg)
		score := calculateConfidenceScore(matchScore, len(hotelsInGroup), len(providersInGroup))

		result.Groups = append(result.Groups, domain.Group{
			ID:                  groupID,
			ConfidenceScore:     score,
			MatchScore:          matchScore,
			Hotels:              hotelsInGroup,
			MatchReasons:        reasons,
			PairwiseMatrix:      pairwiseMatrix,
			FeatureContribution: featureContribution,
		})
	}

	for _, hotel := range hotels {
		if !used[hotel.ID] {
			result.Unmatched = append(result.Unmatched, hotel)
		}
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
		if len(cluster) > 1 {
			mu.Lock()
			groupID := fmt.Sprintf("group-%d", len(groups))
			groups[groupID] = cluster
			mu.Unlock()
		} else {
			mu.Lock()
			used[hotels[i].ID] = false
			mu.Unlock()
		}
	}
}

// calculateFeatureScores — считает сходство пары отелей по каждому признаку
// РОВНО ОДИН РАЗ. Раньше этот же набор вызовов (Compare*) дублировался
// в http.computeFeatureScores для построения PairwiseMatrix/FeatureContribution.
func calculateFeatureScores(h1, h2 domain.Hotel, cfg domain.Config) domain.FeatureScores {
	alg := cfg.Algorithm

	nameScore := algorithms.CompareNamesWithAlgorithm(h1.Name, h2.Name, alg)
	addrScore := algorithms.CompareAddressesWithAlgorithm(h1.Address, h2.Address, alg)
	geoScore := algorithms.CompareCoordinates(h1.Latitude, h1.Longitude, h2.Latitude, h2.Longitude)
	locScore := algorithms.CompareLocationWithAlgorithm(h1.City, h1.Country, h2.City, h2.Country, alg)

	total := cfg.NameWeight*nameScore +
		cfg.AddressWeight*addrScore +
		cfg.GeoWeight*geoScore +
		cfg.LocationWeight*locScore

	return domain.FeatureScores{
		Name:     nameScore,
		Address:  addrScore,
		Geo:      geoScore,
		Location: locScore,
		Total:    total,
	}
}

// calculateMatchScore — используется в processBlock (блокировка/кластеризация),
// где PairwiseMatrix/FeatureContribution не нужны, поэтому оставлена лёгкая сигнатура
func calculateMatchScore(h1, h2 domain.Hotel, cfg domain.Config) (float64, []string) {
	fs := calculateFeatureScores(h1, h2, cfg)
	reasons := findMatchReasons(fs.Name, fs.Address, fs.Geo, fs.Location)
	return fs.Total, reasons
}

// calculateGroupConfidence — считает среднюю попарную оценку внутри группы
// и ПОПУТНО (без повторных вызовов алгоритмов) строит PairwiseMatrix и
// FeatureContribution, которые раньше пересчитывались отдельно в http-слое.
func calculateGroupConfidence(hotels []domain.Hotel, cfg domain.Config) (
	float64, []string, []domain.PairwiseSimilarity, domain.FeatureContribution,
) {
	if len(hotels) <= 1 {
		return 1.0, nil, nil, domain.FeatureContribution{}
	}

	var total float64
	var count int
	var sumName, sumAddr, sumGeo, sumLoc float64

	reasonsSet := make(map[string]struct{})
	n := len(hotels)
	matrix := make([]domain.PairwiseSimilarity, 0, n*(n-1)/2)

	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			fs := calculateFeatureScores(hotels[i], hotels[j], cfg)
			reasons := findMatchReasons(fs.Name, fs.Address, fs.Geo, fs.Location)

			total += fs.Total
			count++
			sumName += fs.Name
			sumAddr += fs.Address
			sumGeo += fs.Geo
			sumLoc += fs.Location

			matrix = append(matrix, domain.PairwiseSimilarity{
				IndexA:     i,
				IndexB:     j,
				Similarity: fs.Total,
			})

			for _, reason := range reasons {
				reasonsSet[reason] = struct{}{}
			}
		}
	}

	if count == 0 {
		return 1.0, nil, nil, domain.FeatureContribution{}
	}

	uniqueReasons := make([]string, 0, len(reasonsSet))
	for reason := range reasonsSet {
		uniqueReasons = append(uniqueReasons, reason)
	}

	contribution := domain.FeatureContribution{
		Name:    sumName / float64(count),
		Address: sumAddr / float64(count),
		Geo:     sumGeo / float64(count),
		City:    sumLoc / float64(count),
	}

	return total / float64(count), uniqueReasons, matrix, contribution
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
