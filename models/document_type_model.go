package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DocumentType struct {
	ID                 int            `gorm:"primaryKey;autoIncrement" json:"id"`
	UUID               uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4()" json:"uuid"`
	Name               string         `gorm:"type:varchar(255)" json:"name"`
	Prefix             string         `gorm:"type:varchar(255)" json:"prefix"`
	RoleHasRules       []RoleHasRule  `gorm:"foreignKey:Type;references:Prefix" json:"-"`
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

// RoleHasRuleDTO represents a filtered version of RoleHasRule for JSON output
type RoleHasRuleDTO struct {
	RoleGuardName string `json:"role_guard_name"`
	Action        string `json:"action"`
}

// MarshalJSON customizes the JSON output for DocumentType
func (d *DocumentType) MarshalJSON() ([]byte, error) {
	// Create an alias to avoid infinite recursion
	type Alias DocumentType
	alias := Alias(*d)

	// Transform RoleHasRules into the filtered version
	var rules []RoleHasRuleDTO
	for _, rule := range d.RoleHasRules {
		rules = append(rules, RoleHasRuleDTO{
			RoleGuardName: rule.RoleGuardName,
			Action:        rule.Action,
		})
	}

	// Serialize the struct with the transformed rules
	return json.Marshal(&struct {
		RoleHasRules []RoleHasRuleDTO `json:"role_has_rules"`
		*Alias
	}{
		RoleHasRules: rules,
		Alias:        &alias,
	})
}
