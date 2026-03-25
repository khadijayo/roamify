package config

import (
	"log"

	"github.com/khadijayo/roamify/internal/modules/challenges"
	"github.com/khadijayo/roamify/internal/modules/notifications"
	"github.com/khadijayo/roamify/internal/modules/passport"
	"github.com/khadijayo/roamify/internal/modules/posts"
	"github.com/khadijayo/roamify/internal/modules/trips"
	"github.com/khadijayo/roamify/internal/modules/users"
	"github.com/khadijayo/roamify/internal/modules/wishlist"
)

func AutoMigrate() {
	err := DB.AutoMigrate(

		&users.User{},
		&users.VibeProfile{},
		&users.UserFollow{},
		&users.UserPrivacySetting{},
		&notifications.UserNotificationSetting{},

		&trips.Trip{},
		&trips.TripMember{},
		&trips.TripItineraryItem{},
		&trips.TripExpense{},
		&trips.ChatMessage{},

		&posts.Post{},
		&posts.PostTag{},
		&posts.PostLike{},

		&wishlist.WishlistItem{},
		&wishlist.WishlistCollection{},
		&wishlist.WishlistCollectionItem{},

		&challenges.Challenge{},
		&challenges.UserChallengeProgress{},
		&challenges.TriviaQuestion{},
		&challenges.TriviaAttempt{},

		&passport.PassportVaultRecord{},
		&passport.PassportStamp{},
	)
	if err != nil {
		log.Fatalf("[migrate] AutoMigrate failed: %v", err)
	}
	log.Println("[migrate] all tables migrated successfully")
}
