package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Health struct {
	ID                     int            `gorm:"primaryKey;autoIncrement" json:"id"`
	UUID                   uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4()" json:"uuid"`
	CreatedAt              time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt              time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt              gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	UserId                 string         `gorm:"type:varchar(255)" json:"user_id"`
	Nama                   string         `gorm:"type:varchar(255)" json:"nama"`
	JenisKelamin           string         `gorm:"type:varchar(255)" json:"jenis_kelamin"`
	Umur                   string         `gorm:"type:varchar(255)" json:"umur"`
	Bb                     string         `gorm:"type:varchar(255)" json:"bb"`
	Tb                     string         `gorm:"type:varchar(255)" json:"tb"`
	Systol                 string         `gorm:"type:varchar(255)" json:"systol"`
	Diastol                string         `gorm:"type:varchar(255)" json:"diastol"`
	HeartRate              string         `gorm:"type:varchar(255)" json:"heart_rate"`
	Profesi                string         `gorm:"type:varchar(255)" json:"profesi"`
	Risk                   string         `gorm:"type:varchar(255)" json:"risk"`
	Bmi                    string         `gorm:"type:varchar(255)" json:"bmi"`
	RecommendationFood     string         `gorm:"type:varchar(255)" json:"recommendation_food"`
	RecommendationSport    string         `gorm:"type:varchar(255)" json:"recommendation_sport"`
	RecommendationMedicine string         `gorm:"type:varchar(255)" json:"recommendation_medicine"`
	CreatedBy              string         `gorm:"type:varchar(255)" json:"created_by"`
	UpdatedBy              string         `gorm:"type:varchar(255)" json:"updated_by"`
	DeletedBy              string         `gorm:"type:varchar(255)" json:"deleted_by"`
}

// TableName overrides the default table name for GORM
func (Health) TableName() string {
	return "health_data"
}

// BeforeCreate hook to generate UUID before creating a new record
func (d *Health) BeforeCreate(tx *gorm.DB) (err error) {
	d.UUID = uuid.New()
	return
}
