package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DocumentControl represents the document control model
type DocumentControl struct {
	ID                 int        `gorm:"primaryKey;autoIncrement" json:"id"`
	UUID               uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4()" json:"uuid"`
	DocumentName       string     `gorm:"type:varchar(255)" json:"document_name"`
	Description        string     `gorm:"type:text" json:"description"`
	DocumentNumber     string     `gorm:"type:varchar(255)" json:"document_number"`
	ClauseNumber       string     `gorm:"type:varchar(255)" json:"clause_number"`
	RevisionNumber     int        `gorm:"type:int" json:"revision_number"`
	PublishDate        time.Time  `gorm:"type:date" json:"publish_date"`
	PageCount          int        `gorm:"type:int" json:"page_count"`
	CreatedAt          time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt          *time.Time `gorm:"index" json:"deleted_at"` // for soft delete support
	DocumentTypeID     *int       `gorm:"type:int" json:"document_type_id"`
	DocumentCategoryID *int       `gorm:"type:int" json:"document_category_id"`
	SequenceNumber     *int       `gorm:"type:int" json:"sequence_number"`
	StatusDocumentID   *int       `gorm:"type:int" json:"status_document_id"`
	CreatedBy          *int       `gorm:"type:int" json:"created_by"`
}

// TableName overrides the default table name
func (DocumentControl) TableName() string {
	return "document_control"
}

// BeforeCreate is a GORM hook that sets a UUID before inserting a new record
func (d *DocumentControl) BeforeCreate(tx *gorm.DB) (err error) {
	d.UUID = uuid.New()
	return
}

// DocumentControl represents the document control model
type DocumentControlJoined struct {
	ID                 int        `gorm:"primaryKey;autoIncrement" json:"id"`
	UUID               uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4()" json:"uuid"`
	DocumentName       string     `gorm:"type:varchar(255)" json:"document_name"`
	Description        string     `gorm:"type:text" json:"description"`
	DocumentNumber     string     `gorm:"type:varchar(255)" json:"document_number"`
	ClauseNumber       string     `gorm:"type:varchar(255)" json:"clause_number"`
	RevisionNumber     int        `gorm:"type:int" json:"revision_number"`
	PublishDate        time.Time  `gorm:"type:date" json:"publish_date"`
	PageCount          int        `gorm:"type:int" json:"page_count"`
	CreatedAt          time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt          *time.Time `gorm:"index" json:"deleted_at"` // for soft delete support
	DocumentTypeID     *int       `gorm:"type:int" json:"document_type_id"`
	DocumentCategoryID *int       `gorm:"type:int" json:"document_category_id"`
	SequenceNumber     *int       `gorm:"type:int" json:"sequence_number"`
	StatusDocumentID   *int       `gorm:"type:int" json:"status_document_id"`
	CreatedBy          *int       `gorm:"type:int" json:"created_by"`
	// Fields for joined data
	CategoryName   string `json:"category_name"`
	CategoryPrefix string `json:"category_prefix"`
	TypeName       string `json:"type_name"`
	TypePrefix     string `json:"type_prefix"`
	StatusName     string `json:"status_name"`
}

// TableName overrides the default table name
func (DocumentControlJoined) TableName() string {
	return "document_control"
}
