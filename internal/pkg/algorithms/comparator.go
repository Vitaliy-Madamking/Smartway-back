package algorithms

import (
	"strings"

	"github.com/hbollon/go-edlib"
)

// АЛГОРИТМЫ СРАВНЕНИЯ СТРОК (полный набор)
//
// БИБЛИОТЕЧНЫЕ АЛГОРИТМЫ (из go-edlib):
// 1. jaro-winkler    - Лучший для названий (учитывает префиксный бонус)
// 2. jaro            - Для городов (быстрый, учитывает перестановки)
// 3. levenshtein     - Для адресов (учитывает вставки/удаления)
// 4. cosine          - Для длинных текстов (учитывает общие подстроки)
//
// КАСТОМНЫЕ АЛГОРИТМЫ (собственные реализации):
// 5. damerau-levenshtein - Учитывает транспозиции (перестановки соседних)
// 6. lcs              - Наибольшая общая подпоследовательность
// 7. soundex          - Фонетический (для английских имён)
// 8. ngram            - N-граммы (для длинных строк)

// CompareWithAlgorithm - универсальная функция сравнения с выбором алгоритма
// Поддерживает: "jaro-winkler", "jaro", "levenshtein", "cosine",
//               "damerau-levenshtein", "lcs", "soundex", "ngram"
func CompareWithAlgorithm(a, b, algorithm string) float64 {
	a = NormalizeString(a)
	b = NormalizeString(b)

	if a == "" && b == "" {
		return 1.0
	}
	if a == "" || b == "" {
		return 0.0
	}
	if a == b {
		return 1.0
	}

	switch algorithm {

	// БИБЛИОТЕЧНЫЕ АЛГОРИТМЫ (из go-edlib)
	

	case "jaro-winkler":
		// Jaro-Winkler - ЛУЧШИЙ для названий отелей
		// Преимущество: префиксный бонус, устойчив к опечаткам
		sim := edlib.JaroWinklerSimilarity(a, b)
		return float64(sim)

	case "jaro":
		// Jaro - для городов и коротких слов
		// Преимущество: быстрый, учитывает перестановки
		sim, _ := edlib.StringsSimilarity(a, b, edlib.Jaro)
		return float64(sim)

	case "levenshtein":
		// Levenshtein - ЛУЧШИЙ для адресов
		// Преимущество: учитывает вставки/удаления/замены
		sim, _ := edlib.StringsSimilarity(a, b, edlib.Levenshtein)
		return float64(sim)

	case "cosine":
		// Cosine - для длинных текстов
		// Преимущество: учитывает общие подстроки, не зависит от длины
		sim, _ := edlib.StringsSimilarity(a, b, edlib.Cosine)
		return float64(sim)

	// КАСТОМНЫЕ АЛГОРИТМЫ (собственные реализации)
	
	case "damerau-levenshtein":
		// Damerau-Levenshtein - для сложных опечаток
		// Преимущество: учитывает перестановки соседних символов
		// Пример: "hte" → "the" (транспозиция)
		return DamerauLevenshteinSimilarity(a, b)

	case "lcs":
		// LCS - Наибольшая общая подпоследовательность
		// Преимущество: учитывает порядок символов
		// Пример: "ABCD" и "ACBD" → LCS = 3 ("ABD")
		lcsLen := LCS(a, b)
		maxLen := max2(len(a), len(b))
		if maxLen == 0 {
			return 1.0
		}
		return float64(lcsLen) / float64(maxLen)

	case "soundex":
		// Soundex - фонетический алгоритм
		// Преимущество: сравнивает по произношению
		// Пример: "Smith" и "Smyth" → одинаковый код "S530"
		if Soundex(a) == Soundex(b) {
			return 1.0
		}
		return 0.0

	case "ngram":
		// N-gram - разбивает на фрагменты по N символов
		// Преимущество: устойчив к перестановкам слов
		// Пример: "hello world" → ["he","el","ll","lo",...]
		return NGramSimilarity(a, b, 2)

	default:
		// По умолчанию Jaro-Winkler (лучший для общего случая)
		sim := edlib.JaroWinklerSimilarity(a, b)
		return float64(sim)
	}
}


// СПЕЦИАЛИЗИРОВАННЫЕ ФУНКЦИИ ДЛЯ КОНКРЕТНЫХ ПОЛЕЙ

// CompareNames - сравнение названий отелей (по умолчанию Jaro-Winkler)
func CompareNames(a, b string) float64 {
	return CompareNamesWithAlgorithm(a, b, "jaro-winkler")
}

// CompareNamesWithAlgorithm - сравнение названий с выбором алгоритма
func CompareNamesWithAlgorithm(a, b, algorithm string) float64 {
	a = NormalizeName(a)
	b = NormalizeName(b)
	if a == "" && b == "" {
		return 1.0
	}
	if a == "" || b == "" {
		return 0.0
	}
	if a == b {
		return 1.0
	}
	if strings.Contains(a, b) || strings.Contains(b, a) {
		return 0.95
	}
	return CompareWithAlgorithm(a, b, algorithm)
}

// CompareAddresses - сравнение адресов (по умолчанию Levenshtein)
func CompareAddresses(a, b string) float64 {
	return CompareAddressesWithAlgorithm(a, b, "levenshtein")
}

// CompareAddressesWithAlgorithm - сравнение адресов с выбором алгоритма
func CompareAddressesWithAlgorithm(a, b, algorithm string) float64 {
	a = NormalizeAddress(a)
	b = NormalizeAddress(b)
	if a == "" && b == "" {
		return 1.0
	}
	if a == "" || b == "" {
		return 0.0
	}
	if a == b {
		return 1.0
	}
	if strings.Contains(a, b) || strings.Contains(b, a) {
		return 0.95
	}
	// Для адресов лучше всего Levenshtein или Damerau-Levenshtein
	if algorithm == "default" || algorithm == "" {
		algorithm = "levenshtein"
	}
	return CompareWithAlgorithm(a, b, algorithm)
}

// CompareLocation - сравнение города и страны (по умолчанию Jaro)
func CompareLocation(city1, country1, city2, country2 string) float64 {
	return CompareLocationWithAlgorithm(city1, country1, city2, country2, "jaro")
}

// CompareLocationWithAlgorithm - сравнение города и страны с выбором алгоритма
func CompareLocationWithAlgorithm(city1, country1, city2, country2, algorithm string) float64 {
	// Страна - точное совпадение (не допускаем ошибок)
	if country1 != country2 {
		return 0.0
	}
	c1 := NormalizeString(city1)
	c2 := NormalizeString(city2)
	if c1 == "" && c2 == "" {
		return 1.0
	}
	if c1 == "" || c2 == "" {
		return 0.5
	}
	if c1 == c2 {
		return 1.0
	}
	// Для городов лучше всего Jaro или Jaro-Winkler
	if algorithm == "default" || algorithm == "" {
		algorithm = "jaro"
	}
	return CompareWithAlgorithm(c1, c2, algorithm)
}