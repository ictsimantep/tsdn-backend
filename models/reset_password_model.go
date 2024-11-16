package models

import "time"

type PasswordResetToken struct {
	ID        uint      `gorm:"primaryKey"`
	Email     string    `gorm:"index"`       // Email associated with the token
	Token     string    `gorm:"uniqueIndex"` // Unique reset token
	ExpiresAt time.Time // Expiration time for the token
	CreatedAt time.Time
}

func (PasswordResetToken) TableName() string {
	return "password_reset_token"
}
