package discovery

import (
	"github.com/google/uuid"
	"github.com/khadijayo/roamify/internal/modules/users"
)

// PersonalizedService extends discovery with user-aware recommendations.
// It depends on the users.Repository to read vibe profile data.
type PersonalizedService struct {
	userRepo users.Repository
}

func NewPersonalizedService(userRepo users.Repository) *PersonalizedService {
	return &PersonalizedService{userRepo: userRepo}
}

// PriceDropAlert represents a deal tailored to the user's budget and interests.
type PriceDropAlert struct {
	Route        string   `json:"route"`
	Provider     string   `json:"provider"`
	CurrentPrice float64  `json:"current_price"`
	DropPercent  int      `json:"drop_percent"`
	Category     string   `json:"category"`
	Tags         []string `json:"tags"`
	MatchReason  string   `json:"match_reason"`
}

// GetPersonalizedPriceDrops returns deals filtered to the user's budget style and interests.
// Currently uses the seeded flight data with vibe-profile filtering applied.
// When you integrate a real flights/hotels API (e.g. Amadeus), replace the
// allDeals slice with live API results — the filtering logic stays the same.
func (s *PersonalizedService) GetPersonalizedPriceDrops(userID uuid.UUID) ([]PriceDropAlert, error) {
	vp, err := s.userRepo.GetVibeProfile(userID)
	if err != nil {
		// No vibe profile yet — return all deals unfiltered
		return defaultAlerts(), nil
	}

	allDeals := allSeedAlerts()

	// Filter by budget style
	var filtered []PriceDropAlert
	for _, deal := range allDeals {
		if matchesBudget(deal, string(vp.BudgetStyle)) {
			deal.MatchReason = buildMatchReason(vp)
			filtered = append(filtered, deal)
		}
	}

	// If nothing matched, fall back to all deals
	if len(filtered) == 0 {
		return defaultAlerts(), nil
	}
	return filtered, nil
}

// RecommendedDestinations returns discovery locations filtered by the user's interests and vibe.
func (s *PersonalizedService) RecommendedDestinations(userID uuid.UUID) ([]DiscoveryLocation, error) {
	vp, err := s.userRepo.GetVibeProfile(userID)
	if err != nil {
		return seededLocations, nil
	}

	// Score each location against the user's interests
	type scored struct {
		loc   DiscoveryLocation
		score int
	}
	var scoredLocs []scored
	for _, loc := range seededLocations {
		score := 0
		locLower := toLower(loc.Vibe + " " + loc.Category + " " + loc.Description)
		for _, interest := range vp.Interests {
			if contains(locLower, toLower(interest)) {
				score++
			}
		}
		for _, vibe := range vp.PreferredVibes {
			if contains(locLower, toLower(vibe)) {
				score++
			}
		}
		scoredLocs = append(scoredLocs, scored{loc: loc, score: score})
	}

	// Sort highest score first
	for i := 0; i < len(scoredLocs)-1; i++ {
		for j := i + 1; j < len(scoredLocs); j++ {
			if scoredLocs[j].score > scoredLocs[i].score {
				scoredLocs[i], scoredLocs[j] = scoredLocs[j], scoredLocs[i]
			}
		}
	}

	result := make([]DiscoveryLocation, 0, len(scoredLocs))
	for _, s := range scoredLocs {
		result = append(result, s.loc)
	}
	return result, nil
}

// ---- helpers ----

func matchesBudget(deal PriceDropAlert, budgetStyle string) bool {
	switch budgetStyle {
	case "backpacker":
		return deal.CurrentPrice < 300
	case "mid-range":
		return deal.CurrentPrice >= 150 && deal.CurrentPrice <= 600
	case "luxury":
		return true // luxury users see all deals
	default:
		return true
	}
}

func buildMatchReason(vp *users.VibeProfile) string {
	if len(vp.Interests) > 0 {
		return "Matched your interest in " + vp.Interests[0]
	}
	if vp.ExplorerType != "" {
		return "Picked for " + vp.ExplorerType + " explorers"
	}
	return "Based on your travel style"
}

func toLower(s string) string {
	result := make([]byte, len(s))
	for i := range s {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 32
		}
		result[i] = c
	}
	return string(result)
}

func contains(hay, needle string) bool {
	return len(needle) > 0 && len(hay) >= len(needle) &&
		(hay == needle || len(hay) > 0 && containsStr(hay, needle))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func allSeedAlerts() []PriceDropAlert {
	return []PriceDropAlert{
		{Route: "NYC → Paris", Provider: "Air France", CurrentPrice: 289, DropPercent: 42, Category: "flight", Tags: []string{"europe", "culture", "art"}},
		{Route: "LA → Tokyo", Provider: "JAL", CurrentPrice: 412, DropPercent: 35, Category: "flight", Tags: []string{"asia", "urban", "food"}},
		{Route: "London → Dubai", Provider: "Emirates", CurrentPrice: 198, DropPercent: 51, Category: "flight", Tags: []string{"luxury", "beach", "shopping"}},
		{Route: "NYC → Bali", Provider: "Qatar Airways", CurrentPrice: 650, DropPercent: 28, Category: "flight", Tags: []string{"beach", "wellness", "nature"}},
		{Route: "Paris → Marrakech", Provider: "Royal Air Maroc", CurrentPrice: 140, DropPercent: 38, Category: "flight", Tags: []string{"culture", "food", "adventure"}},
	}
}

func defaultAlerts() []PriceDropAlert {
	return allSeedAlerts()
}
