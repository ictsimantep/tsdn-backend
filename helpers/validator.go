package helpers

import (
	"github.com/go-playground/validator/v10"
)

// Validator instance
var validate = validator.New()

// ValidateStruct validates a struct and returns an error if invalid
func ValidateStruct(s interface{}) error {
	return validate.Struct(s)
}
