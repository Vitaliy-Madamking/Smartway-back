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
	Groups    []GroupDTO `json:"groups"`
	Unmatched []HotelDTO `json:"unmatched"`
}

type GroupDTO struct {
	GroupID         string     `json:"groupId"`
	ConfidenceScore float64    `json:"confidenceScore"`
	Hotels          []HotelDTO `json:"hotels"`
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
		Groups:    make([]GroupDTO, len(result.Groups)),
		Unmatched: make([]HotelDTO, len(result.Unmatched)),
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
			ConfidenceScore: g.ConfidenceScore,
			Hotels:          hotelsDTO,
		}
	}
	for i, h := range result.Unmatched {
		resp.Unmatched[i] = HotelDTO{
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
	return resp
}