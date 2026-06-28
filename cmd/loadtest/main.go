package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type HotelDTO struct {
	ID        string  `json:"id"`
	Source    string  `json:"source"`
	Name      string  `json:"name"`
	Address   string  `json:"address"`
	City      string  `json:"city"`
	Country   string  `json:"country"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type MatchRequest struct {
	Hotels []HotelDTO `json:"hotels"`
	Config *ConfigDTO `json:"config,omitempty"`
}

type ConfigDTO struct {
	NameWeight     *float64 `json:"nameWeight,omitempty"`
	AddressWeight  *float64 `json:"addressWeight,omitempty"`
	GeoWeight      *float64 `json:"geoWeight,omitempty"`
	LocationWeight *float64 `json:"locationWeight,omitempty"`
	Threshold      *float64 `json:"threshold,omitempty"`
}

var (
	csvFile    = flag.String("file", "data/hotels.csv", "path to CSV")
	serverURL  = flag.String("url", "http://localhost:8080/api/match", "API URL")
	outputFile = flag.String("out", "", "save response to file")
	verbose    = flag.Bool("v", false, "verbose")
	threshold  = flag.Float64("threshold", 0.75, "threshold")
)

func main() {
	flag.Parse()
	if *verbose {
		log.Printf("reading file: %s", *csvFile)
	}
	hotels, err := readCSV(*csvFile)
	if err != nil {
		log.Fatalf("CSV error: %v", err)
	}
	if len(hotels) == 0 {
		log.Fatal("no hotels found")
	}
	log.Printf("read %d hotels", len(hotels))

	req := MatchRequest{
		Hotels: hotels,
		Config: &ConfigDTO{
			Threshold: threshold,
		},
	}
	jsonData, _ := json.Marshal(req)

	resp, err := http.Post(*serverURL, "application/json", bytes.NewReader(jsonData))
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("status %d: %s", resp.StatusCode, string(body))
	}
	if *outputFile != "" {
		os.WriteFile(*outputFile, body, 0644)
		log.Printf("saved to %s", *outputFile)
	} else {
		var pretty bytes.Buffer
		json.Indent(&pretty, body, "", "  ")
		fmt.Println(pretty.String())
	}
}

func readCSV(path string) ([]HotelDTO, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	reader := csv.NewReader(f)
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
	var hotels []HotelDTO
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
		hotels = append(hotels, HotelDTO{
			ID:        providerID,
			Source:    provider,
			Name:      name,
			Address:   address,
			City:      city,
			Country:   country,
			Latitude:  lat,
			Longitude: lon,
		})
	}
	return hotels, nil
}