package http

import "hotel-matcher/internal/domain"

type MatchRequest struct {
	Hotels []HotelDTO `json:"hotels"`
	Config ConfigDTO  `json:"config,omitempty"`
}

type HotelDTO struct {
	ID        string  `json:"id"`
	Source    string  `json:"source"`
	Name      string  `json:"name"`
	Address   string  `json:"address"`
	City      string  `json:"city"`
	Country   string  `json:"country"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type ConfigDTO struct {
	NameWeight     *float64 `json:"nameWeight,omitempty"`
	AddressWeight  *float64 `json:"addressWeight,omitempty"`
	GeoWeight      *float64 `json:"geoWeight,omitempty"`
	LocationWeight *float64 `json:"locationWeight,omitempty"`
	Threshold      *float64 `json:"threshold,omitempty"`
	Algorithm      *string  `json:"algorithm,omitempty"` // "jaro-winkler", "jaro", "levenshtein", "cosine"
}

type MatchResponse struct {
	Metrics MetricsDTO `json:"metrics"`
	Groups  []GroupDTO `json:"groups"`
}

type MetricsDTO struct {
	TotalHotels       int     `json:"totalHotels"`
	TotalGroups       int     `json:"totalGroups"`
	TotalDuplicates   int     `json:"totalDuplicates"`
	TotalProviders    int     `json:"totalProviders"`
	AverageConfidence float64 `json:"averageConfidence"`
}

type GroupDTO struct {
	GroupID         string          `json:"groupId"`
	PrimaryName     string          `json:"primaryName"`
	MatchScore      float64         `json:"matchScore"`
	ConfidenceScore float64         `json:"confidenceScore"`
	ProvidersCount  int             `json:"providersCount"`
	HotelsCount     int             `json:"hotelsCount"`
	MatchReasons    MatchReasonsDTO `json:"matchReasons"`
	Hotels          []HotelDTO      `json:"hotels"`
}

type MatchReasonsDTO struct {
	MatchedSuppliers []string `json:"matched_suppliers"`
	Total            int      `json:"total"`
}

func (r MatchRequest) ToDomain() ([]domain.Hotel, domain.Config) {
	hotels := make([]domain.Hotel, len(r.Hotels))
	for i, h := range r.Hotels {
		hotels[i] = domain.Hotel{
			ID:        h.ID,
			Source:    h.Source,
			Name:      h.Name,
			Address:   h.Address,
			City:      h.City,
			Country:   h.Country,
			Latitude:  h.Latitude,
			Longitude: h.Longitude,
		}
	}
	cfg := domain.DefaultConfig()
	if r.Config.NameWeight != nil {
		cfg.NameWeight = *r.Config.NameWeight
	}
	if r.Config.AddressWeight != nil {
		cfg.AddressWeight = *r.Config.AddressWeight
	}
	if r.Config.GeoWeight != nil {
		cfg.GeoWeight = *r.Config.GeoWeight
	}
	if r.Config.LocationWeight != nil {
		cfg.LocationWeight = *r.Config.LocationWeight
	}
	if r.Config.Threshold != nil {
		cfg.Threshold = *r.Config.Threshold
	}
	if r.Config.Algorithm != nil {
		cfg.Algorithm = *r.Config.Algorithm
	}
	return hotels, cfg
}

func ToDTO(result *domain.Result) MatchResponse {
	resp := MatchResponse{
		Groups: make([]GroupDTO, len(result.Groups)),
		Metrics: MetricsDTO{
			TotalHotels:       result.Metrics.TotalHotels,
			TotalGroups:       result.Metrics.TotalGroups,
			TotalDuplicates:   result.Metrics.TotalDuplicates,
			TotalProviders:    result.Metrics.TotalProviders,
			AverageConfidence: result.Metrics.AverageConfidence,
		},
	}

	for i, g := range result.Groups {
		hotelsDTO := make([]HotelDTO, len(g.Hotels))
		for j, h := range g.Hotels {
			hotelsDTO[j] = HotelDTO{
				ID:        h.ID,
				Source:    h.Source,
				Name:      h.Name,
				Address:   h.Address,
				City:      h.City,
				Country:   h.Country,
				Latitude:  h.Latitude,
				Longitude: h.Longitude,
			}
		}
		resp.Groups[i] = GroupDTO{
			GroupID:         g.ID,
			PrimaryName:     g.PrimaryName,
			MatchScore:      g.MatchScore,
			ConfidenceScore: g.ConfidenceScore,
			ProvidersCount:  g.ProvidersCount,
			HotelsCount:     g.HotelsCount,
			MatchReasons: MatchReasonsDTO{
				MatchedSuppliers: g.MatchReasons.MatchedSuppliers,
				Total:            g.MatchReasons.Total,
			},
			Hotels: hotelsDTO,
		}
	}

	return resp
}
