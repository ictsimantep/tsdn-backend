package models

import (
	"time"

	"github.com/google/uuid"
)

type Setting struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Key       string    `json:"key" gorm:"size:255;not null"`
	Value     string    `json:"value" gorm:"size:255"`
	CreatedAt time.Time `json:"created_at" gorm:"default:current_timestamp"`
	UpdatedAt time.Time `json:"updated_at" gorm:"default:current_timestamp"`
	CreatedBy int       `json:"created_by"`
	UpdatedBy int       `json:"updated_by"`
	UUID      uuid.UUID `json:"uuid" gorm:"type:uuid;default:gen_random_uuid()"`
}
