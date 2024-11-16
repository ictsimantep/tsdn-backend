package models

import (
	"time"

	"github.com/google/uuid"
)

// Form represents the overall form structure
type Form struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	Slug            string     `gorm:"unique" json:"slug"`
	FormTitle       string     `json:"formTitle"`                      // Title of the form
	FormDescription string     `json:"formDescription"`                // Description of the form
	Steps           []FormStep `gorm:"foreignKey:FormID" json:"steps"` // Form steps
	CreatedAt       time.Time  `json:"createdAt"`                      // Automatically created at timestamp
	UUID            uuid.UUID  `json:"uuid" gorm:"type:uuid;default:gen_random_uuid()"`
}

// FormStep represents a step in the form wizard
type FormStep struct {
	ID              uint        `gorm:"primaryKey" json:"id"`
	FormID          uint        `json:"formId"` // References the Form
	StepTitle       string      `json:"stepTitle"`
	StepOrder       uint        `json:"stepOrder"`                          // Title of the step
	StepDescription string      `json:"stepDescription"`                    // Description of the step
	FormItems       []FormField `gorm:"foreignKey:StepID" json:"formItems"` // Fields in the step
	CreatedAt       time.Time   `json:"createdAt"`
	UUID            uuid.UUID   `json:"uuid" gorm:"type:uuid;default:gen_random_uuid()"`
}

// FormField represents each field within a step
type FormField struct {
	ID              uint             `gorm:"primaryKey" json:"id"`
	StepID          uint             `json:"stepId"`      // References the Step
	Label           string           `json:"label"`       // Field label
	Type            string           `json:"type"`        // Field type (e.g., text, email)
	Name            string           `json:"name"`        // Field name
	Placeholder     string           `json:"placeholder"` // Field placeholder
	Required        bool             `json:"required"`    // Is field required?
	FormItemOptions []FormItemOption `gorm:"foreignKey:FormItemID" json:"formItemOptions"`
	CreatedAt       time.Time        `json:"createdAt"`
	UUID            uuid.UUID        `json:"uuid" gorm:"type:uuid;default:gen_random_uuid()"`
}

// FieldOption represents a selectable option for dropdowns
type FormItemOption struct {
	ID          uint      `gorm:"primaryKey"`
	FormItemID  uint      `gorm:"not null" json:"formItemID"`
	OptionValue string    `gorm:"not null"  json:"optionValue"`
	OptionLabel string    `gorm:"not null"  json:"optionLabel"`
	UUID        uuid.UUID `json:"uuid" gorm:"type:uuid;default:gen_random_uuid()"`
}

// FormSubmission represents the submitted form data
type FormSubmission struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	FormID      uint      `json:"formId"`     // References the Form
	FieldName   string    `json:"fieldName"`  // Field name (e.g., firstName)
	FieldValue  string    `json:"fieldValue"` // Submitted value
	Email       string    `json:"email"`
	SubmittedAt time.Time `json:"submittedAt"`
	UUID        uuid.UUID `json:"uuid" gorm:"type:uuid;default:gen_random_uuid()"`
}
