package discovery

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

// -------------------- SERVICE INTERFACE --------------------

type Service interface {
	GetHomeDashboard() (map[string]interface{}, error)
	VibeSearch(query string) (*VibeSearchResponse, error)
	GetAtlasLocations(filter *AtlasFilterRequest) ([]DiscoveryLocation, error)
	GetLocation(id uuid.UUID) (*DiscoveryLocation, error)
	GlobalSearch(query string) (*GlobalSearchResponse, error)
	TravelAssistant(req *AssistantRequest) (*AssistantResponse, error)
}

// -------------------- MOCK DATA --------------------

var seededLocations = []DiscoveryLocation{
	{ID: uuid.New(), Name: "Tbilisi", Country: "Georgia", Vibe: "Dark Academia", PriceFrom: 320, Category: "city", Lat: 41.7151, Lng: 44.8271, HiddenGem: true, Instagrammable: true, LocalLegend: true, StreetViewEnabled: true, Description: "Old-town texture, moody cafes, and history-heavy neighborhoods.", BookingURL: "https://www.booking.com", ImageURL: "https://images.unsplash.com/photo-1589656966895-2f33e7653819?w=400&q=80"},
	{ID: uuid.New(), Name: "Porto", Country: "Portugal", Vibe: "Artsy Chaos", PriceFrom: 280, Category: "city", Lat: 41.1579, Lng: -8.6291, HiddenGem: false, Instagrammable: true, LocalLegend: true, StreetViewEnabled: true, Description: "Riverside tiles, indie galleries, and creative nightlife.", BookingURL: "https://www.booking.com", ImageURL: "https://images.unsplash.com/photo-1555881400-74d7acaacd8b?w=400&q=80"},
	{ID: uuid.New(), Name: "Bali", Country: "Indonesia", Vibe: "Coastal Dream", PriceFrom: 240, Category: "beach", Lat: -8.4095, Lng: 115.1889, HiddenGem: false, Instagrammable: true, LocalLegend: false, StreetViewEnabled: true, Description: "Waterfalls, rice terraces, and wellness-forward stays.", BookingURL: "https://www.booking.com", ImageURL: "https://images.unsplash.com/photo-1537996194471-e657df975ab4?w=400&q=80"},
	{ID: uuid.New(), Name: "Chefchaouen", Country: "Morocco", Vibe: "Dream Core", PriceFrom: 260, Category: "city", Lat: 35.1688, Lng: -5.2636, HiddenGem: true, Instagrammable: true, LocalLegend: true, StreetViewEnabled: true, Description: "Blue medina alleys with rich local artisan culture.", BookingURL: "https://www.booking.com", ImageURL: "https://images.unsplash.com/photo-1553659971-f01207815844?w=400&q=80"},
}

var seededFlights = []PriceDropItem{
	{Route: "NYC -> Paris", Provider: "Air France", CurrentPrice: 289, DropPercent: 42},
	{Route: "LA -> Tokyo", Provider: "JAL", CurrentPrice: 412, DropPercent: 35},
	{Route: "London -> Dubai", Provider: "Emirates", CurrentPrice: 198, DropPercent: 51},
}

// -------------------- SERVICE IMPLEMENTATION --------------------

type realService struct {
	geminiKey  string
	httpClient *http.Client
}

// Constructor (matches your main.go)
func NewService(geminiKey string) Service {
	return &realService{
		geminiKey:  geminiKey,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// -------------------- CORE METHODS --------------------

func (s *realService) GetHomeDashboard() (map[string]interface{}, error) {
	squad := []SquadSuggestion{
		{TripName: "Bali Aesthetic Tour", Destination: "Bali, Indonesia", CurrentMembers: 4, Capacity: 6, PriceEstimate: 1200, Vibe: "Zen & Tropical"},
		{TripName: "Tokyo Neon Nights", Destination: "Tokyo, Japan", CurrentMembers: 3, Capacity: 5, PriceEstimate: 2100, Vibe: "Urban Explorer"},
	}

	return map[string]interface{}{
		"vibe_spots":        seededLocations,
		"squad_suggestions": squad,
		"price_drops":       seededFlights,
	}, nil
}

func (s *realService) VibeSearch(query string) (*VibeSearchResponse, error) {
	query = strings.ToLower(strings.TrimSpace(query))

	var results []DiscoveryLocation
	for _, loc := range seededLocations {
		hay := strings.ToLower(loc.Name + " " + loc.Country + " " + loc.Vibe + " " + loc.Description)
		if query == "" || strings.Contains(hay, query) {
			results = append(results, loc)
		}
	}

	return &VibeSearchResponse{
		Query:   query,
		Results: results,
	}, nil
}

func (s *realService) GetAtlasLocations(filter *AtlasFilterRequest) ([]DiscoveryLocation, error) {
	var results []DiscoveryLocation

	for _, loc := range seededLocations {
		if filter.HiddenGemsOnly && !loc.HiddenGem {
			continue
		}
		if filter.Category != "" && !strings.EqualFold(filter.Category, loc.Category) {
			continue
		}
		if filter.Aesthetic != "" && !strings.Contains(strings.ToLower(loc.Vibe), strings.ToLower(filter.Aesthetic)) {
			continue
		}
		results = append(results, loc)
	}

	if strings.ToLower(filter.SortBy) == "price" {
		sort.Slice(results, func(i, j int) bool {
			return results[i].PriceFrom < results[j].PriceFrom
		})
	}

	return results, nil
}

func (s *realService) GetLocation(id uuid.UUID) (*DiscoveryLocation, error) {
	for _, loc := range seededLocations {
		if loc.ID == id {
			return &loc, nil
		}
	}
	return nil, nil
}

func (s *realService) GlobalSearch(query string) (*GlobalSearchResponse, error) {
	resp, _ := s.VibeSearch(query)

	var hotels []DiscoveryLocation
	for _, loc := range resp.Results {
		if loc.Category == "city" || loc.Category == "beach" {
			hotels = append(hotels, loc)
		}
	}

	return &GlobalSearchResponse{
		Query:        query,
		Destinations: resp.Results,
		Hotels:       hotels,
		Flights:      seededFlights,
	}, nil
}

// -------------------- GEMINI AI --------------------

func (s *realService) TravelAssistant(req *AssistantRequest) (*AssistantResponse, error) {
	if s.geminiKey == "" {
		return &AssistantResponse{
			Suggestion: "Start central, explore nearby, keep flexibility.",
			RoutePlan:  []string{"Morning: Explore", "Afternoon: Activities", "Evening: Relax"},
			NextActivities: []string{
				"Find hidden gems",
				"Try local food",
				"Plan next step",
			},
		}, nil
	}

	prompt := req.Prompt
	if req.Destination != "" {
		prompt = fmt.Sprintf("Destination: %s\n%s", req.Destination, req.Prompt)
	}

	if len(req.Waypoints) > 0 {
		prompt += "\nWaypoints: " + strings.Join(req.Waypoints, ", ")
	}

	prompt += "\nReturn ONLY JSON: {suggestion:string, route_plan:[], next_activities:[]}"

	bodyMap := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]string{
					{"text": prompt},
				},
			},
		},
	}

	body, _ := json.Marshal(bodyMap)

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash:generateContent?key=%s", s.geminiKey)

	httpReq, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var geminiResp struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.Unmarshal(respBody, &geminiResp); err != nil {
		return nil, err
	}

	if len(geminiResp.Candidates) == 0 {
		return nil, errors.New("empty gemini response")
	}

	text := geminiResp.Candidates[0].Content.Parts[0].Text

	var result struct {
		Suggestion     string   `json:"suggestion"`
		RoutePlan      []string `json:"route_plan"`
		NextActivities []string `json:"next_activities"`
	}

	if err := json.Unmarshal([]byte(text), &result); err != nil {
		return &AssistantResponse{Suggestion: text}, nil
	}

	return &AssistantResponse{
		Suggestion:     result.Suggestion,
		RoutePlan:      result.RoutePlan,
		NextActivities: result.NextActivities,
	}, nil
}