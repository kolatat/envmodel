package envmodel

import "fmt"

// ParseError occurs during the parsing of a specific struct field. It contains detail about that field, and the reason
// for the error.
type ParseError struct {
	KeyName   string
	FieldName string
	TypeName  string
	Value     string
	Reason    error
}

func (err *ParseError) Error() string {
	if "" == err.Value {
		return fmt.Sprintf("assigning %s to %s: %s",
			err.KeyName, err.FieldName,
			err.Reason)
	}
	return fmt.Sprintf("assigning %s to %s, parsing '%s' as %s: %s",
		err.KeyName, err.FieldName,
		err.Value, err.TypeName,
		err.Reason)
}

func (err *ParseError) Unwrap() error {
	return err.Reason
}
