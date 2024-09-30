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
