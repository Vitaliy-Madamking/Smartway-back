package algorithms

import "strings"

func NormalizeString(s string) string {
	if s == "" {
		return s
	}
	s = strings.ToLower(s)
	fields := strings.Fields(s)
	return strings.Join(fields, " ")
}

func NormalizeName(s string) string {
	s = NormalizeString(s)
	replacer := strings.NewReplacer(
		"(", "", ")", "",
		"[", "", "]", "",
		"{", "", "}", "",
	)
	s = replacer.Replace(s)
	return strings.Join(strings.Fields(s), " ")
}

func NormalizeAddress(s string) string {
	s = NormalizeString(s)
	repl := map[string]string{
		"ул.":   "улица",
		"пр.":   "проспект",
		"пер.":  "переулок",
		"пл.":   "площадь",
		"д.":    "дом",
		"кв.":   "квартира",
		"стр.":  "строение",
		"корп.": "корпус",
		"б-р":   "бульвар",
		"наб.":  "набережная",
	}
	for old, new := range repl {
		s = strings.ReplaceAll(s, old, new)
	}
	return strings.Join(strings.Fields(s), " ")
}