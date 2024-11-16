package models

import (
	"time"
)

type Gallery struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	Year      string    `json:"year" gorm:"size:10;not null"`
	Category  string    `json:"category" gorm:"size:255;default:'General';not null"`
	ImageURL  string    `json:"image_url" gorm:"unique;not null"`
	CreatedAt time.Time `json:"created_at" gorm:"default:current_timestamp"`
	UpdatedAt time.Time `json:"updated_at" gorm:"default:current_timestamp"`
	CreatedBy int       `json:"created_by"`
	UpdatedBy int       `json:"updated_by"`
	UUID      string    `json:"uuid" gorm:"type:uuid;default:gen_random_uuid()"`
}

func (Gallery) TableName() string {
	return "galleries"
}
