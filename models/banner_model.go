package models

import (
	"time"
)

type Banner struct {
	ID        uint       `json:"id" gorm:"primaryKey"`
	Title     string     `json:"title" gorm:"size:255;not null"`
	Subtitle  string     `json:"subtitle" gorm:"size:255"`
	Link      string     `json:"link" gorm:"size:255"`
	ImageURL  *string    `json:"image_url,omitempty"`
	UUID      string     `json:"uuid" gorm:"type:uuid;default:gen_random_uuid()"`
	Active    bool       `json:"active" gorm:"default:true"`
	CreatedAt time.Time  `json:"created_at" gorm:"default:current_timestamp"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"default:current_timestamp"`
	CreatedBy int        `json:"created_by"`
	UpdatedBy int        `json:"updated_by"`
	DeletedAt *time.Time `gorm:"index"` // Add this line
}
