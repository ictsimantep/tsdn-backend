// models/footer.go
package models

import (
	"time"

	"gorm.io/gorm"
)

// Footer represents the main footer data
type Footer struct {
	gorm.Model
	ID             int          `json:"id" gorm:"primaryKey"`
	Description    string       `json:"description"`
	DescriptionTwo string       `json:"descriptionTwo"`
	Phone          string       `json:"phone"`
	Mail           string       `json:"mail"`
	Website        string       `json:"website"`
	Address        string       `json:"address"`
	FooterLinks    []FooterLink `json:"footerLinks" gorm:"foreignKey:FooterID"`
	UUID           string       `json:"uuid" gorm:"type:uuid;default:gen_random_uuid()"`
	Active         bool         `json:"active" gorm:"default:true"`
	CreatedAt      time.Time    `json:"created_at" gorm:"default:current_timestamp"`
	UpdatedAt      time.Time    `json:"updated_at" gorm:"default:current_timestamp"`
	CreatedBy      int          `json:"created_by"`
	UpdatedBy      int          `json:"updated_by"`
	DeletedAt      *time.Time   `gorm:"index"` // Add this line
}

// FooterLink represents links for useful links, our company, or social links
type FooterLink struct {
	gorm.Model
	ID       int    `json:"id" gorm:"primaryKey"`
	FooterID uint   `json:"footerID"`
	LinkType string `json:"linkType"`       // "useful", "company", "social"
	Text     string `json:"text,omitempty"` // for usefulLinks and ourCompany
	Icon     string `json:"icon,omitempty"` // for socialLink
	Link     string `json:"link"`           // URL for the link
	UUID     string `json:"uuid" gorm:"type:uuid;default:gen_random_uuid()"`
}
