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
	v := validator.New()
	// "slug": lowercase letters, digits and hyphens only. Registered here so
	// every DTO validated through this Validator can use it.
	if err := v.RegisterValidation("slug", isSlug); err != nil {
		panic(err)
	}
	return &Validator{
		validate: v,
	}
}

func isSlug(fl validator.FieldLevel) bool {
	for _, char := range fl.Field().String() {
		if !(char >= 'a' && char <= 'z' || char >= '0' && char <= '9' || char == '-') {
			return false
		}
	}
	return true
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
