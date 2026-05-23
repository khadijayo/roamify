package seed

import (
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/khadijayo/roamify/internal/modules/trips"
	"github.com/khadijayo/roamify/internal/modules/users"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// SeedTrips creates 15 realistic trips assigned to the seeded users.
// Mix of past (completed), ongoing, and future (planning) trips.
func SeedTrips(db *gorm.DB, seedUsers []users.User) ([]trips.Trip, error) {
	if len(seedUsers) == 0 {
		return nil, fmt.Errorf("seedTrips requires at least one user")
	}

	// Guard: skip if trips with these titles already exist
	var count int64
	db.Model(&trips.Trip{}).Where("title IN ?", seededTripTitles()).Count(&count)
	if count > 0 {
		log.Println("[seed] trips already seeded – skipping")
		var existing []trips.Trip
		db.Where("title IN ?", seededTripTitles()).Find(&existing)
		return existing, nil
	}

	now := time.Now()

	type tripData struct {
		title           string
		destination     string
		tripType        trips.TripType
		vibeTags        []string
		budget          float64
		travelersPlanned int
		startOffset     int // days from now (negative = past)
		duration        int // days
		status          trips.TripStatus
		coverURL        string
		notes           string
		ownerIdx        int // index into seedUsers slice
	}

	rows := []tripData{
		// ── Past / Completed ─────────────────────────────────────────────────
		{
			title:           "Cherry Blossom Hunt in Kyoto",
			destination:     "Kyoto, Japan",
			tripType:        "group",
			vibeTags:        []string{"culture", "nature", "photography"},
			budget:          2800,
			travelersPlanned: 3,
			startOffset:    -120,
			duration:        7,
			status:          trips.TripStatusCompleted,
			coverURL:        "https://images.unsplash.com/photo-1545569341-9eb8b30979d9?w=1200&q=80",
			notes:           "Book the ryokan in Gion well in advance – sold out months ahead during sakura season.",
			ownerIdx:        0,
		},
		{
			title:           "Amalfi Coast Road Trip",
			destination:     "Amalfi, Italy",
			tripType:        "couple",
			vibeTags:        []string{"beach", "romantic", "food"},
			budget:          3500,
			travelersPlanned: 2,
			startOffset:    -90,
			duration:        10,
			status:          trips.TripStatusCompleted,
			coverURL:        "https://images.unsplash.com/photo-1533587851505-d119e13fa0d7?w=1200&q=80",
			notes:           "Rent a scooter in Positano – best way to explore the cliffside villages.",
			ownerIdx:        1,
		},
		{
			title:           "Moroccan Desert Adventure",
			destination:     "Merzouga, Morocco",
			tripType:        "solo",
			vibeTags:        []string{"desert", "adventure", "culture"},
			budget:          1200,
			travelersPlanned: 1,
			startOffset:    -60,
			duration:        8,
			status:          trips.TripStatusCompleted,
			coverURL:        "https://images.unsplash.com/photo-1509316785289-025f5b846b35?w=1200&q=80",
			notes:           "Sleep under the stars at a Berber camp. Pure magic.",
			ownerIdx:        2,
		},
		{
			title:           "Safari in the Masai Mara",
			destination:     "Masai Mara, Kenya",
			tripType:        "group",
			vibeTags:        []string{"wildlife", "nature", "adventure"},
			budget:          5500,
			travelersPlanned: 4,
			startOffset:    -45,
			duration:        6,
			status:          trips.TripStatusCompleted,
			coverURL:        "https://images.unsplash.com/photo-1523805009345-7448845a9e53?w=1200&q=80",
			notes:           "July–October is migration season – absolutely worth it.",
			ownerIdx:        5,
		},
		{
			title:           "Northern Lights Chase in Iceland",
			destination:     "Reykjavik, Iceland",
			tripType:        "couple",
			vibeTags:        []string{"nature", "adventure", "photography"},
			budget:          4200,
			travelersPlanned: 2,
			startOffset:    -30,
			duration:        5,
			status:          trips.TripStatusCompleted,
			coverURL:        "https://images.unsplash.com/photo-1531366936337-7c912a4589a7?w=1200&q=80",
			notes:           "Stay outside the city for better aurora visibility. Vík is stunning.",
			ownerIdx:        6,
		},
		// ── Ongoing ───────────────────────────────────────────────────────────
		{
			title:           "Bali Slow Travel Month",
			destination:     "Ubud, Bali",
			tripType:        "solo",
			vibeTags:        []string{"wellness", "culture", "digital-nomad"},
			budget:          2000,
			travelersPlanned: 1,
			startOffset:    -10,
			duration:        30,
			status:          trips.TripStatusOngoing,
			coverURL:        "https://images.unsplash.com/photo-1518548419970-58e3b4079ab2?w=1200&q=80",
			notes:           "Co-working at Dojo Bali, yoga at sunrise, rice terrace walks in the evenings.",
			ownerIdx:        6,
		},
		{
			title:           "Trans-Siberian Rail Journey",
			destination:     "Moscow to Vladivostok, Russia",
			tripType:        "solo",
			vibeTags:        []string{"adventure", "trains", "culture"},
			budget:          3000,
			travelersPlanned: 1,
			startOffset:    -5,
			duration:        14,
			status:          trips.TripStatusOngoing,
			coverURL:        "https://images.unsplash.com/photo-1474487548417-781cb71495f3?w=1200&q=80",
			notes:           "Download offline maps and bring instant noodles for the long stretches.",
			ownerIdx:        2,
		},
		// ── Future / Planning ─────────────────────────────────────────────────
		{
			title:           "Greek Island Hopping 2025",
			destination:     "Santorini & Mykonos, Greece",
			tripType:        "group",
			vibeTags:        []string{"beach", "party", "food"},
			budget:          4800,
			travelersPlanned: 5,
			startOffset:    30,
			duration:        12,
			status:          trips.TripStatusPlanning,
			coverURL:        "https://images.unsplash.com/photo-1570077188670-e3a8d69ac5ff?w=1200&q=80",
			notes:           "Ferry tickets from Athens to Santorini – book early for summer.",
			ownerIdx:        0,
		},
		{
			title:           "Tokyo & Osaka Food Tour",
			destination:     "Tokyo & Osaka, Japan",
			tripType:        "couple",
			vibeTags:        []string{"food", "culture", "city"},
			budget:          3600,
			travelersPlanned: 2,
			startOffset:    45,
			duration:        9,
			status:          trips.TripStatusPlanning,
			coverURL:        "https://images.unsplash.com/photo-1540959733332-eab4deabeeaf?w=1200&q=80",
			notes:           "Must try: Osaka's dotonbori, Tokyo's hidden izakayas, Kyoto matcha everything.",
			ownerIdx:        4,
		},
		{
			title:           "Patagonia Trekking Expedition",
			destination:     "Torres del Paine, Chile",
			tripType:        "group",
			vibeTags:        []string{"trekking", "nature", "adventure"},
			budget:          5200,
			travelersPlanned: 4,
			startOffset:    60,
			duration:        10,
			status:          trips.TripStatusPlanning,
			coverURL:        "https://images.unsplash.com/photo-1501854140801-50d01698950b?w=1200&q=80",
			notes:           "W Trek or O Trek? Start booking campsites NOW – fills up insanely fast.",
			ownerIdx:        2,
		},
		{
			title:           "Lisbon & Porto Weekend Escape",
			destination:     "Lisbon & Porto, Portugal",
			tripType:        "couple",
			vibeTags:        []string{"city", "food", "history"},
			budget:          1800,
			travelersPlanned: 2,
			startOffset:    14,
			duration:        5,
			status:          trips.TripStatusPlanning,
			coverURL:        "https://images.unsplash.com/photo-1588345921523-c2dcdb7f1dcd?w=1200&q=80",
			notes:           "Pastéis de nata at Pastéis de Belém is non-negotiable.",
			ownerIdx:        1,
		},
		{
			title:           "New York City Winter Vibes",
			destination:     "New York City, USA",
			tripType:        "group",
			vibeTags:        []string{"city", "shopping", "nightlife"},
			budget:          6000,
			travelersPlanned: 3,
			startOffset:    90,
			duration:        6,
			status:          trips.TripStatusPlanning,
			coverURL:        "https://images.unsplash.com/photo-1496442226666-8d4d0e62e6e9?w=1200&q=80",
			notes:           "Stay in Brooklyn for better rates, subway everywhere.",
			ownerIdx:        3,
		},
		{
			title:           "Vietnam Street Food Trail",
			destination:     "Hanoi to Ho Chi Minh City, Vietnam",
			tripType:        "group",
			vibeTags:        []string{"food", "backpacker", "culture"},
			budget:          1400,
			travelersPlanned: 3,
			startOffset:    20,
			duration:        14,
			status:          trips.TripStatusPlanning,
			coverURL:        "https://images.unsplash.com/photo-1583417319070-4a69db38a482?w=1200&q=80",
			notes:           "Night train between cities is cheap and an experience in itself.",
			ownerIdx:        7,
		},
		{
			title:           "Scottish Highlands Road Trip",
			destination:     "Scottish Highlands, UK",
			tripType:        "couple",
			vibeTags:        []string{"nature", "road-trip", "history"},
			budget:          2200,
			travelersPlanned: 2,
			startOffset:    75,
			duration:        8,
			status:          trips.TripStatusPlanning,
			coverURL:        "https://images.unsplash.com/photo-1506905925346-21bda4d32df4?w=1200&q=80",
			notes:           "Glen Coe sunrise is unmissable. Wild camping is legal in Scotland!",
			ownerIdx:        7,
		},
		{
			title:           "Cape Town & Garden Route",
			destination:     "Cape Town, South Africa",
			tripType:        "group",
			vibeTags:        []string{"nature", "culture", "adventure"},
			budget:          3800,
			travelersPlanned: 4,
			startOffset:    110,
			duration:        11,
			status:          trips.TripStatusPlanning,
			coverURL:        "https://images.unsplash.com/photo-1580060839134-75a5edca2e99?w=1200&q=80",
			notes:           "Rent a car for the Garden Route – Knysna Heads are breathtaking.",
			ownerIdx:        5,
		},
	}

	var created []trips.Trip

	for _, r := range rows {
		start := now.AddDate(0, 0, r.startOffset)
		end := start.AddDate(0, 0, r.duration)
		notes := r.notes
		coverURL := r.coverURL

		owner := seedUsers[r.ownerIdx%len(seedUsers)]
		createdAt := now.AddDate(0, 0, r.startOffset-randomIntRange(1, 14))

		t := trips.Trip{
			ID:               uuid.New(),
			OwnerUserID:      owner.ID,
			Title:            r.title,
			Destination:      r.destination,
			TripType:         r.tripType,
			VibeTags:         pq.StringArray(r.vibeTags),
			TravelersPlanned: r.travelersPlanned,
			StartDate:        &start,
			EndDate:          &end,
			Budget:           r.budget,
			CoverImageURL:    &coverURL,
			Notes:            &notes,
			Status:           r.status,
			CreatedAt:        createdAt,
			UpdatedAt:        createdAt,
		}

		if err := db.Create(&t).Error; err != nil {
			return nil, fmt.Errorf("create trip %q: %w", r.title, err)
		}

		// Add owner as the trip's first member (owner role)
		joined := createdAt
		member := trips.TripMember{
			ID:         uuid.New(),
			TripID:     t.ID,
			UserID:     owner.ID,
			Role:       trips.RoleOwner,
			JoinStatus: trips.JoinStatusJoined,
			JoinedAt:   &joined,
			CreatedAt:  createdAt,
			UpdatedAt:  createdAt,
		}
		if err := db.Create(&member).Error; err != nil {
			log.Printf("[seed] warn: owner member for trip %q: %v", r.title, err)
		}

		// Add 1–2 extra members for group trips
		if r.tripType == "group" && len(seedUsers) > 1 {
			extraCount := randomIntRange(1, 3)
			added := map[uuid.UUID]bool{owner.ID: true}
			for range extraCount {
				u := seedUsers[randomIntRange(0, len(seedUsers))]
				if added[u.ID] {
					continue
				}
				added[u.ID] = true
				joinedAt := createdAt.Add(time.Duration(randomIntRange(1, 48)) * time.Hour)
				m := trips.TripMember{
					ID:         uuid.New(),
					TripID:     t.ID,
					UserID:     u.ID,
					Role:       trips.RoleMember,
					JoinStatus: trips.JoinStatusJoined,
					JoinedAt:   &joinedAt,
					CreatedAt:  joinedAt,
					UpdatedAt:  joinedAt,
				}
				db.Create(&m) //nolint:errcheck
			}
		}

		created = append(created, t)
	}

	log.Printf("[seed] ✅ created %d trips", len(created))
	return created, nil
}

func seededTripTitles() []string {
	return []string{
		"Cherry Blossom Hunt in Kyoto",
		"Amalfi Coast Road Trip",
		"Moroccan Desert Adventure",
		"Safari in the Masai Mara",
		"Northern Lights Chase in Iceland",
		"Bali Slow Travel Month",
		"Trans-Siberian Rail Journey",
		"Greek Island Hopping 2025",
		"Tokyo & Osaka Food Tour",
		"Patagonia Trekking Expedition",
		"Lisbon & Porto Weekend Escape",
		"New York City Winter Vibes",
		"Vietnam Street Food Trail",
		"Scottish Highlands Road Trip",
		"Cape Town & Garden Route",
	}
}
