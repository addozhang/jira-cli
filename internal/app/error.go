package app

import "fmt"

const ErrorExitCode = 10

type Error struct {
	Message string
	Next    string
	Cause   error
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

func (e *Error) Unwrap() error { return e.Cause }

func NewError(message, next string) *Error {
	return &Error{Message: message, Next: next}
}

func WrapError(message, next string, cause error) *Error {
	return &Error{Message: message, Next: next, Cause: cause}
}
