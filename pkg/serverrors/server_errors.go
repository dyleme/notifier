package serverrors

type InternalError interface {
	error
	ServerError()
}

type RepositoryError struct {
	err error
}

func NewRepositoryError(err error) RepositoryError {
	return RepositoryError{err: err}
}

func (r RepositoryError) ServerError() {
}

func (r RepositoryError) Error() string {
	return r.err.Error()
}

type ServiceError struct {
	err error
}

func NewServiceError(err error) ServiceError {
	return ServiceError{err: err}
}

func (r ServiceError) Error() string {
	return r.err.Error()
}

func (r ServiceError) ServerError() {
}
