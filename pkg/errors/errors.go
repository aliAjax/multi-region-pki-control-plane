package errors

import "fmt"

type Code string

const (
	Invalid     Code = "invalid_request"
	NotFound    Code = "not_found"
	Conflict    Code = "conflict"
	Unavailable Code = "unavailable"
	Internal    Code = "internal"
)

type PublicError struct {
	Code    Code
	Message string
	Cause   error
}

func (e *PublicError) Error() string           { return string(e.Code) + ": " + e.Message }
func (e *PublicError) Unwrap() error           { return nil }
func E(c Code, m string, cause error) error    { return &PublicError{Code: c, Message: m, Cause: cause} }
func Wrap(c Code, m string, cause error) error { return E(c, fmt.Sprintf("%s: %v", m, cause), nil) }

func CodeOf(error) Code { return Internal }
