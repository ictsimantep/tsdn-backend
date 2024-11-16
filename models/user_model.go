package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username  string `gorm:"unique"`
	Password  string
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"type:timestamptz" sql:"index"`
}

type UserRegister struct {
	ID         uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	UUID       uuid.UUID      `json:"uuid" gorm:"type:uuid;default:gen_random_uuid()"`
	Username   string         `json:"username" gorm:"type:text"`
	Password   string         `json:"password" gorm:"type:text"`
	Fullname   string         `json:"fullname" gorm:"type:varchar(255)"`
	Mobile     string         `json:"mobile" gorm:"type:varchar(255)"`
	Email      string         `json:"email" gorm:"type:varchar(255)"`
	CreatedAt  time.Time      `json:"created_at" gorm:"type:timestamptz"`
	UpdatedAt  time.Time      `json:"updated_at" gorm:"type:timestamptz"`
	DeletedAt  gorm.DeletedAt `json:"deleted_at" gorm:"type:timestamptz" sql:"index"`
	CreatedBy  *uint          `json:"created_by" gorm:"type:integer"`
	UpdatedBy  *uint          `json:"updated_by" gorm:"type:integer"`
	VerifiedAt *time.Time     `json:"verified_at" gorm:"type:timestamp"`
}

func (UserRegister) TableName() string {
	return "users"
}

// BeforeCreate is a GORM hook that sets a UUID before creating a user.
func (user *UserRegister) BeforeCreate(tx *gorm.DB) (err error) {
	if user.UUID == uuid.Nil {
		user.UUID = uuid.New()
	}
	return
}

type UserGetData struct {
	ID        uint   `gorm:"primaryKey"`
	Username  string `json:"username"`
	Fullname  string `json:"fullname"`
	Email     string `json:"email"`
	Mobile    string `json:"mobile"`
	Password  string `json:"-"` // Exclude password from JSON response
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"type:timestamptz" sql:"index"`
}

func (UserGetData) TableName() string {
	return "users"
}

type UserByAdmin struct {
	ID         uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	UUID       uuid.UUID      `json:"uuid" gorm:"type:uuid;default:gen_random_uuid()"`
	Username   string         `json:"username" gorm:"type:text"`
	Fullname   string         `json:"fullname" gorm:"type:varchar(255)"`
	Mobile     string         `json:"mobile" gorm:"type:varchar(255)"`
	Email      string         `json:"email" gorm:"type:varchar(255)"`
	CreatedAt  time.Time      `json:"created_at" gorm:"type:timestamptz"`
	UpdatedAt  time.Time      `json:"updated_at" gorm:"type:timestamptz"`
	DeletedAt  gorm.DeletedAt `json:"deleted_at" gorm:"type:timestamptz" sql:"index"`
	CreatedBy  *uint          `json:"created_by" gorm:"type:integer"`
	UpdatedBy  *uint          `json:"updated_by" gorm:"type:integer"`
	VerifiedAt *time.Time     `json:"verified_at" gorm:"type:timestamp"`
}

func (UserByAdmin) TableName() string {
	return "users"
}
