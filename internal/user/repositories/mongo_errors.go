package repositories

import (
	"strings"

	"github.com/devbenho/luka-platform/pkg/database"
	"go.mongodb.org/mongo-driver/mongo"
)

func mapMongoError(err error, entity string) error {
	if err == nil {
		return nil
	}

	switch {
	case mongo.IsDuplicateKeyError(err):
		if strings.Contains(err.Error(), "email") {
			return &database.DBDuplicateError{
				Entity: entity,
				Field:  "email",
			}
		}
		return &database.DBDuplicateError{
			Entity: entity,
			Field:  "username",
		}
	case mongo.IsTimeout(err):
		return &database.DBConnectionError{
			Operation: "timeout",
			Err:       err,
		}
	case mongo.IsNetworkError(err):
		return &database.DBConnectionError{
			Operation: "network error",
			Err:       err,
		}
	// handle validation errors
	case strings.Contains(err.Error(), "validation failed"):
		return &database.DBValidationError{
			Field: "email",
		}
	case err == mongo.ErrNoDocuments:
		return &database.DBNotFoundError{
			Entity: entity,
		}
	default:
		return &database.DBInternalError{
			Message: err.Error(),
		}
	}
}
