package http

import (
	"encoding/csv"
	"hotel-matcher/internal/domain"
	"io"
	"strconv"
	"strings"
)

func parseCSV(r io.Reader) ([]domain.Hotel, error) {
	reader := csv.NewReader(r)
	reader.Comma = ','
	reader.TrimLeadingSpace = true
	reader.FieldsPerRecord = -1

	header, err := reader.Read()
	if err != nil {
		return nil, err
	}
	idxMap := make(map[string]int)
	for i, col := range header {
		idxMap[strings.ToLower(strings.TrimSpace(col))] = i
	}

	var hotels []domain.Hotel
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}
		if len(record) < len(header) {
			continue
		}
		get := func(key string) string {
			if idx, ok := idxMap[strings.ToLower(key)]; ok && idx < len(record) {
				return strings.TrimSpace(record[idx])
			}
			return ""
		}
		provider := get("provider")
		providerID := get("provider_hotel_id")
		name := get("name")
		address := get("address")
		city := get("city")
		country := get("country")
		latStr := get("latitude")
		lonStr := get("longitude")

		if providerID == "" || name == "" {
			continue
		}
		lat, _ := strconv.ParseFloat(latStr, 64)
		lon, _ := strconv.ParseFloat(lonStr, 64)

		hotel := domain.Hotel{
			ID:        providerID,
			Source:    provider,
			Name:      name,
			Address:   address,
			City:      city,
			Country:   country,
			Latitude:  lat,
			Longitude: lon,
		}
		hotels = append(hotels, hotel)
	}
	return hotels, nil
}