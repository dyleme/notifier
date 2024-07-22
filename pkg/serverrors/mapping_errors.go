package serverrors

import "fmt"

type MappingError struct {
	Cause error
	Field string
}

func (me MappingError) Error() string {
	return me.Cause.Error()
}

func (me MappingError) MappingError() string {
	return fmt.Sprintf("%v for field %q", me.Cause.Error(), me.Field)
}

func NewMappingError(cause error, field string) error {
	return MappingError{Cause: cause, Field: field}
}
