package models

import (
	"time"

	"github.com/google/uuid"
)

type Person struct {
	ID           int       `json:"id" gorm:"primaryKey"`
	Name         string    `json:"name" gorm:"size:255;not null"`
	Organization string    `json:"organization" gorm:"size:255"`
	Quote        string    `json:"quote"`
	ImageURL     string    `json:"image_url" gorm:"size:255"`
	Active       bool      `json:"active" gorm:"default:true"`
	CreatedAt    time.Time `json:"created_at" gorm:"default:current_timestamp"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"default:current_timestamp"`
	CreatedBy    int       `json:"created_by"`
	UpdatedBy    int       `json:"updated_by"`
	UUID         uuid.UUID `json:"uuid" gorm:"type:uuid;default:gen_random_uuid()"`
}

func (Person) TableName() string {
	return "person"
}
