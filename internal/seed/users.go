package seed

import (
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/khadijayo/roamify/internal/modules/users"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// SeedUsers creates 8 realistic demo users.
// It is idempotent: if any seeded email already exists the whole step is skipped.
func SeedUsers(db *gorm.DB) ([]users.User, error) {
	// ── Guard: skip if already seeded ─────────────────────────────────────────
	var count int64
	db.Model(&users.User{}).Where("email IN ?", seededEmails()).Count(&count)
	if count > 0 {
		log.Println("[seed] users already seeded – skipping")
		var existing []users.User
		db.Where("email IN ?", seededEmails()).Find(&existing)
		return existing, nil
	}

	type userData struct {
		fullName  string
		email     string
		bio       string
		avatarURL string
		pace      users.TravelPace
		budget    users.BudgetStyle
		with      users.TravelWith
		explorer  string
		vibes     []string
		interests []string
	}

	rows := []userData{
		{
			fullName:  "Ayla Reyes",
			email:     "ayla.reyes@roamify.demo",
			bio:       "Chasing sunsets and street food across Southeast Asia. Always planning the next escape 🌏",
			avatarURL: "https://images.unsplash.com/photo-1529626455594-4ff0802cfb7e?w=400&q=80",
			pace:      users.PaceAdventure,
			budget:    users.BudgetMidRange,
			with:      users.WithSquad,
			explorer:  "Cultural Nomad",
			vibes:     []string{"beach", "culture", "foodie"},
			interests: []string{"photography", "hiking", "local cuisine"},
		},
		{
			fullName:  "Marco Esposito",
			email:     "marco.esposito@roamify.demo",
			bio:       "Italian wanderer based in Barcelona. Coffee snob, architecture geek, slow traveler.",
			avatarURL: "https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?w=400&q=80",
			pace:      users.PaceChill,
			budget:    users.BudgetMidRange,
			with:      users.WithPartner,
			explorer:  "Slow Traveler",
			vibes:     []string{"city", "history", "cafe-hopping"},
			interests: []string{"architecture", "espresso", "art museums"},
		},
		{
			fullName:  "Priya Nair",
			email:     "priya.nair@roamify.demo",
			bio:       "Solo female traveler. 42 countries and counting 🗺️ | Wildlife & wilderness enthusiast.",
			avatarURL: "https://images.unsplash.com/photo-1494790108377-be9c29b29330?w=400&q=80",
			pace:      users.PaceBalanced,
			budget:    users.BudgetBackpacker,
			with:      users.WithSolo,
			explorer:  "Wilderness Seeker",
			vibes:     []string{"nature", "adventure", "off-the-beaten-path"},
			interests: []string{"wildlife", "camping", "trekking"},
		},
		{
			fullName:  "Jordan Blake",
			email:     "jordan.blake@roamify.demo",
			bio:       "NYC-based travel blogger. Luxury hotels, rooftop bars, and first-class flights.",
			avatarURL: "https://images.unsplash.com/photo-1500648767791-00dcc994a43e?w=400&q=80",
			pace:      users.PaceChill,
			budget:    users.BudgetLuxury,
			with:      users.WithPartner,
			explorer:  "Luxury Globetrotter",
			vibes:     []string{"luxury", "city", "nightlife"},
			interests: []string{"fine dining", "spa", "rooftop bars"},
		},
		{
			fullName:  "Yuki Tanaka",
			email:     "yuki.tanaka@roamify.demo",
			bio:       "Tokyo → the world. Obsessed with minimalist travel and hidden gems in Japan 🇯🇵",
			avatarURL: "https://images.unsplash.com/photo-1534528741775-53994a69daeb?w=400&q=80",
			pace:      users.PaceBalanced,
			budget:    users.BudgetMidRange,
			with:      users.WithSolo,
			explorer:  "Hidden Gem Hunter",
			vibes:     []string{"culture", "minimalist", "food"},
			interests: []string{"onsen", "anime", "street food"},
		},
		{
			fullName:  "Kofi Mensah",
			email:     "kofi.mensah@roamify.demo",
			bio:       "Accra to the world 🌍 Exploring the African continent and beyond. Squad travel always.",
			avatarURL: "https://images.unsplash.com/photo-1506794778202-cad84cf45f1d?w=400&q=80",
			pace:      users.PaceAdventure,
			budget:    users.BudgetBackpacker,
			with:      users.WithSquad,
			explorer:  "Continent Hopper",
			vibes:     []string{"nature", "culture", "adventure"},
			interests: []string{"safaris", "music festivals", "local transport"},
		},
		{
			fullName:  "Sofia Larsson",
			email:     "sofia.larsson@roamify.demo",
			bio:       "Swedish nomad. Remote worker who turns every trip into a co-working adventure ☕💻",
			avatarURL: "https://images.unsplash.com/photo-1544005313-94ddf0286df2?w=400&q=80",
			pace:      users.PaceBalanced,
			budget:    users.BudgetMidRange,
			with:      users.WithSolo,
			explorer:  "Digital Nomad",
			vibes:     []string{"cafe", "coworking", "city"},
			interests: []string{"productivity", "coffee shops", "yoga"},
		},
		{
			fullName:  "Liam O'Connor",
			email:     "liam.oconnor@roamify.demo",
			bio:       "Irish lad living out of a 20L backpack. Budget travel tips, hostel reviews & pub crawls.",
			avatarURL: "https://images.unsplash.com/photo-1548372290-8d01b6c8e78c?w=400&q=80",
			pace:      users.PaceAdventure,
			budget:    users.BudgetBackpacker,
			with:      users.WithSquad,
			explorer:  "Budget Backpacker",
			vibes:     []string{"backpacker", "nightlife", "social"},
			interests: []string{"hostels", "pub crawls", "hiking"},
		},
	}

	// hash password once – all demo users share "Roamify2025!"
	hash, err := bcrypt.GenerateFromPassword([]byte("Roamify2025!"), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("bcrypt: %w", err)
	}
	hashStr := string(hash)

	var created []users.User
	now := time.Now()

	for i, r := range rows {
		email := r.email
		avatarURL := r.avatarURL

		// Spread creation times over the past 6 months for realism
		createdAt := now.AddDate(0, -6, 0).Add(time.Duration(i*13*24) * time.Hour)

		u := users.User{
			ID:           uuid.New(),
			Email:        &email,
			FullName:     r.fullName,
			AvatarURL:    &avatarURL,
			PasswordHash: &hashStr,
			Role:         users.RoleUser,
			Status:       users.StatusActive,
			IsVerified:   true,
			CreatedAt:    createdAt,
			UpdatedAt:    createdAt,
		}

		if err := db.Create(&u).Error; err != nil {
			return nil, fmt.Errorf("create user %s: %w", r.email, err)
		}

		// Vibe profile
		vp := users.VibeProfile{
			ID:                 uuid.New(),
			UserID:             u.ID,
			ExplorerType:       r.explorer,
			PreferredVibes:     pq.StringArray(r.vibes),
			TravelPace:         r.pace,
			BudgetStyle:        r.budget,
			TravelWith:         r.with,
			Interests:          pq.StringArray(r.interests),
			OnboardingComplete: true,
			ExplorerLevel:      randomIntRange(1, 8),
			RoamifyPoints:      randomIntRange(100, 2500),
			CountriesVisited:   randomIntRange(3, 40),
			CreatedAt:          createdAt,
			UpdatedAt:          createdAt,
		}
		if err := db.Create(&vp).Error; err != nil {
			log.Printf("[seed] warn: vibe profile for %s: %v", r.email, err)
		}

		created = append(created, u)
	}

	log.Printf("[seed] ✅ created %d users", len(created))
	return created, nil
}

func seededEmails() []string {
	return []string{
		"ayla.reyes@roamify.demo",
		"marco.esposito@roamify.demo",
		"priya.nair@roamify.demo",
		"jordan.blake@roamify.demo",
		"yuki.tanaka@roamify.demo",
		"kofi.mensah@roamify.demo",
		"sofia.larsson@roamify.demo",
		"liam.oconnor@roamify.demo",
	}
}
