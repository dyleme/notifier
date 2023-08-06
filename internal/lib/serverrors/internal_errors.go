package serverrors

import "fmt"

type InternalError interface {
	error
	InternalError() string
}

type RepositoryError struct {
	err error
}

func NewRepositoryError(err error) RepositoryError {
	return RepositoryError{
		err: err,
	}
}

func (re RepositoryError) Error() string {
	return fmt.Sprintf("repository %s", re.err)
}

func (re RepositoryError) InternalError() string {
	return re.err.Error()
}
