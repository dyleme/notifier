package serverrors

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

type InvalidAuth struct {
	msg string
}

func NewInvalidAuth(msg string) InvalidAuth {
	return InvalidAuth{msg: msg}
}

func (i InvalidAuth) Error() string {
	return i.msg
}

func (i InvalidAuth) BusinessError() string {
	return i.msg
}
