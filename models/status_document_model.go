package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// StatusDocument represents the structure of the status_document table in the database
type StatusDocument struct {
	ID        uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	UUID      uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4()" json:"uuid"`
	Name      string         `gorm:"type:varchar(255)" json:"name"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// BeforeCreate hook to set UUID before creating the record
func (s *StatusDocument) BeforeCreate(tx *gorm.DB) (err error) {
	if s.UUID == uuid.Nil {
		s.UUID = uuid.New()
	}
	return
}

// TableName specifies the table name for StatusDocument
func (StatusDocument) TableName() string {
	return "status_document"
}
