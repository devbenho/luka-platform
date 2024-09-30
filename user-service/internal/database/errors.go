package database

import "fmt"

type DBConnectionError struct {
	Operation string
	Err       error
}

func (e *DBConnectionError) Error() string {
	return fmt.Sprintf("database connection error during %s: %v", e.Operation, e.Err)
}

func (e *DBConnectionError) Unwrap() error {
	return e.Err
}

type DBQueryError struct {
	Query string
}

func (e *DBQueryError) Error() string {
	return fmt.Sprintf("database query error: %s", e.Query)
}

type DBTransactionError struct {
	Operation string
	Err       error
}

func (e *DBTransactionError) Error() string {
	return fmt.Sprintf("database transaction error during %s: %v", e.Operation, e.Err)
}

func (e *DBTransactionError) Unwrap() error {
	return e.Err
}

type DBNotFoundError struct {
	Entity string
}

func (e *DBNotFoundError) Error() string {
	return fmt.Sprintf("%s not found", e.Entity)
}

type DBValidationError struct {
	Field string
}

func (e *DBValidationError) Error() string {
	return fmt.Sprintf("validation failed on field '%s'", e.Field)
}

type DBInternalError struct {
	Message string
}

func (e *DBInternalError) Error() string {

	return fmt.Sprintf("internal database error: %s", e.Message)
}

type DBDuplicateError struct {
	Entity string
	Field  string
	Value  string
}

func (e *DBDuplicateError) Error() string {
	return fmt.Sprintf("%s with %s '%s' already exists", e.Entity, e.Field, e.Value)
}
