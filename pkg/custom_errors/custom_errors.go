package customerrors

import "fmt"

type CustomError struct {
	Message string
	Err     error
	Details map[string]any
}

func (e *CustomError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s %v", e.Message, e.Err)
	}
	return e.Message
}

func New(message string, err error) *CustomError {
	return &CustomError{Message: message, Err: err}
}

func NewWithDetails(message string, err error, details map[string]any) *CustomError {
	return &CustomError{Message: message, Err: err, Details: details}
}
