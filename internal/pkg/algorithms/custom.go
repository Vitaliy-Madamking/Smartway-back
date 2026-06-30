package algorithms

import (
	"math"
	"strings"
)

// ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ

// min2 возвращает минимальное из двух целых чисел
func min2(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// min3 возвращает минимальное из трёх целых чисел
func min3(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

// max2 возвращает максимальное из двух целых чисел
func max2(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// 1. LCS - Longest Common Subsequence (Наибольшая общая подпоследовательность)
// Находит длину наибольшей последовательности символов, которая встречается в обеих строках в одном порядке
// Пример: LCS("ABCD", "ACBD") = 3 ("ABD" или "ACD")
// Сложность: O(n*m) по времени и памяти
func LCS(a, b string) int {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	// Матрица динамического программирования
	dp := make([][]int, len(a)+1)
	for i := range dp {
		dp[i] = make([]int, len(b)+1)
	}
	// Заполняем матрицу снизу вверх
	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			if a[i-1] == b[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1 // символы совпали
			} else {
				dp[i][j] = max2(dp[i-1][j], dp[i][j-1]) // берём максимум
			}
		}
	}
	return dp[len(a)][len(b)]
}

// 2. Levenshtein Distance (расстояние Левенштейна) - кастомная версия
// Считает минимальное количество операций для превращения строки a в строку b
// Операции: вставка, удаление, замена символа
// Пример: Levenshtein("kitten", "sitting") = 3
// Сложность: O(n*m) по времени, O(min(n,m)) по памяти (оптимизированная версия)
func LevenshteinDistance(a, b string) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}
	// Используем две строки вместо матрицы для экономии памяти
	prev := make([]int, len(b)+1)
	curr := make([]int, len(b)+1)
	// Инициализация: расстояние от пустой строки до b[:j] = j
	for j := 0; j <= len(b); j++ {
		prev[j] = j
	}
	// Основной цикл по символам строки a
	for i := 1; i <= len(a); i++ {
		curr[0] = i // расстояние от a[:i] до пустой строки = i
		for j := 1; j <= len(b); j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1 // замена
			}
			// Минимум из трёх операций: удаление, вставка, замена
			curr[j] = min3(
				prev[j]+1,      // удаление из a
				curr[j-1]+1,    // вставка в a
				prev[j-1]+cost, // замена или совпадение
			)
		}
		prev, curr = curr, prev // меняем местами для следующей итерации
	}
	return prev[len(b)]
}

// LevenshteinSimilarity нормализует расстояние в процент схожести (0..1)
func LevenshteinSimilarity(a, b string) float64 {
	dist := LevenshteinDistance(a, b)
	maxLen := max2(len(a), len(b))
	if maxLen == 0 {
		return 1.0
	}
	return 1.0 - float64(dist)/float64(maxLen)
}

// 3. Damerau-Levenshtein Distance (с учётом транспозиций)
// Расширение Levenshtein, которое учитывает перестановки соседних символов
// Пример: DamerauLevenshtein("ca", "ac") = 1 (транспозиция)
// Сложность: O(n*m) по времени и памяти
func DamerauLevenshteinDistance(a, b string) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}
	dp := make([][]int, len(a)+1)
	for i := range dp {
		dp[i] = make([]int, len(b)+1)
	}
	// Инициализация матрицы
	for i := 0; i <= len(a); i++ {
		dp[i][0] = i
	}
	for j := 0; j <= len(b); j++ {
		dp[0][j] = j
	}
	// Заполняем матрицу
	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}
			// Стандартные операции Levenshtein
			dp[i][j] = min3(
				dp[i-1][j]+1,      // удаление
				dp[i][j-1]+1,      // вставка
				dp[i-1][j-1]+cost, // замена
			)
			// Дополнительно: транспозиция (перестановка соседних символов)
			if i > 1 && j > 1 && a[i-1] == b[j-2] && a[i-2] == b[j-1] {
				dp[i][j] = min2(dp[i][j], dp[i-2][j-2]+1)
			}
		}
	}
	return dp[len(a)][len(b)]
}

// DamerauLevenshteinSimilarity нормализует расстояние в процент схожести (0..1)
func DamerauLevenshteinSimilarity(a, b string) float64 {
	dist := DamerauLevenshteinDistance(a, b)
	maxLen := max2(len(a), len(b))
	if maxLen == 0 {
		return 1.0
	}
	return 1.0 - float64(dist)/float64(maxLen)
}

// 4. Jaro Similarity (кастомная версия)
// Учитывает общие символы и их порядок, но без префиксного бонуса
// Используется для сравнения коротких строк (города, имена)
// Диапазон: 0.0 - 1.0, где 1.0 = полностью совпадают
func JaroSimilarity(a, b string) float64 {
	if a == "" && b == "" {
		return 1.0
	}
	if a == "" || b == "" {
		return 0.0
	}
	a = NormalizeString(a)
	b = NormalizeString(b)
	if a == b {
		return 1.0
	}
	// Максимальное расстояние для поиска совпадающих символов
	matchDistance := max2(len(a), len(b))/2 - 1
	if matchDistance < 0 {
		matchDistance = 0
	}
	// Массивы для отметки совпавших символов
	matchesA := make([]bool, len(a))
	matchesB := make([]bool, len(b))
	matches := 0
	// ШАГ 1: Находим совпадающие символы в пределах matchDistance
	for i := 0; i < len(a); i++ {
		start := max2(0, i-matchDistance)
		end := min2(len(b), i+matchDistance+1)
		for j := start; j < end; j++ {
			if matchesB[j] {
				continue
			}
			if a[i] != b[j] {
				continue
			}
			matchesA[i] = true
			matchesB[j] = true
			matches++
			break
		}
	}
	if matches == 0 {
		return 0.0
	}
	// ШАГ 2: Считаем транспозиции (перестановки)
	transpositions := 0
	j := 0
	for i := 0; i < len(a); i++ {
		if !matchesA[i] {
			continue
		}
		for !matchesB[j] {
			j++
		}
		if a[i] != b[j] {
			transpositions++
		}
		j++
	}
	transpositions /= 2
	// ШАГ 3: Формула Jaro
	return (float64(matches)/float64(len(a)) +
		float64(matches)/float64(len(b)) +
		float64(matches-transpositions)/float64(matches)) / 3.0
}

// 5. Jaro-Winkler Similarity (кастомная версия)
// Jaro + префиксный бонус (учитывает совпадение первых символов)
// Используется для названий отелей, брендов, имён
// Даёт бонус, если начало строк совпадает
func JaroWinklerSimilarityCustom(a, b string) float64 {
	jaro := JaroSimilarity(a, b)
	// Считаем длину общего префикса (максимум 4 символа)
	prefixLength := 0
	maxPrefix := min2(4, min2(len(a), len(b)))
	for i := 0; i < maxPrefix; i++ {
		if a[i] == b[i] {
			prefixLength++
		} else {
			break
		}
	}
	// Добавляем бонус: prefixLength * 0.1 * (1 - jaro)
	return jaro + float64(prefixLength)*0.1*(1-jaro)
}

// 6. Soundex (фонетический алгоритм)
// Преобразует слово в код из 4 символов (буква + 3 цифры)
// Сравнивает слова по произношению, а не написанию
// Пример: "Smith" → S530, "Smyth" → S530 (одинаковый код)
// Используется только для английского языка
func Soundex(s string) string {
	if s == "" {
		return ""
	}
	s = strings.ToUpper(s)
	result := string(s[0]) // первая буква всегда сохраняется
	// Маппинг букв в цифры по правилам Soundex
	mapping := map[rune]string{
		'B': "1", 'P': "1", 'F': "1", 'V': "1",
		'C': "2", 'G': "2", 'J': "2", 'K': "2", 'Q': "2", 'S': "2", 'X': "2", 'Z': "2",
		'D': "3", 'T': "3",
		'L': "4",
		'M': "5", 'N': "5",
		'R': "6",
	}
	prevCode := ""
	// Проходим по остальным буквам (максимум 4 символа в коде)
	for i := 1; i < len(s) && len(result) < 4; i++ {
		char := rune(s[i])
		code, exists := mapping[char]
		if !exists {
			continue // гласные и H, W игнорируются
		}
		if code != prevCode {
			result += code // добавляем только если код изменился
			prevCode = code
		}
	}
	// Дополняем нулями до 4 символов
	for len(result) < 4 {
		result += "0"
	}
	return result
}

// SoundexSimilarity возвращает 1 если Soundex коды совпадают, иначе 0
func SoundexSimilarity(a, b string) float64 {
	if Soundex(a) == Soundex(b) {
		return 1.0
	}
	return 0.0
}

// 7. N-Gram Similarity (сравнение по N-граммам)
// Разбивает строки на фрагменты по N символов и сравнивает множества
// Пример: N=2, "hello" → {"he","el","ll","lo"}
// Используется для сравнения длинных строк, устойчив к перестановкам
func NGramSimilarity(a, b string, n int) float64 {
	if a == "" && b == "" {
		return 1.0
	}
	if a == "" || b == "" {
		return 0.0
	}
	if a == b {
		return 1.0
	}
	ngramsA := getNgrams(a, n)
	ngramsB := getNgrams(b, n)
	if len(ngramsA) == 0 || len(ngramsB) == 0 {
		return 0.0
	}
	// Считаем пересечение (общие N-граммы)
	intersection := 0
	for ngram := range ngramsA {
		if ngramsB[ngram] {
			intersection++
		}
	}
	// Коэффициент Жаккара: |A ∩ B| / max(|A|, |B|)
	return float64(intersection) / float64(max2(len(ngramsA), len(ngramsB)))
}

// getNgrams возвращает множество N-грамм для строки
func getNgrams(s string, n int) map[string]bool {
	result := make(map[string]bool)
	if len(s) < n {
		result[s] = true
		return result
	}
	for i := 0; i <= len(s)-n; i++ {
		result[s[i:i+n]] = true
	}
	return result
}

// 8. Cosine Similarity (косинусное расстояние)
// Представляет строки как векторы частот слов и считает косинус угла между ними
// Пример: "cat dog" и "dog cat" → 1.0 (похожи)
// Используется для сравнения длинных текстов, не зависит от длины строк
func CosineSimilarity(a, b string) float64 {
	if a == "" && b == "" {
		return 1.0
	}
	if a == "" || b == "" {
		return 0.0
	}
	if a == b {
		return 1.0
	}
	// Разбиваем строки на слова
	wordsA := strings.Fields(a)
	wordsB := strings.Fields(b)
	// Строим частотные словари
	freqA := make(map[string]int)
	freqB := make(map[string]int)
	for _, word := range wordsA {
		freqA[word]++
	}
	for _, word := range wordsB {
		freqB[word]++
	}
	// Считаем скалярное произведение и нормы векторов
	dotProduct := 0.0
	normA := 0.0
	normB := 0.0
	allWords := make(map[string]bool)
	for word := range freqA {
		allWords[word] = true
	}
	for word := range freqB {
		allWords[word] = true
	}
	for word := range allWords {
		countA := freqA[word]
		countB := freqB[word]
		dotProduct += float64(countA * countB)
		normA += float64(countA * countA)
		normB += float64(countB * countB)
	}
	if normA == 0 || normB == 0 {
		return 0.0
	}
	// Косинусное расстояние = dotProduct / (|A| * |B|)
	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}