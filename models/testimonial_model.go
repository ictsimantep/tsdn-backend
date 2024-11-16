package models

import (
	"time"

	"github.com/google/uuid"
)

type Testimonial struct {
	ID        int       `json:"id" gorm:"primaryKey;autoIncrement"`
	Name      string    `json:"name" gorm:"size:255;not null"`
	ImageURL  string    `json:"image_url" gorm:"size:255"`
	Photo     string    `json:"photo" gorm:"size:255"`
	Content   string    `json:"content"`
	Active    bool      `json:"active" gorm:"default:true"`
	CreatedAt time.Time `json:"created_at" gorm:"default:current_timestamp"`
	UpdatedAt time.Time `json:"updated_at" gorm:"default:current_timestamp"`
	CreatedBy int       `json:"created_by"`
	UpdatedBy int       `json:"updated_by"`
	UUID      uuid.UUID `json:"uuid" gorm:"type:uuid;default:gen_random_uuid()"`
}
