package models

import "time"

type Category struct {
	ID   uint   `json:"id"` // Ubah menjadi uint agar konsisten dengan foreign key di Publication
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// Publication represents the publication entity
type Publication struct {
	ID          uint       `json:"id" gorm:"primaryKey"`
	Title       string     `json:"title" validate:"required,min=3"`        // Wajib diisi dan minimal 3 karakter
	Slug        string     `json:"slug" validate:"required"`               // Wajib diisi
	Description string     `json:"description" validate:"required,min=10"` // Wajib diisi dan minimal 10 karakter
	Img         *string    `json:"img,omitempty"`
	Position    string     `json:"position" gorm:"default:'0'" `            // Wajib diisi
	CategoryID  uint       `json:"category_id" validate:"required,numeric"` // Wajib dan berupa angka
	Caption     *string    `json:"caption,omitempty"`
	UUID        string     `json:"uuid" gorm:"type:uuid;default:gen_random_uuid()"`
	Like        int        `json:"like" gorm:"default:0" validate:"gte=0"` // Tidak boleh negatif
	View        int        `json:"view" gorm:"default:0" validate:"gte=0"` // Tidak boleh negatif
	Status      int        `json:"status" gorm:"default:1" `               // Wajib, hanya boleh 1 atau 2
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `gorm:"index"` // Add this line
	Category    Category   `json:"category" gorm:"foreignKey:CategoryID"`
}

type Tag struct {
	ID            int    `json:"id"`
	PublicationID int    `json:"publicationId"`
	Tag           string `json:"tag"`
}

type Comment struct {
	ID            int    `json:"id"`
	PublicationID int    `json:"publicationId"`
	Name          string `json:"name"`
	Img           string `json:"img"`
	Date          string `json:"date"`
	Description   string `json:"description"`
}
