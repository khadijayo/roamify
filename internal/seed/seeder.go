package seed

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/khadijayo/roamify/internal/modules/challenges"
	"github.com/khadijayo/roamify/internal/modules/notifications"
	"github.com/khadijayo/roamify/internal/modules/posts"
	"github.com/khadijayo/roamify/internal/modules/trips"
	"github.com/khadijayo/roamify/internal/modules/users"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	seedEmailDomain = "@roamify.demo"
	demoPassword    = "Roamify2026!"
)

var rng = rand.New(rand.NewSource(time.Now().UnixNano())) //nolint:gosec

// TravelQuestion fills the product gap for presentation data. The app does not
// currently ship a questions module, so the seeder owns this small table.
type TravelQuestion struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID     uuid.UUID `gorm:"type:uuid;not null;index"                       json:"user_id"`
	Title      string    `gorm:"type:varchar(255);not null"                     json:"title"`
	Body       string    `gorm:"type:text;not null"                             json:"body"`
	Category   string    `gorm:"type:varchar(50);not null;index"                json:"category"`
	Location   string    `gorm:"type:varchar(120)"                              json:"location"`
	IsAnswered bool      `gorm:"default:false"                                  json:"is_answered"`
	CreatedAt  time.Time `                                                       json:"created_at"`
	UpdatedAt  time.Time `                                                       json:"updated_at"`
}

func (TravelQuestion) TableName() string { return "travel_questions" }

type FlightEntry struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Airline       string    `gorm:"type:varchar(120);not null"                     json:"airline"`
	DepartureCity string    `gorm:"type:varchar(120);not null;index"               json:"departure_city"`
	ArrivalCity   string    `gorm:"type:varchar(120);not null;index"               json:"arrival_city"`
	Price         float64   `gorm:"type:numeric(12,2);not null"                    json:"price"`
	CurrencyCode  string    `gorm:"type:varchar(3);default:'USD'"                  json:"currency_code"`
	DepartureTime time.Time `gorm:"not null"                                       json:"departure_time"`
	ArrivalTime   time.Time `gorm:"not null"                                       json:"arrival_time"`
	CreatedAt     time.Time `                                                      json:"created_at"`
	UpdatedAt     time.Time `                                                      json:"updated_at"`
}

func (FlightEntry) TableName() string { return "flight_entries" }

type HotelEntry struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name         string    `gorm:"type:varchar(180);not null"                     json:"name"`
	City         string    `gorm:"type:varchar(120);not null;index"               json:"city"`
	Country      string    `gorm:"type:varchar(120);not null;index"               json:"country"`
	Rating       float64   `gorm:"type:numeric(3,1);not null"                     json:"rating"`
	NightlyPrice float64   `gorm:"type:numeric(12,2);not null"                    json:"nightly_price"`
	CurrencyCode string    `gorm:"type:varchar(3);default:'USD'"                  json:"currency_code"`
	ImageURL     string    `gorm:"type:text"                                      json:"image_url"`
	CreatedAt    time.Time `                                                      json:"created_at"`
	UpdatedAt    time.Time `                                                      json:"updated_at"`
}

func (HotelEntry) TableName() string { return "hotel_entries" }

type Result struct {
	Users           int
	Trips           int
	Posts           int
	Questions       int
	Comments        int
	Challenges      int
	Participations  int
	TriviaQuestions int
	TriviaAttempts  int
	Likes           int
	Flights         int
	Hotels          int
	Notifications   int
}

type context struct {
	users      []users.User
	trips      []trips.Trip
	posts      []posts.Post
	challenges []challenges.Challenge
	trivia     []challenges.TriviaQuestion
}

// Run clears previous Roamify demo data and creates a fresh presentation-ready dataset.
func Run(db *gorm.DB) (*Result, error) {
	start := time.Now()
	result := &Result{}
	ctx := &context{}

	err := db.Transaction(func(tx *gorm.DB) error {
		if err := migrateSeederTables(tx); err != nil {
			return err
		}
		if err := clearSeedData(tx); err != nil {
			return err
		}

		var err error
		ctx.users, err = seedUsers(tx)
		if err != nil {
			return err
		}
		result.Users = len(ctx.users)

		ctx.trips, err = seedTrips(tx, ctx.users)
		if err != nil {
			return err
		}
		result.Trips = len(ctx.trips)

		ctx.posts, err = seedPosts(tx, ctx.users)
		if err != nil {
			return err
		}
		result.Posts = len(ctx.posts)

		result.Questions, err = seedQuestions(tx, ctx.users)
		if err != nil {
			return err
		}

		ctx.challenges, err = seedChallenges(tx)
		if err != nil {
			return err
		}
		result.Challenges = len(ctx.challenges)

		ctx.trivia, err = seedTriviaQuestions(tx)
		if err != nil {
			return err
		}
		result.TriviaQuestions = len(ctx.trivia)

		result.Participations, err = seedChallengeParticipations(tx, ctx.users, ctx.challenges)
		if err != nil {
			return err
		}

		result.TriviaAttempts, err = seedTriviaAttempts(tx, ctx.users, ctx.trivia)
		if err != nil {
			return err
		}

		result.Flights, err = seedFlights(tx)
		if err != nil {
			return err
		}

		result.Hotels, err = seedHotels(tx)
		if err != nil {
			return err
		}

		result.Comments, err = seedComments(tx, ctx.users, ctx.posts, ctx.trips)
		if err != nil {
			return err
		}

		result.Likes, err = seedLikes(tx, ctx.users, ctx.posts)
		if err != nil {
			return err
		}

		result.Notifications, err = seedNotifications(tx, ctx.users, ctx.trips, ctx.posts, ctx.challenges)
		return err
	})
	if err != nil {
		return nil, err
	}

	log.Printf("[seed] success in %s", time.Since(start).Round(time.Millisecond))
	log.Printf("[seed] inserted users=%d trips=%d posts=%d questions=%d comments=%d challenges=%d participations=%d trivia_questions=%d trivia_attempts=%d likes=%d flights=%d hotels=%d notifications=%d",
		result.Users, result.Trips, result.Posts, result.Questions, result.Comments, result.Challenges, result.Participations, result.TriviaQuestions, result.TriviaAttempts, result.Likes, result.Flights, result.Hotels, result.Notifications)

	return result, nil
}

func migrateSeederTables(db *gorm.DB) error {
	return db.AutoMigrate(&TravelQuestion{}, &FlightEntry{}, &HotelEntry{})
}

func clearSeedData(db *gorm.DB) error {
	var ids []uuid.UUID
	if err := db.Model(&users.User{}).Unscoped().Where("email LIKE ?", "%"+seedEmailDomain).Pluck("id", &ids).Error; err != nil {
		return fmt.Errorf("find seed users: %w", err)
	}

	if len(ids) > 0 {
		var postIDs []uuid.UUID
		if err := db.Model(&posts.Post{}).Unscoped().Where("author_user_id IN ?", ids).Pluck("id", &postIDs).Error; err != nil {
			return err
		}
		var tripIDs []uuid.UUID
		if err := db.Model(&trips.Trip{}).Unscoped().Where("owner_user_id IN ?", ids).Pluck("id", &tripIDs).Error; err != nil {
			return err
		}

		if err := db.Unscoped().Where("user_id IN ?", ids).Delete(&notifications.Notification{}).Error; err != nil {
			return err
		}
		if err := db.Unscoped().Where("user_id IN ?", ids).Delete(&notifications.UserNotificationSetting{}).Error; err != nil {
			return err
		}
		if err := db.Unscoped().Where("user_id IN ?", ids).Delete(&users.UserPrivacySetting{}).Error; err != nil {
			return err
		}
		if err := db.Unscoped().Where("user_id IN ?", ids).Delete(&users.VibeProfile{}).Error; err != nil {
			return err
		}
		if err := db.Unscoped().Where("follower_id IN ? OR following_id IN ?", ids, ids).Delete(&users.UserFollow{}).Error; err != nil {
			return err
		}
		if err := db.Unscoped().Where("user_id IN ?", ids).Delete(&posts.PostLike{}).Error; err != nil {
			return err
		}
		if err := db.Unscoped().Where("user_id IN ?", ids).Delete(&posts.PostComment{}).Error; err != nil {
			return err
		}
		if len(postIDs) > 0 {
			if err := db.Unscoped().Where("post_id IN ?", postIDs).Delete(&posts.PostTag{}).Error; err != nil {
				return err
			}
		}
		if err := db.Unscoped().Where("author_user_id IN ?", ids).Delete(&posts.Post{}).Error; err != nil {
			return err
		}
		if len(tripIDs) > 0 {
			if err := db.Unscoped().Where("trip_id IN ?", tripIDs).Delete(&trips.ChatMessage{}).Error; err != nil {
				return err
			}
			if err := db.Unscoped().Where("trip_id IN ?", tripIDs).Delete(&trips.TripExpense{}).Error; err != nil {
				return err
			}
			if err := db.Unscoped().Where("trip_id IN ?", tripIDs).Delete(&trips.TripItineraryItem{}).Error; err != nil {
				return err
			}
			if err := db.Unscoped().Where("trip_id IN ?", tripIDs).Delete(&trips.TripMember{}).Error; err != nil {
				return err
			}
		}
		if err := db.Unscoped().Where("user_id IN ?", ids).Delete(&trips.ChatMessage{}).Error; err != nil {
			return err
		}
		if err := db.Unscoped().Where("created_by_user_id IN ?", ids).Delete(&trips.TripExpense{}).Error; err != nil {
			return err
		}
		if err := db.Unscoped().Where("created_by_user_id IN ?", ids).Delete(&trips.TripItineraryItem{}).Error; err != nil {
			return err
		}
		if err := db.Unscoped().Where("user_id IN ?", ids).Delete(&trips.TripMember{}).Error; err != nil {
			return err
		}
		if err := db.Unscoped().Where("owner_user_id IN ?", ids).Delete(&trips.Trip{}).Error; err != nil {
			return err
		}
		if err := db.Unscoped().Where("user_id IN ?", ids).Delete(&TravelQuestion{}).Error; err != nil {
			return err
		}
		if err := db.Unscoped().Where("user_id IN ?", ids).Delete(&challenges.UserChallengeProgress{}).Error; err != nil {
			return err
		}
		if err := db.Unscoped().Where("user_id IN ?", ids).Delete(&challenges.TriviaAttempt{}).Error; err != nil {
			return err
		}
		if err := db.Unscoped().Where("id IN ?", ids).Delete(&users.User{}).Error; err != nil {
			return err
		}
	}

	if err := db.Exec("DELETE FROM post_tags WHERE post_id NOT IN (SELECT id FROM posts)").Error; err != nil {
		return err
	}
	if err := db.Exec("DELETE FROM trip_members WHERE trip_id NOT IN (SELECT id FROM trips)").Error; err != nil {
		return err
	}
	var seedChallengeIDs []uuid.UUID
	if err := db.Model(&challenges.Challenge{}).Where("title LIKE ?", "Seed Challenge:%").Pluck("id", &seedChallengeIDs).Error; err != nil {
		return err
	}
	if len(seedChallengeIDs) > 0 {
		if err := db.Unscoped().Where("challenge_id IN ?", seedChallengeIDs).Delete(&challenges.UserChallengeProgress{}).Error; err != nil {
			return err
		}
	}
	if err := db.Where("title LIKE ?", "Seed Challenge:%").Delete(&challenges.Challenge{}).Error; err != nil {
		return err
	}
	var seedTriviaIDs []uuid.UUID
	if err := db.Model(&challenges.TriviaQuestion{}).Where("question LIKE ?", "Seed Trivia:%").Pluck("id", &seedTriviaIDs).Error; err != nil {
		return err
	}
	if len(seedTriviaIDs) > 0 {
		if err := db.Unscoped().Where("trivia_question_id IN ?", seedTriviaIDs).Delete(&challenges.TriviaAttempt{}).Error; err != nil {
			return err
		}
	}
	if err := db.Where("question LIKE ?", "Seed Trivia:%").Delete(&challenges.TriviaQuestion{}).Error; err != nil {
		return err
	}
	if err := db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&FlightEntry{}).Error; err != nil {
		return err
	}
	return db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&HotelEntry{}).Error
}

func seedUsers(db *gorm.DB) ([]users.User, error) {
	type userRow struct {
		name, email, explorer string
		pace                  users.TravelPace
		budget                users.BudgetStyle
		with                  users.TravelWith
		vibes, interests      []string
		countries, level      int
		points                int
	}

	rows := []userRow{
		{"Amina Haddad", "amina.haddad@roamify.demo", "Culture Curious Beginner", users.PaceBalanced, users.BudgetBackpacker, users.WithSolo, []string{"culture", "markets"}, []string{"street food", "museums", "photography"}, 4, 1, 210},
		{"Lucas Martin", "lucas.martin@roamify.demo", "Weekend City Explorer", users.PaceChill, users.BudgetMidRange, users.WithPartner, []string{"city", "food"}, []string{"architecture", "wine", "walking tours"}, 12, 3, 780},
		{"Yuki Tanaka", "yuki.tanaka@roamify.demo", "Hidden Gem Hunter", users.PaceBalanced, users.BudgetMidRange, users.WithSolo, []string{"food", "minimalist"}, []string{"ramen", "onsen", "local trains"}, 18, 4, 1240},
		{"Sofia Larsson", "sofia.larsson@roamify.demo", "Digital Nomad", users.PaceChill, users.BudgetMidRange, users.WithSolo, []string{"coworking", "wellness"}, []string{"coffee", "yoga", "remote work"}, 25, 5, 1960},
		{"Kofi Mensah", "kofi.mensah@roamify.demo", "Adventure Planner", users.PaceAdventure, users.BudgetBackpacker, users.WithSquad, []string{"nature", "music"}, []string{"festivals", "hiking", "local transport"}, 16, 4, 1350},
		{"Priya Nair", "priya.nair@roamify.demo", "Solo Trail Seeker", users.PaceAdventure, users.BudgetBackpacker, users.WithSolo, []string{"trekking", "wildlife"}, []string{"camping", "wildlife", "trains"}, 32, 7, 3120},
		{"Marco Esposito", "marco.esposito@roamify.demo", "Slow Travel Foodie", users.PaceChill, users.BudgetMidRange, users.WithPartner, []string{"history", "cafe"}, []string{"espresso", "pasta", "old towns"}, 21, 5, 1840},
		{"Nadia Benali", "nadia.benali@roamify.demo", "Budget Backpacker", users.PaceBalanced, users.BudgetBackpacker, users.WithSquad, []string{"beach", "hostels"}, []string{"surfing", "night markets", "budget tips"}, 9, 2, 520},
		{"Maya Johnson", "maya.johnson@roamify.demo", "Luxury Escape Curator", users.PaceChill, users.BudgetLuxury, users.WithPartner, []string{"luxury", "spa"}, []string{"boutique hotels", "fine dining", "rooftops"}, 28, 6, 2400},
		{"Omar El-Sayed", "omar.elsayed@roamify.demo", "Desert Road Tripper", users.PaceAdventure, users.BudgetMidRange, users.WithSquad, []string{"desert", "road-trip"}, []string{"camping", "4x4 routes", "local guides"}, 14, 3, 930},
		{"Emma O'Connor", "emma.oconnor@roamify.demo", "First-Time Flyer", users.PaceBalanced, users.BudgetBackpacker, users.WithFamily, []string{"safe", "family"}, []string{"packing", "airport tips", "family stays"}, 3, 1, 160},
		{"Diego Alvarez", "diego.alvarez@roamify.demo", "Festival Hopper", users.PaceAdventure, users.BudgetMidRange, users.WithSquad, []string{"nightlife", "music"}, []string{"festivals", "street art", "tacos"}, 19, 4, 1420},
		{"Chen Wei", "chen.wei@roamify.demo", "Transit Expert", users.PaceBalanced, users.BudgetMidRange, users.WithSolo, []string{"city", "transport"}, []string{"metros", "rail passes", "maps"}, 27, 6, 2280},
		{"Leila Mansouri", "leila.mansouri@roamify.demo", "Museum Weekend Traveler", users.PaceChill, users.BudgetMidRange, users.WithPartner, []string{"art", "history"}, []string{"galleries", "bookshops", "cafes"}, 11, 3, 690},
		{"Noah Smith", "noah.smith@roamify.demo", "Photo Walk Regular", users.PaceBalanced, users.BudgetBackpacker, users.WithSolo, []string{"photography", "urban"}, []string{"street photos", "cheap eats", "hostels"}, 7, 2, 440},
		{"Hana Kim", "hana.kim@roamify.demo", "Wellness Wanderer", users.PaceChill, users.BudgetLuxury, users.WithPartner, []string{"wellness", "beach"}, []string{"retreats", "tea houses", "design hotels"}, 23, 5, 1740},
		{"Samira Khan", "samira.khan@roamify.demo", "Visa Research Pro", users.PaceBalanced, users.BudgetMidRange, users.WithFamily, []string{"planning", "culture"}, []string{"visa rules", "family routes", "food tours"}, 15, 4, 1110},
		{"Liam Murphy", "liam.murphy@roamify.demo", "Hostel Social Traveler", users.PaceAdventure, users.BudgetBackpacker, users.WithSquad, []string{"social", "backpacker"}, []string{"pub crawls", "hostel kitchens", "hiking"}, 30, 6, 2660},
		{"Isabella Rossi", "isabella.rossi@roamify.demo", "Romantic Escape Planner", users.PaceChill, users.BudgetLuxury, users.WithPartner, []string{"romantic", "food"}, []string{"boutique stays", "wine bars", "sunsets"}, 20, 5, 1580},
		{"Tariq Al-Farsi", "tariq.alfarsi@roamify.demo", "Aviation Deal Finder", users.PaceBalanced, users.BudgetMidRange, users.WithSolo, []string{"flights", "city"}, []string{"fare alerts", "airport lounges", "weekend trips"}, 35, 8, 3860},
	}

	hashBytes, err := bcrypt.GenerateFromPassword([]byte(demoPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash demo password: %w", err)
	}
	hash := string(hashBytes)
	now := time.Now()
	created := make([]users.User, 0, len(rows))

	for i, row := range rows {
		avatar := fmt.Sprintf("https://i.pravatar.cc/300?img=%d", i+12)
		authProvider := "email"
		createdAt := now.AddDate(0, -8, i*3)
		u := users.User{
			ID:           uuid.New(),
			Email:        ptr(row.email),
			FullName:     row.name,
			AvatarURL:    &avatar,
			PasswordHash: &hash,
			AuthProvider: &authProvider,
			Role:         users.RoleUser,
			Status:       users.StatusActive,
			IsVerified:   true,
			CreatedAt:    createdAt,
			UpdatedAt:    createdAt,
		}
		if err := db.Create(&u).Error; err != nil {
			return nil, fmt.Errorf("create user %s: %w", row.email, err)
		}
		profile := users.VibeProfile{
			ID:                 uuid.New(),
			UserID:             u.ID,
			ExplorerType:       row.explorer,
			PreferredVibes:     pq.StringArray(row.vibes),
			TravelPace:         row.pace,
			BudgetStyle:        row.budget,
			TravelWith:         row.with,
			Interests:          pq.StringArray(row.interests),
			OnboardingComplete: true,
			ExplorerLevel:      row.level,
			RoamifyPoints:      row.points,
			CountriesVisited:   row.countries,
			CreatedAt:          createdAt,
			UpdatedAt:          createdAt,
		}
		if err := db.Create(&profile).Error; err != nil {
			return nil, fmt.Errorf("create vibe profile %s: %w", row.email, err)
		}
		if err := db.Create(&users.UserPrivacySetting{ID: uuid.New(), UserID: u.ID, DataSharingEnabled: true, MapVisibility: "public", UpdatedAt: now}).Error; err != nil {
			return nil, err
		}
		if err := db.Create(&notifications.UserNotificationSetting{ID: uuid.New(), UserID: u.ID, TripRemindersEnabled: true, SquadUpdatesEnabled: true, PriceDropAlertsEnabled: i%3 == 0, UpdatedAt: now}).Error; err != nil {
			return nil, err
		}
		created = append(created, u)
	}
	log.Printf("[seed] users inserted: %d", len(created))
	return created, nil
}

func seedTrips(db *gorm.DB, seedUsers []users.User) ([]trips.Trip, error) {
	destinations := []struct {
		city, country, image, note string
		budget                     float64
		tags                       []string
	}{
		{"Paris", "France", "https://images.unsplash.com/photo-1502602898657-3e91760cbb34?w=1200&q=80", "Reserve small bistros near Canal Saint-Martin and leave one morning for the Louvre.", 2400, []string{"art", "food", "city"}},
		{"Tokyo", "Japan", "https://images.unsplash.com/photo-1540959733332-eab4deabeeaf?w=1200&q=80", "Mix Shibuya nights with quiet mornings in Yanaka.", 3200, []string{"food", "culture", "city"}},
		{"Rome", "Italy", "https://images.unsplash.com/photo-1529260830199-42c24126f198?w=1200&q=80", "Book Colosseum tickets early and save evenings for Trastevere.", 2100, []string{"history", "food", "walks"}},
		{"Barcelona", "Spain", "https://images.unsplash.com/photo-1583422409516-2895a77efded?w=1200&q=80", "Gaudi mornings, beach afternoons, tapas after sunset.", 2300, []string{"architecture", "beach", "nightlife"}},
		{"Istanbul", "Turkey", "https://images.unsplash.com/photo-1524231757912-21f4fe3a7200?w=1200&q=80", "Use ferries between continents and spend time in Kadikoy.", 1700, []string{"culture", "markets", "food"}},
		{"Dubai", "UAE", "https://images.unsplash.com/photo-1512453979798-5ea266f8880c?w=1200&q=80", "Pair old Dubai creek walks with one desert night.", 3600, []string{"luxury", "desert", "shopping"}},
		{"London", "United Kingdom", "https://images.unsplash.com/photo-1513635269975-59663e0ac1ad?w=1200&q=80", "Stay near a tube line and build rainy-day museum backups.", 2900, []string{"museums", "city", "theatre"}},
		{"New York", "USA", "https://images.unsplash.com/photo-1496442226666-8d4d0e62e6e9?w=1200&q=80", "Walk neighborhoods by borough and leave room for spontaneous food stops.", 3400, []string{"city", "food", "photography"}},
		{"Marrakech", "Morocco", "https://images.unsplash.com/photo-1597212618440-806262de4f6b?w=1200&q=80", "Hire a guide for the medina and plan one hammam afternoon.", 1500, []string{"markets", "culture", "desert"}},
		{"Bali", "Indonesia", "https://images.unsplash.com/photo-1518548419970-58e3b4079ab2?w=1200&q=80", "Balance Ubud rice terraces with a few slow beach days.", 2200, []string{"wellness", "beach", "nature"}},
	}

	now := time.Now()
	created := make([]trips.Trip, 0, 40)
	for i := 0; i < 40; i++ {
		d := destinations[i%len(destinations)]
		owner := seedUsers[i%len(seedUsers)]
		start := now.AddDate(0, 0, -120+i*7)
		end := start.AddDate(0, 0, 4+(i%8))
		status := trips.TripStatusPlanning
		if end.Before(now) {
			status = trips.TripStatusCompleted
		} else if start.Before(now) {
			status = trips.TripStatusOngoing
		}
		title := fmt.Sprintf("%s %s Escape", d.city, tripThemes()[i%len(tripThemes())])
		if i >= len(destinations) {
			title = fmt.Sprintf("%s %s Route %02d", d.city, tripThemes()[i%len(tripThemes())], i+1)
		}
		trip := trips.Trip{
			ID:               uuid.New(),
			OwnerUserID:      owner.ID,
			Title:            title,
			Destination:      d.city + ", " + d.country,
			TripType:         trips.TripType([]string{"solo", "couple", "group", "family"}[i%4]),
			VibeTags:         pq.StringArray(d.tags),
			TravelersPlanned: 1 + i%5,
			StartDate:        &start,
			EndDate:          &end,
			Budget:           d.budget + float64((i%6)*180),
			Spent:            float64((i % 5) * 220),
			CoverImageURL:    &d.image,
			Notes:            &d.note,
			Status:           status,
			CreatedAt:        start.AddDate(0, -1, 0),
			UpdatedAt:        now,
		}
		if err := db.Create(&trip).Error; err != nil {
			return nil, fmt.Errorf("create trip: %w", err)
		}
		if err := seedTripRelationships(db, trip, seedUsers, i); err != nil {
			return nil, err
		}
		created = append(created, trip)
	}
	log.Printf("[seed] trips inserted: %d", len(created))
	return created, nil
}

func seedTripRelationships(db *gorm.DB, trip trips.Trip, seedUsers []users.User, index int) error {
	now := time.Now()
	joined := trip.CreatedAt
	ownerMember := trips.TripMember{ID: uuid.New(), TripID: trip.ID, UserID: trip.OwnerUserID, Role: trips.RoleOwner, JoinStatus: trips.JoinStatusJoined, JoinedAt: &joined, CreatedAt: joined, UpdatedAt: joined}
	if err := db.Create(&ownerMember).Error; err != nil {
		return err
	}
	memberTarget := 1 + index%4
	seen := map[uuid.UUID]bool{trip.OwnerUserID: true}
	for j := 0; j < memberTarget; j++ {
		u := seedUsers[(index+j+3)%len(seedUsers)]
		if seen[u.ID] {
			continue
		}
		seen[u.ID] = true
		joinedAt := trip.CreatedAt.Add(time.Duration(j+1) * 8 * time.Hour)
		member := trips.TripMember{ID: uuid.New(), TripID: trip.ID, UserID: u.ID, Role: trips.RoleMember, JoinStatus: trips.JoinStatusJoined, JoinedAt: &joinedAt, CreatedAt: joinedAt, UpdatedAt: joinedAt}
		if err := db.Create(&member).Error; err != nil {
			return err
		}
	}
	for day := 1; day <= 2; day++ {
		when := trip.StartDate.Add(time.Duration(9+day*3) * time.Hour)
		item := trips.TripItineraryItem{
			ID: tripUUID(), TripID: trip.ID, DayNumber: day, Title: itineraryTitles()[(index+day)%len(itineraryTitles())],
			ItemType: trips.ItemType([]string{"activity", "food", "hotel"}[(index+day)%3]), PeopleCount: trip.TravelersPlanned,
			StartTime: &when, LocationName: trip.Destination, Notes: ptr("Seeded demo itinerary item"), SortOrder: day,
			CreatedByUserID: &trip.OwnerUserID, CreatedAt: now, UpdatedAt: now,
		}
		if err := db.Create(&item).Error; err != nil {
			return err
		}
	}
	expense := trips.TripExpense{
		ID: tripUUID(), TripID: trip.ID, CreatedByUserID: &trip.OwnerUserID, Description: "Hotel deposit",
		LocationName: trip.Destination, Category: "accommodation", Amount: trip.Budget * 0.18, ExpenseDate: trip.CreatedAt,
		CurrencyCode: "USD", CreatedAt: now, UpdatedAt: now,
	}
	return db.Create(&expense).Error
}

func seedPosts(db *gorm.DB, seedUsers []users.User) ([]posts.Post, error) {
	locations := []string{"Paris", "Tokyo", "Rome", "Barcelona", "Istanbul", "Dubai", "London", "New York", "Marrakech", "Bali"}
	contents := []string{
		"Found a quiet street two blocks from the main square with better coffee and half the prices.",
		"Budget tip: book the first train after breakfast instead of the rush-hour departure.",
		"The best food discovery was a tiny family place with no English menu and perfect soup.",
		"Packing win: one merino layer handled chilly flights and late rooftop dinners.",
		"Hidden gem alert: sunset from the residential side beat every crowded viewpoint.",
		"Destination recommendation: stay near transit, then spend your money on experiences.",
		"Street markets are still my favorite way to understand a city in the first hour.",
		"Solo travel note: joining a morning walking tour made the whole city feel easier.",
		"Splurge where it counts: one great hotel night after a red-eye flight changed everything.",
		"Food route idea: pick one neighborhood and snack through it slowly instead of rushing landmarks.",
	}
	tags := [][]string{{"budget", "tips"}, {"hidden-gem", "food"}, {"packing", "advice"}, {"recommendation", "city"}, {"solo", "safety"}}
	created := make([]posts.Post, 0, 20)
	for i := 0; i < 20; i++ {
		image := fmt.Sprintf("https://picsum.photos/seed/roamify-post-%02d/900/700", i+1)
		post := posts.Post{
			ID: uuid.New(), AuthorUserID: seedUsers[i%len(seedUsers)].ID, Content: contents[i%len(contents)],
			Location: locations[i%len(locations)], ImageURL: &image, Visibility: posts.VisibilityPublic,
			CreatedAt: time.Now().AddDate(0, 0, -i*2), UpdatedAt: time.Now().AddDate(0, 0, -i*2),
		}
		if err := db.Create(&post).Error; err != nil {
			return nil, err
		}
		for _, tag := range tags[i%len(tags)] {
			if err := db.Create(&posts.PostTag{ID: uuid.New(), PostID: post.ID, Tag: tag, CreatedAt: post.CreatedAt}).Error; err != nil {
				return nil, err
			}
		}
		created = append(created, post)
	}
	log.Printf("[seed] posts inserted: %d", len(created))
	return created, nil
}

func seedQuestions(db *gorm.DB, seedUsers []users.User) (int, error) {
	categories := []string{"visa", "budget", "transportation", "accommodation", "safety"}
	locations := []string{"Paris", "Tokyo", "Rome", "Barcelona", "Istanbul", "Dubai", "London", "New York", "Marrakech", "Bali"}
	titles := []string{
		"Do I need printed hotel bookings for a tourist visa?",
		"What is a realistic daily budget for food and transit?",
		"Is the airport train better than a taxi after midnight?",
		"Which neighborhood is safest for a first visit?",
		"Are hostels reliable for leaving luggage during the day?",
		"How many days should I keep free from pre-booked activities?",
		"Can I use one transit card across the full metro network?",
		"What should I avoid around the main tourist square?",
	}
	for i := 0; i < 40; i++ {
		q := TravelQuestion{
			ID: uuid.New(), UserID: seedUsers[(i*3)%len(seedUsers)].ID, Title: titles[i%len(titles)],
			Body:     fmt.Sprintf("Planning a trip to %s and looking for recent practical advice from travelers who went this year.", locations[i%len(locations)]),
			Category: categories[i%len(categories)], Location: locations[i%len(locations)], IsAnswered: i%3 != 0,
			CreatedAt: time.Now().AddDate(0, 0, -i), UpdatedAt: time.Now().AddDate(0, 0, -i),
		}
		if err := db.Create(&q).Error; err != nil {
			return i, err
		}
	}
	log.Printf("[seed] questions inserted: %d", 40)
	return 40, nil
}

func seedChallenges(db *gorm.DB) ([]challenges.Challenge, error) {
	base := []string{"Visit 3 countries in one year", "Share 10 travel tips", "Complete a solo trip", "Try 5 local foods", "Upload 20 travel photos"}
	extras := []string{"Take a sunrise photo walk", "Use public transport in a new city", "Book a locally owned stay", "Learn 20 phrases", "Plan a no-car weekend"}
	created := make([]challenges.Challenge, 0, 50)
	for i := 0; i < 50; i++ {
		name := base[i%len(base)]
		if i >= len(base) {
			name = extras[i%len(extras)]
		}
		c := challenges.Challenge{
			ID: uuid.New(), Title: fmt.Sprintf("Seed Challenge: %s #%02d", name, i+1),
			Description: fmt.Sprintf("Complete this travel goal and log your progress in Roamify. Theme: %s.", name),
			Category:    []challenges.ChallengeCategory{challenges.CategoryExplorer, challenges.CategorySocial, challenges.CategoryCollection, challenges.CategoryGamification}[i%4],
			Difficulty:  []challenges.DifficultyLevel{challenges.DifficultyEasy, challenges.DifficultyMedium, challenges.DifficultyHard}[i%3],
			Points:      80 + (i%7)*25, IsActive: true, CreatedAt: time.Now().AddDate(0, -3, i), UpdatedAt: time.Now(),
		}
		if err := db.Create(&c).Error; err != nil {
			return nil, err
		}
		created = append(created, c)
	}
	log.Printf("[seed] challenges inserted: %d", len(created))
	return created, nil
}

func seedChallengeParticipations(db *gorm.DB, seedUsers []users.User, seedChallenges []challenges.Challenge) (int, error) {
	for i := 0; i < 15; i++ {
		accepted := time.Now().AddDate(0, 0, -20+i)
		var completed *time.Time
		status := challenges.StatusAccepted
		points := 0
		if i%2 == 0 {
			done := accepted.AddDate(0, 0, 5+i%3)
			completed = &done
			status = challenges.StatusCompleted
			points = seedChallenges[i%len(seedChallenges)].Points
		}
		row := challenges.UserChallengeProgress{ID: uuid.New(), UserID: seedUsers[i%len(seedUsers)].ID, ChallengeID: seedChallenges[i%len(seedChallenges)].ID, Status: status, AwardedPoints: points, AcceptedAt: accepted, CompletedAt: completed, CreatedAt: accepted, UpdatedAt: time.Now()}
		if err := db.Create(&row).Error; err != nil {
			return i, err
		}
	}
	log.Printf("[seed] challenge participations inserted: %d", 15)
	return 15, nil
}

func seedTriviaQuestions(db *gorm.DB) ([]challenges.TriviaQuestion, error) {
	type triviaRow struct {
		question string
		choices  []string
		answer   string
		points   int
	}

	rows := []triviaRow{
		{"Which city is served by Narita and Haneda airports?", []string{"Tokyo", "Seoul", "Osaka", "Taipei"}, "Tokyo", 60},
		{"What document is usually required for international air travel?", []string{"Passport", "Library card", "Gym pass", "Hotel receipt"}, "Passport", 50},
		{"Which city is famous for the Colosseum?", []string{"Rome", "Paris", "Dubai", "Marrakech"}, "Rome", 50},
		{"What is the local currency used in Japan?", []string{"Yen", "Euro", "Dollar", "Dirham"}, "Yen", 60},
		{"Which city is known for La Sagrada Familia?", []string{"Barcelona", "London", "Istanbul", "Bali"}, "Barcelona", 55},
		{"What should travelers check before a visa appointment?", []string{"Passport validity", "Shoe size", "Playlist length", "Hotel paint color"}, "Passport validity", 70},
		{"Which city has the Grand Bazaar?", []string{"Istanbul", "New York", "Rome", "Paris"}, "Istanbul", 55},
		{"What is a common way to save money in expensive cities?", []string{"Use public transport", "Book only taxis", "Avoid maps", "Skip breakfast always"}, "Use public transport", 50},
		{"Which destination is known for riads and souks?", []string{"Marrakech", "Tokyo", "London", "Dubai"}, "Marrakech", 55},
		{"What should you pack in carry-on for long flights?", []string{"Essential medication", "Full-size shampoo", "Kitchen knives", "Loose paint"}, "Essential medication", 65},
		{"Which city is famous for the Eiffel Tower?", []string{"Paris", "Rome", "Barcelona", "Dubai"}, "Paris", 50},
		{"Which island destination is known for Ubud rice terraces?", []string{"Bali", "Santorini", "Malta", "Ibiza"}, "Bali", 55},
		{"Which city has boroughs including Brooklyn and Manhattan?", []string{"New York", "London", "Paris", "Tokyo"}, "New York", 50},
		{"What is the safest habit when arriving late?", []string{"Use verified transport", "Accept random rides", "Keep phone dead", "Ignore address"}, "Use verified transport", 70},
		{"Which city is known for the Burj Khalifa?", []string{"Dubai", "Istanbul", "Rome", "Marrakech"}, "Dubai", 50},
		{"What helps avoid roaming charges?", []string{"eSIM or local SIM", "Airplane mode forever", "More photos", "Paper tickets only"}, "eSIM or local SIM", 55},
		{"Which city is associated with Big Ben?", []string{"London", "Paris", "Rome", "Tokyo"}, "London", 50},
		{"What is a smart hostel safety habit?", []string{"Use a locker", "Leave passport on bed", "Share PINs", "Ignore reviews"}, "Use a locker", 60},
		{"Which food is strongly associated with Rome?", []string{"Carbonara", "Sushi", "Couscous", "Tacos"}, "Carbonara", 45},
		{"What should budget travelers compare before booking?", []string{"Total price with fees", "Logo color", "Lobby music", "Elevator brand"}, "Total price with fees", 60},
		{"Which transport pass is useful for many city trips?", []string{"Metro card", "Cinema ticket", "Gym card", "Concert wristband"}, "Metro card", 45},
		{"What is a good first step in a new city?", []string{"Save offline maps", "Delete itinerary", "Lose charger", "Skip check-in"}, "Save offline maps", 50},
		{"Which city sits on two continents?", []string{"Istanbul", "Paris", "New York", "Rome"}, "Istanbul", 65},
		{"What is useful for proof of accommodation?", []string{"Booking confirmation", "Restaurant photo", "Old receipt", "Boarding music"}, "Booking confirmation", 50},
		{"Which city is famous for tapas and Gaudi architecture?", []string{"Barcelona", "Tokyo", "Dubai", "London"}, "Barcelona", 55},
		{"What reduces missed-flight risk?", []string{"Arrive early", "Ignore gate changes", "Pack late", "Use no alarms"}, "Arrive early", 55},
		{"Which country is Marrakech in?", []string{"Morocco", "Turkey", "Italy", "Indonesia"}, "Morocco", 45},
		{"What helps with food allergies abroad?", []string{"Translated allergy card", "Guessing ingredients", "Eating anything", "No water"}, "Translated allergy card", 70},
		{"Which city is known for Shibuya Crossing?", []string{"Tokyo", "Seoul", "Bangkok", "Osaka"}, "Tokyo", 50},
		{"What should you check before booking budget flights?", []string{"Baggage rules", "Pilot's favorite color", "Seat fabric", "Airport playlist"}, "Baggage rules", 60},
	}

	created := make([]challenges.TriviaQuestion, 0, len(rows))
	now := time.Now()
	for i, row := range rows {
		q := challenges.TriviaQuestion{
			ID:            uuid.New(),
			Question:      "Seed Trivia: " + row.question,
			Choices:       pq.StringArray(row.choices),
			CorrectAnswer: row.answer,
			Points:        row.points,
			IsActive:      true,
			CreatedAt:     now.AddDate(0, 0, -i),
			UpdatedAt:     now,
		}
		if err := db.Create(&q).Error; err != nil {
			return nil, err
		}
		created = append(created, q)
	}
	log.Printf("[seed] trivia questions inserted: %d", len(created))
	return created, nil
}

func seedTriviaAttempts(db *gorm.DB, seedUsers []users.User, seedTrivia []challenges.TriviaQuestion) (int, error) {
	if len(seedUsers) == 0 || len(seedTrivia) == 0 {
		return 0, nil
	}

	for i := 0; i < 30; i++ {
		q := seedTrivia[i%len(seedTrivia)]
		correct := i%4 != 0
		selected := q.CorrectAnswer
		if !correct && len(q.Choices) > 1 {
			selected = q.Choices[(i+1)%len(q.Choices)]
			if selected == q.CorrectAnswer {
				selected = q.Choices[(i+2)%len(q.Choices)]
			}
		}

		awarded := 0
		if correct {
			awarded = q.Points
		}
		attempt := challenges.TriviaAttempt{
			ID:               uuid.New(),
			UserID:           seedUsers[(i*2)%len(seedUsers)].ID,
			TriviaQuestionID: q.ID,
			SelectedAnswer:   selected,
			IsCorrect:        correct,
			AwardedPoints:    awarded,
			CreatedAt:        time.Now().AddDate(0, 0, -i),
		}
		if err := db.Create(&attempt).Error; err != nil {
			return i, err
		}
	}
	log.Printf("[seed] trivia attempts inserted: %d", 30)
	return 30, nil
}

func seedFlights(db *gorm.DB) (int, error) {
	airlines := []string{"Air France", "Japan Airlines", "ITA Airways", "Vueling", "Turkish Airlines", "Emirates", "British Airways", "Delta", "Royal Air Maroc", "Qatar Airways"}
	routes := [][2]string{{"New York", "Paris"}, {"Los Angeles", "Tokyo"}, {"London", "Rome"}, {"Paris", "Barcelona"}, {"Berlin", "Istanbul"}, {"Mumbai", "Dubai"}, {"Dublin", "London"}, {"Chicago", "New York"}, {"Madrid", "Marrakech"}, {"Singapore", "Bali"}}
	for i := 0; i < 40; i++ {
		dep := time.Now().AddDate(0, 0, 14+i).Add(time.Duration(6+i%12) * time.Hour)
		row := FlightEntry{ID: uuid.New(), Airline: airlines[i%len(airlines)], DepartureCity: routes[i%len(routes)][0], ArrivalCity: routes[i%len(routes)][1], Price: 120 + float64((i%15)*37), CurrencyCode: "USD", DepartureTime: dep, ArrivalTime: dep.Add(time.Duration(2+i%11) * time.Hour), CreatedAt: time.Now(), UpdatedAt: time.Now()}
		if err := db.Create(&row).Error; err != nil {
			return i, err
		}
	}
	log.Printf("[seed] flights inserted: %d", 40)
	return 40, nil
}

func seedHotels(db *gorm.DB) (int, error) {
	cities := [][2]string{{"Paris", "France"}, {"Tokyo", "Japan"}, {"Rome", "Italy"}, {"Barcelona", "Spain"}, {"Istanbul", "Turkey"}, {"Dubai", "UAE"}, {"London", "United Kingdom"}, {"New York", "USA"}, {"Marrakech", "Morocco"}, {"Bali", "Indonesia"}}
	styles := []string{"Garden House", "Central Rooms", "Riverside Hotel", "Old Town Suites", "Market View Inn", "Nomad Boutique", "Harbor Stay", "Grand Terrace"}
	for i := 0; i < 40; i++ {
		city := cities[i%len(cities)]
		row := HotelEntry{ID: uuid.New(), Name: fmt.Sprintf("%s %s", city[0], styles[i%len(styles)]), City: city[0], Country: city[1], Rating: 3.7 + float64(i%13)/10, NightlyPrice: 55 + float64((i%18)*22), CurrencyCode: "USD", ImageURL: fmt.Sprintf("https://picsum.photos/seed/roamify-hotel-%02d/900/600", i+1), CreatedAt: time.Now(), UpdatedAt: time.Now()}
		if err := db.Create(&row).Error; err != nil {
			return i, err
		}
	}
	log.Printf("[seed] hotels inserted: %d", 40)
	return 40, nil
}

func seedComments(db *gorm.DB, seedUsers []users.User, seedPosts []posts.Post, seedTrips []trips.Trip) (int, error) {
	comments := []string{"This is exactly the kind of detail I needed.", "Saving this for my itinerary.", "How early did you book this?", "The budget note is super useful.", "Adding this place to my map.", "I went last spring and agree completely.", "Do you think this works for a solo traveler?", "That food tip sounds amazing.", "Public transport there was easier than expected.", "Great reminder to leave buffer time."}
	for i := 0; i < 30; i++ {
		row := posts.PostComment{ID: uuid.New(), PostID: seedPosts[i%len(seedPosts)].ID, UserID: seedUsers[(i+5)%len(seedUsers)].ID, Content: comments[i%len(comments)], CreatedAt: time.Now().AddDate(0, 0, -i), UpdatedAt: time.Now().AddDate(0, 0, -i)}
		if err := db.Create(&row).Error; err != nil {
			return i, err
		}
	}
	for i := 0; i < 12 && i < len(seedTrips); i++ {
		msg := trips.ChatMessage{ID: uuid.New(), TripID: seedTrips[i].ID, UserID: seedUsers[(i+2)%len(seedUsers)].ID, Message: "I updated the plan and added a flexible backup activity.", CreatedAt: time.Now().AddDate(0, 0, -i)}
		if err := db.Create(&msg).Error; err != nil {
			return 30, err
		}
	}
	log.Printf("[seed] comments inserted: %d", 30)
	return 30, nil
}

func seedLikes(db *gorm.DB, seedUsers []users.User, seedPosts []posts.Post) (int, error) {
	created := 0
	seen := make(map[string]bool)
	for created < 30 {
		post := seedPosts[(created*2+3)%len(seedPosts)]
		user := seedUsers[(created*5+1)%len(seedUsers)]
		if post.AuthorUserID == user.ID {
			user = seedUsers[(created*5+2)%len(seedUsers)]
		}
		key := post.ID.String() + ":" + user.ID.String()
		if seen[key] {
			created++
			continue
		}
		seen[key] = true
		like := posts.PostLike{ID: uuid.New(), PostID: post.ID, UserID: user.ID, CreatedAt: time.Now().AddDate(0, 0, -created)}
		if err := db.Create(&like).Error; err != nil {
			return created, err
		}
		if err := db.Model(&posts.Post{}).Where("id = ?", post.ID).UpdateColumn("likes_count", gorm.Expr("likes_count + 1")).Error; err != nil {
			return created, err
		}
		created++
	}
	log.Printf("[seed] likes inserted: %d", created)
	return created, nil
}

func seedNotifications(db *gorm.DB, seedUsers []users.User, seedTrips []trips.Trip, seedPosts []posts.Post, seedChallenges []challenges.Challenge) (int, error) {
	types := []notifications.NotificationType{notifications.NotifTripInvite, notifications.NotifTripStatusChanged, notifications.NotifMemberJoined, notifications.NotifPostLiked, notifications.NotifChallengeCompleted, notifications.NotifNewFollower, notifications.NotifChatMessage, notifications.NotifPriceDrop}
	titles := []string{"Trip invite received", "Trip status updated", "New member joined", "Your post got a like", "Challenge completed", "New follower", "New trip chat message", "Price drop found"}
	for i := 0; i < 40; i++ {
		refType := "trip"
		refID := seedTrips[i%len(seedTrips)].ID
		if i%4 == 1 {
			refType = "post"
			refID = seedPosts[i%len(seedPosts)].ID
		} else if i%4 == 2 {
			refType = "challenge"
			refID = seedChallenges[i%len(seedChallenges)].ID
		}
		row := notifications.Notification{ID: uuid.New(), UserID: seedUsers[i%len(seedUsers)].ID, Type: types[i%len(types)], Title: titles[i%len(titles)], Body: "Demo activity for your Roamify presentation feed.", RefID: &refID, RefType: &refType, IsRead: i%3 == 0, CreatedAt: time.Now().AddDate(0, 0, -i)}
		if err := db.Create(&row).Error; err != nil {
			return i, err
		}
	}
	log.Printf("[seed] notifications inserted: %d", 40)
	return 40, nil
}

func tripThemes() []string {
	return []string{"Food", "Hidden Gems", "Budget", "Museum", "Weekend", "Photo", "Family", "Solo"}
}

func itineraryTitles() []string {
	return []string{"Old town walking route", "Local food tasting", "Museum and cafe break", "Sunset viewpoint", "Market morning", "Neighborhood photo walk"}
}

func tripUUID() uuid.UUID {
	return uuid.New()
}

func ptr[T any](v T) *T {
	return &v
}

func randomIntRange(min, max int) int {
	if max <= min {
		return min
	}
	return min + rng.Intn(max-min)
}
