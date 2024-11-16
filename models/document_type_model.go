package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DocumentType struct {
	ID                 int            `gorm:"primaryKey;autoIncrement" json:"id"`
	UUID               uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4()" json:"uuid"`
	Name               string         `gorm:"type:varchar(255)" json:"name"`
	Prefix             string         `gorm:"type:varchar(255)" json:"prefix"`
	RoleHasRules       []RoleHasRule  `gorm:"foreignKey:Type;references:Prefix"`
	CreatedAt          time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	DocumentCategoryID *int           `gorm:"type:int" json:"document_category_id"`
}

// TableName overrides the default table name for GORM
func (DocumentType) TableName() string {
	return "document_type"
}

// BeforeCreate hook to generate UUID before creating a new record
func (d *DocumentType) BeforeCreate(tx *gorm.DB) (err error) {
	d.UUID = uuid.New()
	return
}
