package models

import (
	"time"
)

type TrafficIP struct {
	ID        int       `gorm:"primaryKey;autoIncrement"`
	UserID    int       `gorm:"type:integer"`
	IP        string    `gorm:"type:varchar;uniqueIndex"` // Ensures IP is unique
	Country   string    `gorm:"type:varchar"`
	City      string    `gorm:"type:varchar"`
	Latitude  string    `gorm:"type:varchar"` // Assuming these are strings based on your schema
	Longitude string    `gorm:"type:varchar"`
	Url       string    `gorm:"type:varchar"`
	Os        string    `gorm:"type:varchar"` // New field for OS
	Browser   string    `gorm:"type:varchar"` // New field for Browser
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}
