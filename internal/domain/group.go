package domain

// FeatureScores — разложение итогового скора совпадения по признакам
// для пары отелей. Считается один раз в usecase и переиспользуется
// и для MatchReasons, и для агрегированной статистики группы.
type FeatureScores struct {
	Name     float64
	Address  float64
	Geo      float64
	Location float64
	Total    float64
}

// PairwiseSimilarity — попарное сходство двух отелей внутри группы.
// IndexA/IndexB — индексы в срезе Group.Hotels.
type PairwiseSimilarity struct {
	IndexA     int
	IndexB     int
	Similarity float64
}

// FeatureContribution — средний вклад каждого признака в сходство внутри группы
type FeatureContribution struct {
	Name    float64
	Address float64
	Geo     float64
	City    float64
}

// Group — группа сопоставленных отелей
type Group struct {
	ID                  string
	ConfidenceScore     float64
	MatchScore          float64
	Hotels              []Hotel
	MatchReasons        []string
	PairwiseMatrix      []PairwiseSimilarity // посчитано один раз в usecase, без повторного вызова алгоритмов в http-слое
	FeatureContribution FeatureContribution
}

// Result — итоговый ответ матчинга
type Result struct {
	Groups    []Group
	Unmatched []Hotel
}
