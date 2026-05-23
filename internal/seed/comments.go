package seed

import (
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/khadijayo/roamify/internal/modules/posts"
	"github.com/khadijayo/roamify/internal/modules/trips"
	"github.com/khadijayo/roamify/internal/modules/users"
	"gorm.io/gorm"
)

// commentPool is a bank of realistic travel comments to distribute randomly.
var commentPool = []string{
	"This looks absolutely incredible! Adding it to my bucket list right now 🙌",
	"I went there last year – the local food scene is unreal. Don't miss the night market!",
	"How did you handle visa requirements? Planning a similar trip for next year.",
	"The cover photo is stunning. Did you take that yourself or is it from the trip?",
	"We should do a collab trip! I've been eyeing this destination forever.",
	"Budget seems very reasonable for that duration. Any tips on saving on accommodation?",
	"The sunrise hike is 100% worth the early alarm. Best decision of my entire trip.",
	"Just got back from the same route! Happy to answer any questions 😊",
	"Can't wait to see how this unfolds. Following for updates!",
	"This gave me serious wanderlust. Booking flights this weekend lol.",
	"Hidden gem alert 🚨 Most people skip this and go straight to the tourist traps.",
	"The local guides there are amazing – highly recommend hiring one for a day.",
	"Pro tip: avoid the first week of August, it's insanely crowded. Go in September instead.",
	"I'm obsessed with this aesthetic. What camera setup are you using?",
	"This is literally what I needed to see today. Planning my escape as we speak.",
	"Group trips are the best! Squad goals honestly 🌍",
	"The budget breakdown would be so helpful to see. Care to share?",
	"I had the exact same experience – transformative doesn't even begin to cover it.",
	"Just shared this with my travel partner. We're in! Count us in for the next one.",
	"The accommodation you picked looks cozy af. How did you find it?",
	"Lived in this city for 3 years and there are still spots I haven't discovered. Incredible.",
	"Public transit here is a dream. You barely need to rent a car.",
	"The weather this time of year is perfection – you picked an amazing window.",
	"This is the kind of content that makes me want to quit my job and travel full time 😂",
	"Marked! Planning my itinerary around this recommendation. Thank you!",
}

// SeedComments creates 25 comments distributed across posts and trips.
// It seeds PostComments (attached to posts). If your schema also has trip-level
// comments you can extend this function using the same pattern.
func SeedComments(db *gorm.DB, seedUsers []users.User, seedTrips []trips.Trip) error {
	// Guard
	var count int64
	db.Model(&posts.PostComment{}).Count(&count)
	if count >= 20 {
		log.Println("[seed] comments already seeded – skipping")
		return nil
	}

	// We attach comments to trip chat messages OR create standalone post-level comments.
	// Since posts are typically created by users separately, we seed comments as
	// ChatMessage records on trips (which is the native social layer on trips).
	return seedTripChatMessages(db, seedUsers, seedTrips)
}

// seedTripChatMessages seeds 25 chat-style comments on trips using the ChatMessage model.
func seedTripChatMessages(db *gorm.DB, seedUsers []users.User, seedTrips []trips.Trip) error {
	if len(seedUsers) == 0 || len(seedTrips) == 0 {
		return nil
	}

	// Guard on ChatMessage table
	var count int64
	db.Model(&trips.ChatMessage{}).Count(&count)
	if count >= 20 {
		log.Println("[seed] chat messages already seeded – skipping")
		return nil
	}

	now := time.Now()
	target := 25
	created := 0

	for i := range target {
		trip := seedTrips[i%len(seedTrips)]
		user := seedUsers[randomIntRange(0, len(seedUsers))]
		content := commentPool[i%len(commentPool)]

		// Scatter timestamps naturally: 0–90 days ago
		daysAgo := randomIntRange(0, 90)
		hoursVariation := randomIntRange(0, 23)
		createdAt := now.AddDate(0, 0, -daysAgo).Add(-time.Duration(hoursVariation) * time.Hour)

		msg := trips.ChatMessage{
			ID:        uuid.New(),
			TripID:    trip.ID,
			UserID:    user.ID,
			Message:   content,
			CreatedAt: createdAt,
		}

		if err := db.Create(&msg).Error; err != nil {
			log.Printf("[seed] warn: comment %d: %v", i, err)
			continue
		}
		created++
	}

	log.Printf("[seed] ✅ created %d comments (trip chat messages)", created)
	return nil
}
