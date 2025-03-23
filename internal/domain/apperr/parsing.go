package apperr

import "fmt"

type ParsingError struct {
	Cause error
	Field string
}

func (pe ParsingError) Error() string {
	return fmt.Sprintf("parsing %s: %v", pe.Field, pe.Cause)
}
