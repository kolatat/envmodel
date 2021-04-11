package envmodel

import (
	"errors"
	"fmt"
)

var (
	// A field is marked as required, but the environment variable of that key is not defined
	ErrRequiredKeyUndefined = errors.New("required key is not defined")

	// The map entry is not specified in the correct format, expects "key:value"
	ErrInvalidMapEntry = errors.New("invalid map entry")

	ErrInvalidTarget   = errors.New("target must be a struct pointer")
	ErrUnsupportedType = errors.New("unsupported data type")
)

type typedError struct {
	Msg    string
	Reason error
}

func newTypedError(reason error, format string, args ...interface{}) *typedError {
	return &typedError{
		Msg:    fmt.Sprintf(format, args...),
		Reason: reason,
	}
}

func (err *typedError) Error() string {
	return err.Msg
}

func (err *typedError) Unwrap() error {
	return err.Reason
}
