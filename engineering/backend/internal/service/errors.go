package service

import (
	"errors"
	"fmt"
)

// ErrValidation marks a request as rejected by a business rule (a required
// field was empty, a referenced parent id doesn't exist, ...) rather than
// failing because of the database. Handlers map it to 400 Bad Request
// instead of 500.
var ErrValidation = errors.New("validation failed")

// ValidationErrorf builds an error that errors.Is(err, ErrValidation) finds.
func ValidationErrorf(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrValidation, fmt.Sprintf(format, args...))
}
