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

type ServiceError struct {
	err error
}

func NewServiceError(err error) ServiceError {
	return ServiceError{
		err: err,
	}
}

func (se ServiceError) Error() string {
	return fmt.Sprintf("service %s", se.err)
}

func (se ServiceError) InternalError() string {
	return se.err.Error()
}
