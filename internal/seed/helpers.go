package seed

import (
	"math/rand"
	"time"
)

// seededRand is seeded once at package init for reproducible-ish but varied data.
var seededRand = rand.New(rand.NewSource(time.Now().UnixNano())) //nolint:gosec

// randomIntRange returns a random int in [min, max).
func randomIntRange(min, max int) int {
	if max <= min {
		return min
	}
	return min + seededRand.Intn(max-min)
}

// randomFrom picks a random element from a slice.
func randomFrom[T any](s []T) T {
	return s[randomIntRange(0, len(s))]
}

// jitteredTime returns t offset by ±jitterHours hours randomly.
func jitteredTime(t time.Time, jitterHours int) time.Time {
	offset := time.Duration(randomIntRange(-jitterHours, jitterHours)) * time.Hour
	return t.Add(offset)
}

// ptr returns a pointer to any value – handy for optional GORM string fields.
func ptr[T any](v T) *T {
	return &v
}
