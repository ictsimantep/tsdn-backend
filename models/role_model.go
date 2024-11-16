package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Role struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	UUID      uuid.UUID `json:"uuid" gorm:"type:uuid;default:gen_random_uuid()"`
	Name      string    `json:"name" gorm:"type:varchar(255)"`
	GuardName string    `json:"guard_name" gorm:"type:varchar(255)"`
}

// BeforeCreate is a GORM hook that sets a UUID before creating a role.
func (role *Role) BeforeCreate(tx *gorm.DB) (err error) {
	if role.UUID == uuid.Nil {
		role.UUID = uuid.New()
	}
	return
}
