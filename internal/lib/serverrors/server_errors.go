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
