package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RoleHasRule represents the role_has_rule table
type RoleHasRule struct {
	ID            uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	UUID          uuid.UUID `json:"uuid" gorm:"type:uuid;default:gen_random_uuid()"`
	RoleGuardName string    `json:"role_guard_name" gorm:"type:varchar(255)"`
	RulePolicy    string    `json:"rule_policy" gorm:"type:varchar(255)"`
	Action        string    `json:"action" gorm:"type:varchar(255)"`
	Category      string    `json:"category"`
	Type          string    `json:"type"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// BeforeCreate is a GORM hook that sets a UUID before creating a RoleHasRule
func (r *RoleHasRule) BeforeCreate(tx *gorm.DB) (err error) {
	if r.UUID == uuid.Nil {
		r.UUID = uuid.New()
	}
	return
}
