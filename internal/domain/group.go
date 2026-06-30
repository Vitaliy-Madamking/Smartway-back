package domain

// Group — группа сопоставленных отелей (или одиночный отель, если совпадений не нашлось)
type Group struct {
	ID string

	// PrimaryName — "каноническое" имя группы (самое частое имя среди отелей группы)
	PrimaryName string

	// MatchScore — средняя попарная оценка совпадения строк/координат внутри группы (0..1)
	// Чистый результат алгоритма сравнения, без поправок на размер группы и т.д.
	MatchScore float64

	// ConfidenceScore — итоговая уверенность в том, что группа сформирована верно.
	// Учитывает MatchScore, количество отелей в группе и количество разных поставщиков.
	ConfidenceScore float64

	ProvidersCount int
	HotelsCount    int

	MatchReasons MatchReasons

	Hotels []Hotel
}

type MatchReasons struct {
	MatchedSuppliers []string
	Total            int
}

// Result — итоговый ответ матчинга.
// Unmatched больше не используется как отдельный список — каждый отель,
// который ни с кем не совпал, превращается в группу из одного элемента.
type Result struct {
	Metrics ResultMetrics
	Groups  []Group
}

type ResultMetrics struct {
	TotalHotels       int     // всего отелей на входе
	TotalGroups       int     // всего получившихся групп (включая одиночные)
	TotalDuplicates   int     // сколько отелей "схлопнулось" в дубликаты (HotelsCount-1 по каждой группе)
	TotalProviders    int     // количество уникальных поставщиков (Source) среди всех отелей
	AverageConfidence float64 // средняя ConfidenceScore по всем группам
}
