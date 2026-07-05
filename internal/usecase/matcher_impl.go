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
		matchScore, reasons := calculateGroupConfidence(hotelsInGroup, cfg)
		score := calculateConfidenceScore(matchScore, len(hotelsInGroup), len(providersInGroup))

		
		// ВЫЧИСЛЯЕМ featureContribution и pairwiseMatrix
		
		var pairwiseMatrix []domain.PairwiseSimilarity
		var featureContribution domain.FeatureContribution

		if len(hotelsInGroup) >= 2 {
			pairwiseMatrix = calculatePairwiseMatrix(hotelsInGroup, cfg)
			featureContribution = calculateFeatureContribution(hotelsInGroup, cfg)
		}

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

func calculateMatchScore(h1, h2 domain.Hotel, cfg domain.Config) (float64, []string) {
	alg := cfg.Algorithm

	var nameScore, addrScore, geoScore, locScore float64

	if alg == "universal" {
		nameScore = algorithms.CompareNamesWithAlgorithm(h1.Name, h2.Name, "jaro-winkler")
		addrScore = algorithms.CompareAddressesWithAlgorithm(h1.Address, h2.Address, "levenshtein")
		geoScore = algorithms.CompareCoordinates(h1.Latitude, h1.Longitude, h2.Latitude, h2.Longitude)
		locScore = algorithms.CompareLocationWithAlgorithm(h1.City, h1.Country, h2.City, h2.Country, "jaro")
	} else {
		nameScore = algorithms.CompareNamesWithAlgorithm(h1.Name, h2.Name, alg)
		addrScore = algorithms.CompareAddressesWithAlgorithm(h1.Address, h2.Address, alg)
		geoScore = algorithms.CompareCoordinates(h1.Latitude, h1.Longitude, h2.Latitude, h2.Longitude)
		locScore = algorithms.CompareLocationWithAlgorithm(h1.City, h1.Country, h2.City, h2.Country, alg)
	}

	reasons := findMatchReasons(nameScore, addrScore, geoScore, locScore)

	return cfg.NameWeight*nameScore +
			cfg.AddressWeight*addrScore +
			cfg.GeoWeight*geoScore +
			cfg.LocationWeight*locScore,
		reasons
}

func calculateGroupConfidence(hotels []domain.Hotel, cfg domain.Config) (float64, []string) {
	if len(hotels) <= 1 {
		return 1.0, nil
	}

	var total float64
	var count int
	reasonsSet := make(map[string]struct{})

	for i := 0; i < len(hotels); i++ {
		for j := i + 1; j < len(hotels); j++ {
			rating, reasons := calculateMatchScore(hotels[i], hotels[j], cfg)
			total += rating
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


// НОВЫЕ ФУНКЦИИ

func calculatePairwiseMatrix(hotels []domain.Hotel, cfg domain.Config) []domain.PairwiseSimilarity {
	if len(hotels) < 2 {
		return nil
	}

	matrix := make([]domain.PairwiseSimilarity, 0)
	for i := 0; i < len(hotels); i++ {
		for j := i + 1; j < len(hotels); j++ {
			score, _ := calculateMatchScore(hotels[i], hotels[j], cfg)
			matrix = append(matrix, domain.PairwiseSimilarity{
				IndexA:     i,
				IndexB:     j,
				Similarity: score,
			})
		}
	}
	return matrix
}

func calculateFeatureContribution(hotels []domain.Hotel, cfg domain.Config) domain.FeatureContribution {
	if len(hotels) < 2 {
		return domain.FeatureContribution{}
	}

	var totalName, totalAddr, totalGeo, totalCity float64
	var count int

	alg := cfg.Algorithm

	for i := 0; i < len(hotels); i++ {
		for j := i + 1; j < len(hotels); j++ {
			var nameScore, addrScore, geoScore, locScore float64

			if alg == "universal" {
				nameScore = algorithms.CompareNamesWithAlgorithm(hotels[i].Name, hotels[j].Name, "jaro-winkler")
				addrScore = algorithms.CompareAddressesWithAlgorithm(hotels[i].Address, hotels[j].Address, "levenshtein")
				geoScore = algorithms.CompareCoordinates(hotels[i].Latitude, hotels[i].Longitude, hotels[j].Latitude, hotels[j].Longitude)
				locScore = algorithms.CompareLocationWithAlgorithm(hotels[i].City, hotels[i].Country, hotels[j].City, hotels[j].Country, "jaro")
			} else {
				nameScore = algorithms.CompareNamesWithAlgorithm(hotels[i].Name, hotels[j].Name, alg)
				addrScore = algorithms.CompareAddressesWithAlgorithm(hotels[i].Address, hotels[j].Address, alg)
				geoScore = algorithms.CompareCoordinates(hotels[i].Latitude, hotels[i].Longitude, hotels[j].Latitude, hotels[j].Longitude)
				locScore = algorithms.CompareLocationWithAlgorithm(hotels[i].City, hotels[i].Country, hotels[j].City, hotels[j].Country, alg)
			}

			totalName += nameScore
			totalAddr += addrScore
			totalGeo += geoScore
			totalCity += locScore
			count++
		}
	}

	if count == 0 {
		return domain.FeatureContribution{}
	}

	return domain.FeatureContribution{
		Name:    totalName / float64(count),
		Address: totalAddr / float64(count),
		Geo:     totalGeo / float64(count),
		City:    totalCity / float64(count),
	}
}