package seed

import (
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/khadijayo/roamify/internal/modules/posts"
	"github.com/khadijayo/roamify/internal/modules/trips"
	"github.com/khadijayo/roamify/internal/modules/users"
	"gorm.io/gorm"
)

// SeedLikes creates 40 likes distributed across posts.
// Likes are modeled as PostLike records. Each (postID, userID) pair is unique.
// Because posts may not exist yet we fall back to seeding trip-based reactions
// (UserFollow relationships) when the post table is empty.
func SeedLikes(db *gorm.DB, seedUsers []users.User, seedTrips []trips.Trip) error {
	// Guard
	var likeCount int64
	db.Model(&posts.PostLike{}).Count(&likeCount)
	if likeCount >= 30 {
		log.Println("[seed] likes already seeded – skipping")
		return nil
	}

	// Check if there are any posts to attach likes to
	var postCount int64
	db.Model(&posts.Post{}).Count(&postCount)

	if postCount > 0 {
		return seedPostLikes(db, seedUsers)
	}

	// Fallback: seed user-follow relationships as a form of "like" engagement
	return seedUserFollows(db, seedUsers)
}

// seedPostLikes distributes 40 likes across existing posts.
func seedPostLikes(db *gorm.DB, seedUsers []users.User) error {
	var allPosts []posts.Post
	db.Select("id").Find(&allPosts)

	if len(allPosts) == 0 {
		return nil
	}

	target := 40
	created := 0
	now := time.Now()

	// Track (postID, userID) to avoid duplicates
	seen := make(map[string]bool)

	for range target {
		p := allPosts[randomIntRange(0, len(allPosts))]
		u := seedUsers[randomIntRange(0, len(seedUsers))]

		key := fmt.Sprintf("%s:%s", p.ID, u.ID)
		if seen[key] {
			continue
		}
		seen[key] = true

		daysAgo := randomIntRange(0, 60)
		createdAt := now.AddDate(0, 0, -daysAgo)

		like := posts.PostLike{
			ID:        uuid.New(),
			PostID:    p.ID,
			UserID:    u.ID,
			CreatedAt: createdAt,
		}

		if err := db.Create(&like).Error; err != nil {
			log.Printf("[seed] warn: post like: %v", err)
			continue
		}

		// Keep the post's likes_count in sync
		db.Model(&posts.Post{}).Where("id = ?", p.ID).
			UpdateColumn("likes_count", gorm.Expr("likes_count + 1"))

		created++
	}

	log.Printf("[seed] ✅ created %d post likes", created)
	return nil
}

// seedUserFollows creates follow relationships as an engagement proxy when no posts exist.
func seedUserFollows(db *gorm.DB, seedUsers []users.User) error {
	if len(seedUsers) < 2 {
		return nil
	}

	// Guard
	var count int64
	db.Table("user_follows").Count(&count)
	if count >= 20 {
		log.Println("[seed] user follows already seeded – skipping")
		return nil
	}

	type follow struct {
		followerIdx  int
		followingIdx int
	}

	// Build a realistic follow graph (40 edges)
	pairs := []follow{
		{0, 1}, {0, 2}, {0, 4}, {0, 5},
		{1, 0}, {1, 3}, {1, 6},
		{2, 0}, {2, 3}, {2, 7},
		{3, 0}, {3, 1}, {3, 4},
		{4, 0}, {4, 2}, {4, 6}, {4, 7},
		{5, 0}, {5, 2}, {5, 3}, {5, 6},
		{6, 0}, {6, 1}, {6, 4}, {6, 5},
		{7, 0}, {7, 2}, {7, 3}, {7, 5},
		{0, 6}, {1, 5}, {2, 6}, {3, 7},
		{4, 5}, {5, 7}, {6, 7}, {7, 6},
		{1, 7}, {2, 4}, {3, 5}, {4, 1},
	}

	now := time.Now()
	created := 0

	for _, p := range pairs {
		if p.followerIdx >= len(seedUsers) || p.followingIdx >= len(seedUsers) {
			continue
		}

		follower := seedUsers[p.followerIdx]
		following := seedUsers[p.followingIdx]

		if follower.ID == following.ID {
			continue
		}

		// Dedup
		var exists int64
		db.Table("user_follows").
			Where("follower_id = ? AND following_id = ?", follower.ID, following.ID).
			Count(&exists)
		if exists > 0 {
			continue
		}

		daysAgo := randomIntRange(0, 150)
		createdAt := now.AddDate(0, 0, -daysAgo)

		uf := users.UserFollow{
			ID:          uuid.New(),
			FollowerID:  follower.ID,
			FollowingID: following.ID,
			CreatedAt:   createdAt,
		}

		if err := db.Create(&uf).Error; err != nil {
			log.Printf("[seed] warn: user follow %d→%d: %v", p.followerIdx, p.followingIdx, err)
			continue
		}
		created++
	}

	log.Printf("[seed] ✅ created %d user-follow relationships (likes proxy)", created)
	return nil
}
