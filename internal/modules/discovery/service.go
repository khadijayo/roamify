package discovery

import (
	"sort"
	"strings"

	"github.com/google/uuid"
)

type Service interface {
	GetHomeDashboard() (map[string]interface{}, error)
	VibeSearch(query string) (*VibeSearchResponse, error)
	GetAtlasLocations(filter *AtlasFilterRequest) ([]DiscoveryLocation, error)
	GetLocation(id uuid.UUID) (*DiscoveryLocation, error)
	GlobalSearch(query string) (*GlobalSearchResponse, error)
	TravelAssistant(req *AssistantRequest) (*AssistantResponse, error)
}

type service struct{}

func NewService() Service {
	return &service{}
}

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

func (s *service) GetHomeDashboard() (map[string]interface{}, error) {
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

func (s *service) VibeSearch(query string) (*VibeSearchResponse, error) {
	query = strings.TrimSpace(strings.ToLower(query))
	results := make([]DiscoveryLocation, 0)
	for _, loc := range seededLocations {
		hay := strings.ToLower(loc.Name + " " + loc.Country + " " + loc.Vibe + " " + loc.Description)
		if query == "" || strings.Contains(hay, query) {
			results = append(results, loc)
		}
	}
	return &VibeSearchResponse{Query: query, Results: results}, nil
}

func (s *service) GetAtlasLocations(filter *AtlasFilterRequest) ([]DiscoveryLocation, error) {
	results := make([]DiscoveryLocation, 0)
	for _, loc := range seededLocations {
		if filter.HiddenGemsOnly && !loc.HiddenGem {
			continue
		}
		if filter.Category != "" && !strings.EqualFold(loc.Category, filter.Category) {
			continue
		}
		if filter.Aesthetic != "" && !strings.Contains(strings.ToLower(loc.Vibe), strings.ToLower(filter.Aesthetic)) {
			continue
		}
		results = append(results, loc)
	}

	sortBy := strings.ToLower(filter.SortBy)
	if sortBy == "price_low_to_high" || sortBy == "price" {
		sort.Slice(results, func(i, j int) bool { return results[i].PriceFrom < results[j].PriceFrom })
	}
	return results, nil
}

func (s *service) GetLocation(id uuid.UUID) (*DiscoveryLocation, error) {
	for _, loc := range seededLocations {
		if loc.ID == id {
			return &loc, nil
		}
	}
	return nil, nil
}

func (s *service) GlobalSearch(query string) (*GlobalSearchResponse, error) {
	resp, _ := s.VibeSearch(query)
	hotels := make([]DiscoveryLocation, 0)
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

func (s *service) TravelAssistant(req *AssistantRequest) (*AssistantResponse, error) {
	return &AssistantResponse{
		Suggestion: "Start with a central neighborhood, cluster nearby activities, and reserve one flexible block for spontaneous spots.",
		RoutePlan: []string{
			"Morning: Landmark + nearby cafe",
			"Afternoon: Museum district / local market",
			"Evening: Sunset viewpoint + dinner",
		},
		NextActivities: []string{
			"Find a hidden gem within 2km",
			"Book a transport leg for tomorrow",
			"Add one low-cost local experience",
		},
	}, nil
}
