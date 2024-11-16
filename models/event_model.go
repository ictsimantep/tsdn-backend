package models

import (
	"time"
)

type Event struct {
	ID          int       `json:"id" gorm:"primaryKey"`
	Title       string    `json:"title" gorm:"size:255;not null"`
	Category    string    `json:"category" gorm:"size:255"`
	ImageURL    string    `json:"image_url" gorm:"size:255"`
	Content     string    `json:"content"`
	Description string    `json:"description"`
	StartDate   time.Time `json:"start_date"`
	StartTime   string    `json:"start_time"`
	EndDate     time.Time `json:"end_date"`
	Location    string    `json:"location" gorm:"size:255"`
	Slug        string    `json:"slug" gorm:"size:255;unique"`
	CreatedAt   time.Time `json:"created_at" gorm:"default:current_timestamp"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"default:current_timestamp"`
	DeletedAt   time.Time `json:"deleted_at" gorm:"default:current_timestamp"`
	CreatedBy   int       `json:"created_by"`
	UpdatedBy   int       `json:"updated_by"`
	UUID        string    `json:"uuid" gorm:"type:uuid;default:gen_random_uuid()"`
}
