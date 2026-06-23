package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"sort"
	"strconv"
)

// Standard GeoJSON Types matching your recycling-locations.geojson
type GeoJSONFeatureCollection struct {
	Type     string           `json:"type"`
	Features []GeoJSONFeature `json:"features"`
}

type GeoJSONFeature struct {
	Type       string                 `json:"type"`
	Geometry   Geometry               `json:"geometry"`
	Properties map[string]interface{} `json:"properties"`
}

type Geometry struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"` // [longitude, latitude]
}

func main() {
	// Serves frontend static assets (HTML, CSS, JS) from the 'static' directory
	http.Handle("/", http.FileServer(http.Dir("static")))

	// API search endpoint remains exactly identical for frontend tracking
	http.HandleFunc("/api/search", apiSearchHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Zero-Cost Store Locator Backend listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func apiSearchHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	centerLatStr := r.URL.Query().Get("lat")
	centerLngStr := r.URL.Query().Get("lng")
	radiusMetersStr := r.URL.Query().Get("radius")
	if radiusMetersStr == "" {
		radiusMetersStr = "10000" // 10km default
	}

	if centerLatStr == "" || centerLngStr == "" {
		http.Error(w, `{"error": "Missing lat or lng parameter"}`, http.StatusBadRequest)
		return
	}

	centerLat, _ := strconv.ParseFloat(centerLatStr, 64)
	centerLng, _ := strconv.ParseFloat(centerLngStr, 64)
	radiusMeters, _ := strconv.ParseFloat(radiusMetersStr, 64)
	radiusKm := radiusMeters / 1000.0

	// 1. Load data from your native geojson file layer
	collection, err := loadGeoJSONFromFile("data/recycling-locations.geojson")
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "Failed to read geojson file: %v"}`, err), http.StatusInternalServerError)
		return
	}

	var matchedFeatures []GeoJSONFeature

	// 2. Compute spatial distances matching your coordinates [longitude, latitude]
	for _, feature := range collection.Features {
		if len(feature.Geometry.Coordinates) < 2 {
			continue
		}
		lon := feature.Geometry.Coordinates[0]
		lat := feature.Geometry.Coordinates[1]

		dist := haversine(centerLat, centerLng, lat, lon)
		if dist <= radiusKm {
			// Append the calculated distance dynamically into properties for the UI reader
			feature.Properties["distance_km"] = math.Round(dist*100) / 100
			matchedFeatures = append(matchedFeatures, feature)
		}
	}

	// 3. Sort closest locations first
	sort.Slice(matchedFeatures, func(i, j int) bool {
		// Basic type assertion check to avoid panic operations
		valI, okI := matchedFeatures[i].Properties["distance_km"].(float64)
		valJ, okJ := matchedFeatures[j].Properties["distance_km"].(float64)
		if !okI || !okJ {
			return false
		}
		return valI < valJ
	})

	// Limit to top 25 results to match original structural parameters
	if len(matchedFeatures) > 25 {
		matchedFeatures = matchedFeatures[:25]
	}

	// 4. Send back the exact payload wrapper the frontend map expects
	response := map[string]interface{}{
		"status":   "ok",
		"features": matchedFeatures,
	}

	json.NewEncoder(w).Encode(response)
}

func loadGeoJSONFromFile(filepath string) (*GeoJSONFeatureCollection, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var collection GeoJSONFeatureCollection
	err = json.NewDecoder(file).Decode(&collection)
	if err != nil {
		return nil, err
	}
	return &collection, nil
}

// Haversine formula computes distance between two lat/lng coordinates in kilometers
func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371.0 // Earth radius in kilometers
	dLat := (lat2 - lat1) * math.Pi / 180.0
	dLon := (lon2 - lon1) * math.Pi / 180.0

	l1 := lat1 * math.Pi / 180.0
	l2 := lat2 * math.Pi / 180.0

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Sin(dLon/2)*math.Sin(dLon/2)*math.Cos(l1)*math.Cos(l2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}