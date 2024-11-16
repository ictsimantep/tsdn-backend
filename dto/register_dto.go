// dto/register_dto.go
package dto

import "github.com/google/uuid"

type RegisterRequest struct {
	Fullname        string `json:"fullname" validate:"required"`
	Username        string `json:"username" validate:"required"`
	Mobile          string `json:"mobile" validate:"required"`
	Email           string `json:"email" validate:"required"`
	Password        string `json:"password" validate:"required,min=8"`
	PasswordConfirm string `json:"passwordConfirm" validate:"required,eqfield=Password"`
}

type RegisterResponse struct {
	ID       uint      `json:"id"`
	UUID     uuid.UUID `json:"uuid"`
	Fullname string    `json:"fullname"`
	Username string    `json:"username"`
	Mobile   string    `json:"mobile"`
	Email    string    `json:"email"`
	Message  string    `json:"message"`
}
type RegisterRequestAdmin struct {
	Fullname      string `json:"fullname" validate:"required"`
	Username      string `json:"username" validate:"required"`
	Password      string `json:"password" validate:"required"`
	Mobile        string `json:"mobile" validate:"required"`
	Email         string `json:"email" validate:"required,email"`
	RoleGuardName string `json:"role_guard_name" validate:"required"` // Tambahkan ini untuk peran
}
