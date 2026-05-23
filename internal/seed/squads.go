package seed

import (
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/khadijayo/roamify/internal/modules/trips"
	"github.com/khadijayo/roamify/internal/modules/users"
	"gorm.io/gorm"
)

// Squad is a lightweight struct representing a travel squad.
// Roamify uses Trip + TripMember as the underlying data model for squads.
// We model squads as multi-member group trips so they integrate natively with the API.
type Squad struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name        string    `gorm:"type:varchar(255);not null"                    json:"name"`
	Description string    `gorm:"type:text"                                     json:"description"`
	CoverURL    string    `gorm:"type:text"                                     json:"cover_url"`
	OwnerID     uuid.UUID `gorm:"type:uuid;not null"                            json:"owner_id"`
	CreatedAt   time.Time `                                                     json:"created_at"`
	UpdatedAt   time.Time `                                                     json:"updated_at"`
}

func (Squad) TableName() string { return "squads" }

// SeedSquads creates 4 squads by adding group trip records that model squad relationships.
// If the squads table does not exist in the schema this seeder creates Trip-based squad groups instead.
func SeedSquads(db *gorm.DB, seedUsers []users.User, seedTrips []trips.Trip) error {
	if len(seedUsers) < 2 {
		return fmt.Errorf("seedSquads requires at least 2 users")
	}

	// Try inserting into squads table (if it exists), otherwise fall back to trip-based squads
	tableExists := db.Migrator().HasTable("squads")

	if tableExists {
		return seedSquadsTable(db, seedUsers)
	}

	// Fallback: create additional group trip memberships that simulate squads
	return seedSquadMemberships(db, seedUsers, seedTrips)
}

func seedSquadsTable(db *gorm.DB, seedUsers []users.User) error {
	var count int64
	db.Table("squads").Count(&count)
	if count > 0 {
		log.Println("[seed] squads already seeded – skipping")
		return nil
	}

	type squadDef struct {
		name        string
		description string
		coverURL    string
		ownerIdx    int
		memberIdxs  []int
	}

	squads := []squadDef{
		{
			name:        "Sunset Chasers",
			description: "We fly wherever the sunsets are golden. Beach lovers, rooftop addicts, horizon worshippers.",
			coverURL:    "https://images.unsplash.com/photo-1507525428034-b723cf961d3e?w=1200&q=80",
			ownerIdx:    0,
			memberIdxs:  []int{1, 3, 6},
		},
		{
			name:        "Trail Blazers",
			description: "Trekking, camping and off-grid adventures. Leave only footprints.",
			coverURL:    "https://images.unsplash.com/photo-1551632811-561732d1e306?w=1200&q=80",
			ownerIdx:    2,
			memberIdxs:  []int{5, 7},
		},
		{
			name:        "Nomad Collective",
			description: "Remote workers who co-work from every time zone. Laptop + passport = office.",
			coverURL:    "https://images.unsplash.com/photo-1522202176988-66273c2fd55f?w=1200&q=80",
			ownerIdx:    6,
			memberIdxs:  []int{0, 4},
		},
		{
			name:        "Budget Wanderers",
			description: "Proof that amazing trips don't require deep pockets. Hostels, trains & street food.",
			coverURL:    "https://images.unsplash.com/photo-1527856263669-12c3a0af2aa6?w=1200&q=80",
			ownerIdx:    7,
			memberIdxs:  []int{2, 5},
		},
	}

	now := time.Now()
	for i, s := range squads {
		owner := seedUsers[s.ownerIdx%len(seedUsers)]
		createdAt := now.AddDate(0, -5, 0).Add(time.Duration(i*20*24) * time.Hour)

		sq := Squad{
			ID:          uuid.New(),
			Name:        s.name,
			Description: s.description,
			CoverURL:    s.coverURL,
			OwnerID:     owner.ID,
			CreatedAt:   createdAt,
			UpdatedAt:   createdAt,
		}

		if err := db.Create(&sq).Error; err != nil {
			return fmt.Errorf("create squad %q: %w", s.name, err)
		}
	}

	log.Printf("[seed] ✅ created %d squads (squads table)", len(squads))
	return nil
}

// seedSquadMemberships simulates squads by ensuring selected users are members of the same group trips.
func seedSquadMemberships(db *gorm.DB, seedUsers []users.User, seedTrips []trips.Trip) error {
	type squadGroup struct {
		name       string
		tripIdx    int
		memberIdxs []int
	}

	groups := []squadGroup{
		{name: "Sunset Chasers", tripIdx: 7, memberIdxs: []int{0, 1, 3, 6}},
		{name: "Trail Blazers", tripIdx: 9, memberIdxs: []int{2, 5, 7}},
		{name: "Nomad Collective", tripIdx: 5, memberIdxs: []int{0, 4, 6}},
		{name: "Budget Wanderers", tripIdx: 12, memberIdxs: []int{2, 5, 7}},
	}

	now := time.Now()
	for _, g := range groups {
		if g.tripIdx >= len(seedTrips) {
			continue
		}
		trip := seedTrips[g.tripIdx]

		for _, uIdx := range g.memberIdxs {
			if uIdx >= len(seedUsers) {
				continue
			}
			u := seedUsers[uIdx]

			// Don't duplicate memberships
			var exists int64
			db.Model(&trips.TripMember{}).
				Where("trip_id = ? AND user_id = ?", trip.ID, u.ID).
				Count(&exists)
			if exists > 0 {
				continue
			}

			joinedAt := now.AddDate(0, -3, randomIntRange(0, 60))
			m := trips.TripMember{
				ID:         uuid.New(),
				TripID:     trip.ID,
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

	log.Println("[seed] ✅ created 4 squad-style trip groups (via TripMember)")
	return nil
}
