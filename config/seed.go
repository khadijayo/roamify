package config

import (
	"log"

	"github.com/khadijayo/roamify/internal/modules/users"
)

func SeedAdmins() {
	if App == nil || len(App.AdminEmails) == 0 {
		return
	}

	result := DB.Model(&users.User{}).
		Where("LOWER(email) IN ?", App.AdminEmails).
		Update("role", users.RoleAdmin)
	if result.Error != nil {
		log.Printf("[seed] failed to seed admin users: %v", result.Error)
		return
	}
	log.Printf("[seed] admin role ensured for %d configured user(s)", result.RowsAffected)
}
