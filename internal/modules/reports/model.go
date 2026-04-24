package reports

import (
	"time"

	"github.com/google/uuid"
)

type TargetType string
type Status string

const (
	TargetTypeUser TargetType = "user"
	TargetTypePost TargetType = "post"

	StatusPending  Status = "pending"
	StatusResolved Status = "resolved"
)

type Report struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ReporterID uuid.UUID  `gorm:"type:uuid;not null;index"                       json:"reporter_id"`
	TargetType TargetType `gorm:"type:varchar(20);not null;index"                json:"target_type"`
	TargetID   uuid.UUID  `gorm:"type:uuid;not null;index"                       json:"target_id"`
	Reason     string     `gorm:"type:text;not null"                             json:"reason"`
	Status     Status     `gorm:"type:varchar(20);default:'pending';not null"    json:"status"`
	CreatedAt  time.Time  `json:"created_at"`
}

type CreateReportRequest struct {
	TargetType TargetType `json:"target_type" binding:"required,oneof=user post"`
	TargetID   uuid.UUID  `json:"target_id" binding:"required"`
	Reason     string     `json:"reason" binding:"required"`
}
