package domain

import "time"

// Group — группа совпавших отелей (таблица hotel_groups)
type Group struct {
	ID              int64     `json:"id"`
	PrimaryName     string    `json:"primary_name"`
	ConfidenceScore float64   `json:"confidence"`
	MatchScore      float64   `json:"match_score"`
	ProvidersCount  int       `json:"providers_count"`
	HotelsCount     int       `json:"hotels_count"`
	MatchReasons    []byte    `json:"match_reasons,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	Hotels          []Hotel   `json:"hotels"`
}

// Result — результат матчинга
type Result struct {
	Groups    []Group `json:"groups"`
	Unmatched []Hotel `json:"unmatched"`
}
