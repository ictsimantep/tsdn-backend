package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CategoryDocument struct {
	ID        int            `gorm:"primaryKey;autoIncrement" json:"id"`
	UUID      uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4()" json:"uuid"`
	Name      string         `gorm:"type:varchar(255)" json:"name"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Prefix    string         `gorm:"type:varchar(255)" json:"prefix"`
}

// TableName overrides the table name used by GORM
func (CategoryDocument) TableName() string {
	return "category_document"
}

// BeforeCreate is a GORM hook that generates a UUID before inserting a new record
func (c *CategoryDocument) BeforeCreate(tx *gorm.DB) (err error) {
	c.UUID = uuid.New()
	return
}
