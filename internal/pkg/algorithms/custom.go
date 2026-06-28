package algorithms

import (
	"math"
	"strings"
)

// ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ


func min2(a, b int) int {
	if a < b {
		return a
	}
	return b
}

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

func max2(a, b int) int {
	if a > b {
		return a
	}
	return b
}


// 1. LCS - Longest Common Subsequence

func LCS(a, b string) int {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	dp := make([][]int, len(a)+1)
	for i := range dp {
		dp[i] = make([]int, len(b)+1)
	}
	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			if a[i-1] == b[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else {
				dp[i][j] = max2(dp[i-1][j], dp[i][j-1])
			}
		}
	}
	return dp[len(a)][len(b)]
}


// 2. Levenshtein Distance (кастомная версия)

func LevenshteinDistance(a, b string) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}
	prev := make([]int, len(b)+1)
	curr := make([]int, len(b)+1)
	for j := 0; j <= len(b); j++ {
		prev[j] = j
	}
	for i := 1; i <= len(a); i++ {
		curr[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}
			curr[j] = min3(
				prev[j]+1,      // удаление
				curr[j-1]+1,    // вставка
				prev[j-1]+cost, // замена
			)
		}
		prev, curr = curr, prev
	}
	return prev[len(b)]
}

func LevenshteinSimilarity(a, b string) float64 {
	dist := LevenshteinDistance(a, b)
	maxLen := max2(len(a), len(b))
	if maxLen == 0 {
		return 1.0
	}
	return 1.0 - float64(dist)/float64(maxLen)
}


// 3. Damerau-Levenshtein Distance

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
	for i := 0; i <= len(a); i++ {
		dp[i][0] = i
	}
	for j := 0; j <= len(b); j++ {
		dp[0][j] = j
	}
	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}
			dp[i][j] = min3(
				dp[i-1][j]+1,      // удаление
				dp[i][j-1]+1,      // вставка
				dp[i-1][j-1]+cost, // замена
			)
			if i > 1 && j > 1 && a[i-1] == b[j-2] && a[i-2] == b[j-1] {
				dp[i][j] = min2(dp[i][j], dp[i-2][j-2]+1)
			}
		}
	}
	return dp[len(a)][len(b)]
}

func DamerauLevenshteinSimilarity(a, b string) float64 {
	dist := DamerauLevenshteinDistance(a, b)
	maxLen := max2(len(a), len(b))
	if maxLen == 0 {
		return 1.0
	}
	return 1.0 - float64(dist)/float64(maxLen)
}


// 4. Jaro Similarity (кастомная версия)

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
	matchDistance := max2(len(a), len(b))/2 - 1
	if matchDistance < 0 {
		matchDistance = 0
	}
	matchesA := make([]bool, len(a))
	matchesB := make([]bool, len(b))
	matches := 0
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
	return (float64(matches)/float64(len(a)) +
		float64(matches)/float64(len(b)) +
		float64(matches-transpositions)/float64(matches)) / 3.0
}


// 5. Jaro-Winkler Similarity (кастомная версия)

func JaroWinklerSimilarityCustom(a, b string) float64 {
	jaro := JaroSimilarity(a, b)
	prefixLength := 0
	maxPrefix := min2(4, min2(len(a), len(b)))
	for i := 0; i < maxPrefix; i++ {
		if a[i] == b[i] {
			prefixLength++
		} else {
			break
		}
	}
	return jaro + float64(prefixLength)*0.1*(1-jaro)
}


// 6. Soundex 

func Soundex(s string) string {
	if s == "" {
		return ""
	}

	s = strings.ToUpper(s)
	result := string(s[0])

	mapping := map[rune]string{
		'B': "1", 'P': "1", 'F': "1", 'V': "1",
		'C': "2", 'G': "2", 'J': "2", 'K': "2", 'Q': "2", 'S': "2", 'X': "2", 'Z': "2",
		'D': "3", 'T': "3",
		'L': "4",
		'M': "5", 'N': "5",
		'R': "6",
	}

	prevCode := ""
	for i := 1; i < len(s) && len(result) < 4; i++ {
		char := rune(s[i])
		code, exists := mapping[char]
		if !exists {
			continue
		}
		if code != prevCode {
			result += code
			prevCode = code
		}
	}

	for len(result) < 4 {
		result += "0"
	}

	return result
}

func SoundexSimilarity(a, b string) float64 {
	if Soundex(a) == Soundex(b) {
		return 1.0
	}
	return 0.0
}


// 7. N-Gram Similarity

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
	intersection := 0
	for ngram := range ngramsA {
		if ngramsB[ngram] {
			intersection++
		}
	}
	return float64(intersection) / float64(max2(len(ngramsA), len(ngramsB)))
}

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


// 8. Cosine Similarity

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

	wordsA := strings.Fields(a)
	wordsB := strings.Fields(b)

	freqA := make(map[string]int)
	freqB := make(map[string]int)

	for _, word := range wordsA {
		freqA[word]++
	}
	for _, word := range wordsB {
		freqB[word]++
	}

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

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}