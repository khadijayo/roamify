package discovery

import "github.com/google/uuid"

type DiscoveryLocation struct {
	ID                uuid.UUID `json:"id"`
	Name              string    `json:"name"`
	Country           string    `json:"country"`
	Vibe              string    `json:"vibe"`
	PriceFrom         float64   `json:"price_from"`
	Category          string    `json:"category"`
	Lat               float64   `json:"lat"`
	Lng               float64   `json:"lng"`
	HiddenGem         bool      `json:"hidden_gem"`
	Instagrammable    bool      `json:"instagrammable"`
	LocalLegend       bool      `json:"local_legend"`
	StreetViewEnabled bool      `json:"street_view_enabled"`
	Description       string    `json:"description"`
	BookingURL        string    `json:"booking_url"`
	ImageURL          string    `json:"image_url"`
}

type SquadSuggestion struct {
	TripName       string  `json:"trip_name"`
	Destination    string  `json:"destination"`
	CurrentMembers int     `json:"current_members"`
	Capacity       int     `json:"capacity"`
	PriceEstimate  float64 `json:"price_estimate"`
	Vibe           string  `json:"vibe"`
}

type PriceDropItem struct {
	Route        string  `json:"route"`
	Provider     string  `json:"provider"`
	CurrentPrice float64 `json:"current_price"`
	DropPercent  int     `json:"drop_percent"`
}

type VibeSearchResponse struct {
	Query   string              `json:"query"`
	Results []DiscoveryLocation `json:"results"`
}

type AtlasFilterRequest struct {
	Aesthetic      string `form:"aesthetic"`
	SortBy         string `form:"sort_by"`
	HiddenGemsOnly bool   `form:"hidden_gems_only"`
	Category       string `form:"category"`
}

type GlobalSearchResponse struct {
	Query        string              `json:"query"`
	Destinations []DiscoveryLocation `json:"destinations"`
	Hotels       []DiscoveryLocation `json:"hotels"`
	Flights      []PriceDropItem     `json:"flights"`
}

type AssistantRequest struct {
	Prompt      string   `json:"prompt" binding:"required"`
	Waypoints   []string `json:"waypoints"`
	Destination string   `json:"destination"`
}

type AssistantResponse struct {
	Suggestion     string   `json:"suggestion"`
	RoutePlan      []string `json:"route_plan"`
	NextActivities []string `json:"next_activities"`
}
