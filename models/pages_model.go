// models/page.go
package models

import (
	"time"
)

// Page represents the structure of a page with camelCase JSON serialization
type Page struct {
	ID        uint       `json:"id"`      // "id" in JSON
	Title     string     `json:"title"`   // "title" in JSON
	Slug      string     `json:"slug"`    // "slug" in JSON
	Image     string     `json:"image"`   // "image" in JSON
	Details   string     `json:"details"` // "details" in JSON
	UUID      string     `json:"uuid" gorm:"type:uuid;default:gen_random_uuid()"`
	CreatedAt time.Time  `json:"created_at" gorm:"default:current_timestamp"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"default:current_timestamp"`
	CreatedBy int        `json:"created_by"`
	UpdatedBy int        `json:"updated_by"`
	DeletedAt *time.Time `gorm:"index"` // Add this line
}
