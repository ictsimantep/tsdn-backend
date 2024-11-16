package dto

import "time"

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

type UpdateUserRequest struct {
	Fullname string `json:"fullname,omitempty" validate:"omitempty,min=2,max=100"`
	Email    string `json:"email,omitempty" validate:"omitempty,email"`
	Mobile   string `json:"mobile,omitempty" validate:"omitempty,e164"`
	Password string `json:"password,omitempty" validate:"omitempty,min=8"`
}

type UpdateUserResponse struct {
	ID       uint   `json:"id"`
	UUID     string `json:"uuid"`
	Fullname string `json:"fullname"`
	Username string `json:"username"`
	Mobile   string `json:"mobile"`
	Email    string `json:"email"`
	Message  string `json:"message"`
}

// UserDetailResponse is the response struct for user detail
type UserDetailResponse struct {
	ID         uint       `json:"id"`
	UUID       string     `json:"uuid"`
	Fullname   string     `json:"fullname"`
	Username   string     `json:"username"`
	Mobile     string     `json:"mobile"`
	Email      string     `json:"email"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	VerifiedAt *time.Time `json:"verified_at,omitempty"`
}
