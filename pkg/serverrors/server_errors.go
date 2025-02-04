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

type InvalidBusinessStateError struct {
	object string
	reason string
}

func (e InvalidBusinessStateError) Error() string {
	return e.object + "is in invalid state: " + e.reason
}

func NewInvalidBusinessStateError(object, reason string) InvalidBusinessStateError {
	return InvalidBusinessStateError{
		object: object,
		reason: reason,
	}
}
