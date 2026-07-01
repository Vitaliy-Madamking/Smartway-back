package domain

// Group — группа сопоставленных отелей
type Group struct {
	ID              string
	ConfidenceScore float64
	Hotels          []Hotel
	MatchReasons    []string
}

// Result — итоговый ответ матчинга
type Result struct {
	Groups    []Group
	Unmatched []Hotel
}
