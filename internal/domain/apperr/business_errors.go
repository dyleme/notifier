package apperr

import (
	"fmt"
)

type BusinessError interface {
	error
	BusinessError() string
}

type NotFoundError struct {
	objectName string
	err        error
}

func NewNotFoundError(err error, objectName string) NotFoundError {
	return NotFoundError{
		objectName: objectName,
		err:        err,
	}
}

func (n NotFoundError) Error() string {
	return fmt.Sprintf("not found %s object", n.objectName)
}

func (n NotFoundError) BusinessError() string {
	return n.err.Error()
}

type NoDeletionsError struct {
	objectName string
}

func NewNoDeletionsError(objectName string) NoDeletionsError {
	return NoDeletionsError{objectName: objectName}
}

func (n NoDeletionsError) Error() string {
	return fmt.Sprintf("no deletions for %s object", n.objectName)
}

func (n NoDeletionsError) BusinessError() string {
	return fmt.Sprintf("no deletions for %s object", n.objectName)
}

type UniqueError struct {
	name  string
	value string
}

func NewUniqueError(name, value string) UniqueError {
	return UniqueError{
		name:  name,
		value: value,
	}
}

func (u UniqueError) Error() string {
	return fmt.Sprintf("%q %s already exists", u.value, u.name)
}

func (u UniqueError) BusinessError() string {
	return fmt.Sprintf("%q %s already exists", u.value, u.name)
}

type InvalidAuthError struct {
	msg string
}

func NewInvalidAuth(msg string) InvalidAuthError {
	return InvalidAuthError{msg: msg}
}

func (i InvalidAuthError) Error() string {
	return i.msg
}

func (i InvalidAuthError) BusinessError() string {
	return i.msg
}

type BusinessLogicError struct {
	msg string
}

func NewBusinessLogicError(msg string) BusinessLogicError {
	return BusinessLogicError{msg: msg}
}

func (b BusinessLogicError) Error() string {
	return b.msg
}

func (b BusinessLogicError) BusinessError() string {
	return b.msg
}
