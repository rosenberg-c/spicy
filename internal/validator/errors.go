package validator

import (
	"errors"
	"fmt"
)

type ValidationError struct {
	Input    string
	Response string
	Err      error
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed for %q: %v", e.Input, e.Err)
}

func (e *ValidationError) Unwrap() error {
	return e.Err
}

// Sentinel errors
var (
	ErrInvalidAction = errors.New("invalid action response")
	ErrMissingField  = errors.New("missing required field")
	ErrInvalidJSON   = errors.New("invalid JSON response")
)
