package database

import (
	"context"
	"errors"
	"fmt"
	"strings"

	customerrors "github.com/jailtonjunior94/financial/pkg/custom_errors"
	"github.com/go-sql-driver/mysql"
	"github.com/lib/pq"
)

// ConvertDBError converts known database errors to domain errors
// Returns the original error if it cannot be categorized
func ConvertDBError(err error, tableName string) error {
	if err == nil {
		return nil
	}

	// PostgreSQL/CockroachDB errors
	if pqErr, ok := err.(*pq.Error); ok {
		return convertPostgresError(pqErr, tableName)
	}

	// MySQL errors
	if mysqlErr, ok := err.(*mysql.MySQLError); ok {
		return convertMySQLError(mysqlErr, tableName)
	}

	// Context errors
	if errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("database timeout: %w", err)
	}

	if errors.Is(err, context.Canceled) {
		return fmt.Errorf("database operation canceled: %w", err)
	}

	// Return original error if not categorized
	return err
}

// convertPostgresError converts PostgreSQL/CockroachDB specific errors to domain errors
func convertPostgresError(pqErr *pq.Error, tableName string) error {
	switch pqErr.Code {
	case "23505": // unique_violation
		return convertUniqueViolation(pqErr, tableName)
	case "23503": // foreign_key_violation
		return convertFKViolation(pqErr, tableName)
	case "23502": // not_null_violation
		return fmt.Errorf("required field '%s' is missing: %w", pqErr.Column, pqErr)
	case "23514": // check_violation
		return fmt.Errorf("check constraint violation: %w", pqErr)
	case "40001": // serialization_failure
		return fmt.Errorf("transaction conflict, please retry: %w", pqErr)
	case "40P01": // deadlock_detected
		return fmt.Errorf("deadlock detected: %w", pqErr)
	default:
		return fmt.Errorf("database error (%s): %w", pqErr.Code, pqErr)
	}
}

// convertUniqueViolation converts unique constraint violations to domain errors
func convertUniqueViolation(pqErr *pq.Error, tableName string) error {
	constraintName := strings.ToLower(pqErr.Constraint)

	// Detect based on constraint name
	switch {
	case strings.Contains(constraintName, "email"):
		return customerrors.ErrEmailAlreadyExists
	case strings.Contains(constraintName, "users_email"):
		return customerrors.ErrEmailAlreadyExists
	case strings.Contains(constraintName, "name") && tableName == "categories":
		return fmt.Errorf("category name already exists: %w", pqErr)
	default:
		return fmt.Errorf("duplicate entry on %s: %w", constraintName, pqErr)
	}
}

// convertFKViolation converts foreign key violations to domain errors
func convertFKViolation(pqErr *pq.Error, tableName string) error {
	constraintName := strings.ToLower(pqErr.Constraint)

	switch {
	case strings.Contains(constraintName, "parent_id"):
		return customerrors.ErrInvalidParentCategory
	case strings.Contains(constraintName, "user_id"):
		return customerrors.ErrUserNotFound
	case strings.Contains(constraintName, "category_id"):
		return customerrors.ErrCategoryNotFound
	case strings.Contains(constraintName, "budget_id"):
		return customerrors.ErrBudgetNotFound
	default:
		return fmt.Errorf("foreign key violation on %s: %w", constraintName, pqErr)
	}
}

// convertMySQLError converts MySQL specific errors to domain errors
func convertMySQLError(mysqlErr *mysql.MySQLError, tableName string) error {
	switch mysqlErr.Number {
	case 1062: // ER_DUP_ENTRY
		return convertMySQLDuplicateEntry(mysqlErr, tableName)
	case 1452: // ER_NO_REFERENCED_ROW_2
		return convertMySQLFKViolation(mysqlErr, tableName)
	case 1451: // ER_ROW_IS_REFERENCED_2
		return fmt.Errorf("cannot delete: record is referenced: %w", mysqlErr)
	case 1205: // ER_LOCK_WAIT_TIMEOUT
		return fmt.Errorf("lock wait timeout: %w", mysqlErr)
	case 1213: // ER_LOCK_DEADLOCK
		return fmt.Errorf("deadlock detected: %w", mysqlErr)
	default:
		return fmt.Errorf("mysql error (%d): %w", mysqlErr.Number, mysqlErr)
	}
}

// convertMySQLDuplicateEntry converts MySQL duplicate entry errors to domain errors
func convertMySQLDuplicateEntry(mysqlErr *mysql.MySQLError, tableName string) error {
	msg := strings.ToLower(mysqlErr.Message)

	if strings.Contains(msg, "email") {
		return customerrors.ErrEmailAlreadyExists
	}

	return fmt.Errorf("duplicate entry: %w", mysqlErr)
}

// convertMySQLFKViolation converts MySQL foreign key violations to domain errors
func convertMySQLFKViolation(mysqlErr *mysql.MySQLError, tableName string) error {
	msg := strings.ToLower(mysqlErr.Message)

	switch {
	case strings.Contains(msg, "parent_id"):
		return customerrors.ErrInvalidParentCategory
	case strings.Contains(msg, "user_id"):
		return customerrors.ErrUserNotFound
	case strings.Contains(msg, "category_id"):
		return customerrors.ErrCategoryNotFound
	case strings.Contains(msg, "budget_id"):
		return customerrors.ErrBudgetNotFound
	default:
		return fmt.Errorf("foreign key violation: %w", mysqlErr)
	}
}
