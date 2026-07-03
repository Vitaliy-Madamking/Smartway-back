package benchmarks

import (
	"testing"

	"hotel-matcher/internal/pkg/algorithms"
)

// TestDummy - пустой тест для запуска пакета
func TestDummy(t *testing.T) {

}



// Тестовые данные
var testPairs = []struct {
	name string
	a    string
	b    string
}{
	{"identical", "Rodeway Inn Union", "Rodeway Inn Union"},
	{"similar_names", "Rodeway Inn Union", "Rodeway Inn"},
	{"different_names", "Rodeway Inn Union", "Hilton Garden Inn"},
	{"addresses_similar", "1235 S Duncan Byp", "1235 Duncan Bypass"},
	{"addresses_different", "1235 S Duncan Byp", "801 R Street"},
	{"cities_similar", "London", "Londn"},
}

// Бенчмарк Jaro-Winkler
func BenchmarkJaroWinkler(b *testing.B) {
	for _, pair := range testPairs {
		b.Run(pair.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				algorithms.JaroWinklerSimilarityCustom(pair.a, pair.b)
			}
		})
	}
}

// Бенчмарк Jaro
func BenchmarkJaro(b *testing.B) {
	for _, pair := range testPairs {
		b.Run(pair.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				algorithms.JaroSimilarity(pair.a, pair.b)
			}
		})
	}
}

// Бенчмарк Levenshtein
func BenchmarkLevenshtein(b *testing.B) {
	for _, pair := range testPairs {
		b.Run(pair.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				algorithms.LevenshteinSimilarity(pair.a, pair.b)
			}
		})
	}
}

// Бенчмарк Damerau-Levenshtein
func BenchmarkDamerauLevenshtein(b *testing.B) {
	for _, pair := range testPairs {
		b.Run(pair.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				algorithms.DamerauLevenshteinSimilarity(pair.a, pair.b)
			}
		})
	}
}

// Бенчмарк Soundex
func BenchmarkSoundex(b *testing.B) {
	for _, pair := range testPairs {
		b.Run(pair.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				algorithms.SoundexSimilarity(pair.a, pair.b)
			}
		})
	}
}

// Бенчмарк N-gram
func BenchmarkNGram(b *testing.B) {
	for _, pair := range testPairs {
		b.Run(pair.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				algorithms.NGramSimilarity(pair.a, pair.b, 2)
			}
		})
	}
}

// Сравнение всех алгоритмов
func BenchmarkAllAlgorithms(b *testing.B) {
	algorithmsList := []string{
		"jaro-winkler",
		"jaro",
		"levenshtein",
		"damerau-levenshtein",
		"soundex",
		"ngram",
	}

	for _, alg := range algorithmsList {
		b.Run(alg, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				algorithms.CompareWithCustomAlgorithm(
					"Rodeway Inn Union",
					"Rodeway Inn",
					alg,
				)
			}
		})
	}
}