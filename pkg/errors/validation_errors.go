package errors

import "fmt"

type ValidationError struct {
	Field string      `json:"field"`
	Tag   string      `json:"tag"`
	Value interface{} `json:"value"`
}

type ValidationErrors []ValidationError

func (ve ValidationErrors) Error() string {
	return fmt.Sprintf("%v", []ValidationError(ve))
}

func NewValidationError(field, tag string, value interface{}) ValidationError {
	return ValidationError{
		Field: field,
		Tag:   tag,
		Value: value,
	}
}

func (ve ValidationErrors) Fields() []string {
	fields := make([]string, len(ve))
	for i, e := range ve {
		fields[i] = e.Field
	}
	return fields
}

func (ve ValidationErrors) Values() []interface{} {
	values := make([]interface{}, len(ve))
	for i, e := range ve {
		values[i] = e.Value
	}
	return values
}
