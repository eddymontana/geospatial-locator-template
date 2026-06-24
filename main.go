package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

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
	Coordinates []float64 `json:"coordinates"`
}

// init() executes automatically before main() boots up
func init() {
	envPath := ".env"
	file, err := os.Open(envPath)
	if err != nil {
		log.Println("Note: Local .env configuration file not found. Falling back to system environment variables.")
		return
	}
	defer file.Close()

	log.Println("SUCCESS: Found local .env file. Parsing configuration keys...")
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip empty lines or documentation/comment blocks
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Split strictly on the first '=' character to preserve values containing equal signs
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		
		// Strip optional outer quotes from the string values if present
		value = strings.Trim(value, `"'`)

		os.Setenv(key, value)
		log.Printf("Loaded Environment Variable: %s", key)
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Warning: Encountered error while parsing .env layout: %v", err)
	}
}

func main() {
	// Intercept the root path to dynamically inject environment variables into index.html
	http.HandleFunc("/", serveIndexWithEnv)

	// Serve remaining assets (app.js, style.css) from static directory safely
	http.HandleFunc("/app.js", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join("static", "app.js"))
	})
	http.HandleFunc("/style.css", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join("static", "style.css"))
	})

	http.HandleFunc("/api/search", apiSearchHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Secure Spatial Locator Backend running on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

// Dynamically injects MAPS_API_KEY into the script tag placeholder at runtime
func serveIndexWithEnv(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	indexPath := filepath.Join("static", "index.html")
	content, err := os.ReadFile(indexPath)
	if err != nil {
		http.Error(w, "HTML configuration asset missing.", http.StatusInternalServerError)
		return
	}

	// Fetch API Key from environment variable instead of static code tracking
	apiKey := os.Getenv("MAPS_API_KEY")
	if apiKey == "" {
		log.Println("WARNING: MAPS_API_KEY environment parameter is completely empty.")
		apiKey = "GOOGLE_MAPS_API_KEY_PLACEHOLDER" // Fallback to avoid breaking structural parsing
	}

	// Replace the placeholder dynamically on request lifecycle loops
	htmlOutput := strings.Replace(string(content), "GOOGLE_MAPS_API_KEY_PLACEHOLDER", apiKey, 1)

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, htmlOutput)
}

func apiSearchHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	centerLatStr := r.URL.Query().Get("lat")
	centerLngStr := r.URL.Query().Get("lng")
	radiusMetersStr := r.URL.Query().Get("radius")
	if radiusMetersStr == "" {
		radiusMetersStr = "10000"
	}

	if centerLatStr == "" || centerLngStr == "" {
		http.Error(w, `{"error": "Missing lat or lng parameter"}`, http.StatusBadRequest)
		return
	}

	centerLat, _ := strconv.ParseFloat(centerLatStr, 64)
	centerLng, _ := strconv.ParseFloat(centerLngStr, 64)
	radiusMeters, _ := strconv.ParseFloat(radiusMetersStr, 64)
	radiusKm := radiusMeters / 1000.0

	collection, err := loadGeoJSONFromFile("data/recycling-locations.geojson")
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "Failed to read geojson file: %v"}`, err), http.StatusInternalServerError)
		return
	}

	var matchedFeatures []GeoJSONFeature

	for _, feature := range collection.Features {
		if len(feature.Geometry.Coordinates) < 2 {
			continue
		}
		lon := feature.Geometry.Coordinates[0]
		lat := feature.Geometry.Coordinates[1]

		dist := haversine(centerLat, centerLng, lat, lon)
		if dist <= radiusKm {
			feature.Properties["distance_km"] = math.Round(dist*100) / 100
			matchedFeatures = append(matchedFeatures, feature)
		}
	}

	sort.Slice(matchedFeatures, func(i, j int) bool {
		valI, okI := matchedFeatures[i].Properties["distance_km"].(float64)
		valJ, okJ := matchedFeatures[j].Properties["distance_km"].(float64)
		if !okI || !okJ {
			return false
		}
		return valI < valJ
	})

	if len(matchedFeatures) > 25 {
		matchedFeatures = matchedFeatures[:25]
	}

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

func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371.0
	dLat := (lat2 - lat1) * math.Pi / 180.0
	dLon := (lon2 - lon1) * math.Pi / 180.0
	l1 := lat1 * math.Pi / 180.0
	l2 := lat2 * math.Pi / 180.0
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Sin(dLon/2)*math.Sin(dLon/2)*math.Cos(l1)*math.Cos(l2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}