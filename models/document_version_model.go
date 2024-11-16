package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DocumentVersion struct {
	ID                int        `gorm:"primaryKey;autoIncrement" json:"id"`
	UUID              uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4()" json:"uuid"`
	File              string     `gorm:"type:varchar(255)" json:"file"`
	DocumentControlID *int       `gorm:"type:int" json:"document_control_id"`
	CreatedAt         time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt         *time.Time `gorm:"index" json:"deleted_at"` // soft delete support
	Version           *int       `gorm:"type:int" json:"version"`
	StatusDocumentID  *int       `gorm:"type:int" json:"status_document_id"`
	Note              string     `gorm:"type:text" json:"note"`
}

// TableName overrides the default table name
func (DocumentVersion) TableName() string {
	return "document_version"
}

// BeforeCreate is a GORM hook that sets a UUID before inserting a new record
func (d *DocumentVersion) BeforeCreate(tx *gorm.DB) (err error) {
	d.UUID = uuid.New()
	return
}
