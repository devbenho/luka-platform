package validation

import (
	"github.com/devbenho/luka-platform/pkg/errors"
	"github.com/go-playground/validator/v10"
)

// Validator is a struct that holds the validator instance
type Validator struct {
	validate *validator.Validate
}

// NewValidator creates a new Validator instance
func NewValidator() *Validator {
	return &Validator{
		validate: validator.New(),
	}
}

// ValidateStruct validates a struct based on the tags
func (v *Validator) ValidateStruct(s interface{}) error {
	if err := v.validate.Struct(s); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			var appErrors errors.ValidationErrors
			for _, e := range validationErrors {
				newError := errors.NewValidationError(
					e.Field(),
					e.Tag(),
					e.Value(),
				)
				appErrors = append(appErrors, newError)
			}
			return appErrors
		}
		return errors.NewBadRequestError("validation failed")
	}
	return nil
}

func (v *Validator) ValidateField(field interface{}, tag string) error {
	return v.validate.Var(field, tag)
}
