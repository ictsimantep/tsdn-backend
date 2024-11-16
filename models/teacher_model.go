package models

import (
	"time"

	"github.com/google/uuid"
)

type Teacher struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"size:255;not null"`
	Photo     string    `json:"photo" gorm:"size:255"`
	Position  string    `json:"position" gorm:"size:255"`
	Location  string    `json:"location" gorm:"size:255"`
	Content   string    `json:"content"`
	Facebook  string    `json:"facebook" gorm:"size:255"`
	Twitter   string    `json:"twitter" gorm:"size:255"`
	Instagram string    `json:"instagram" gorm:"size:255"`
	Linkedin  string    `json:"linkedin" gorm:"size:255"`
	Handphone string    `json:"handphone" gorm:"size:20"`
	Email     string    `json:"email" gorm:"size:255"`
	CreatedAt time.Time `json:"created_at" gorm:"default:current_timestamp"`
	UpdatedAt time.Time `json:"updated_at" gorm:"default:current_timestamp"`
	CreatedBy int       `json:"created_by"`
	UpdatedBy int       `json:"updated_by"`
	UUID      uuid.UUID `json:"uuid" gorm:"type:uuid;default:gen_random_uuid()"`
}
