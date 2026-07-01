package algorithms

import (
	"regexp"
	"strings"
)

// NormalizeString - базовая нормализация
func NormalizeString(s string) string {
    if s == "" {
        return s
    }
    // Приводим к нижнему регистру
    s = strings.ToLower(s)
    // Удаляем лишние пробелы
    fields := strings.Fields(s)
    return strings.Join(fields, " ")
}

// NormalizeName - нормализация названий отелей
func NormalizeName(s string) string {
    s = NormalizeString(s)
    
    // Удаляем скобки и их содержимое
    re := regexp.MustCompile(`\([^)]*\)`)
    s = re.ReplaceAllString(s, "")
    
    // Удаляем квадратные скобки
    re = regexp.MustCompile(`\[[^\]]*\]`)
    s = re.ReplaceAllString(s, "")
    
    // Удаляем суффиксы типа " - Adults Only", " - Adults Only"
    suffixes := []string{
        "- adults only", "adults only", 
        " - adults", "adults",
        " - resort", "resort",
        " hotel", " hotel & suites",
        " inn", " motel",
        " lodge", " villa",
        " apartments", " apartments",
    }
    for _, suffix := range suffixes {
        s = strings.ReplaceAll(s, suffix, "")
    }
    
    // Удаляем лишние пробелы
    return strings.Join(strings.Fields(s), " ")
}

// NormalizeAddress - нормализация адресов
func NormalizeAddress(s string) string {
    s = NormalizeString(s)
    
    // Русские сокращения → полные слова
    ruRepl := map[string]string{
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
    for old, new := range ruRepl {
        s = strings.ReplaceAll(s, old, new)
    }
    
    // Английские сокращения
    enRepl := map[string]string{
        "st.": "street",
        "st ":  "street ",
        "ave.": "avenue",
        "ave ": "avenue ",
        "rd.":  "road",
        "rd ":  "road ",
        "blvd.": "boulevard",
        "blvd ": "boulevard ",
        "ln.":   "lane",
        "ln ":   "lane ",
        "dr.":   "drive",
        "dr ":   "drive ",
        "ct.":   "court",
        "ct ":   "court ",
        "pl.":   "place",
        "pl ":   "place ",
    }
    for old, new := range enRepl {
        s = strings.ReplaceAll(s, old, new)
    }
    
    // Удаляем почтовые индексы (цифры 5-6 символов)
    re := regexp.MustCompile(`\b\d{5,6}\b`)
    s = re.ReplaceAllString(s, "")
    
    // Удаляем лишние пробелы
    return strings.Join(strings.Fields(s), " ")
}

// NormalizeLocation - нормализация города и страны
func NormalizeLocation(city, country string) string {
    city = NormalizeString(city)
    country = NormalizeString(country)
    
    // Удаляем " (and vicinity)" и подобное
    city = strings.ReplaceAll(city, "(and vicinity)", "")
    city = strings.ReplaceAll(city, "(vicinity)", "")
    
    // Удаляем лишние пробелы
    city = strings.Join(strings.Fields(city), " ")
    country = strings.Join(strings.Fields(country), " ")
    
    return city + "|" + country
}