package k8s

import "fmt"

// ErrorReason used to check error types
type ErrorReason string

// Error Reasons const used to check error types
const (
	InputError    ErrorReason = "InputError"
	AlreadyExists ErrorReason = "AlreadyExists"
)

type k8sError struct {
	reason ErrorReason
	msg    string
}

func (e k8sError) Error() string { return e.msg }

func checkError(err error, reason ErrorReason) bool {
	if ke, ok := err.(*k8sError); ok && ke.reason == reason {
		return true
	}
	return false
}

func newError(reason ErrorReason, a ...interface{}) error {
	return &k8sError{
		reason: reason,
		msg:    fmt.Sprint(a...),
	}
}

func newErrorf(reason ErrorReason, format string, a ...interface{}) error {
	return &k8sError{
		reason: reason,
		msg:    fmt.Sprintf(format, a...),
	}
}

// NewInputError creates a new "input" error
func NewInputError(a ...interface{}) error {
	return newError(InputError, a...)
}

// NewInputErrorf formats a message according to a format specifier and returns the error
func NewInputErrorf(format string, a ...interface{}) error {
	return newErrorf(InputError, format, a...)
}

// NewAlreadyExistsError creates a new "already exists" error
func NewAlreadyExistsError(a ...interface{}) error {
	return newError(AlreadyExists, a...)
}

// NewAlreadyExistsErrorf formats a message according to a format specifier and returns the error
func NewAlreadyExistsErrorf(format string, a ...interface{}) error {
	return newErrorf(AlreadyExists, format, a...)
}

// IsInputError checks if the error is of type "input"
func IsInputError(err error) bool {
	return checkError(err, InputError)
}

// IsAlreadyExistsError checks if the error is of type "already exists"
func IsAlreadyExistsError(err error) bool {
	return checkError(err, AlreadyExists)
}
