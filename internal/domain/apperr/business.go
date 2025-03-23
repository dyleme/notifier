package apperr

import (
	"fmt"
)

type NotFoundError struct {
	Object string
}

func (n NotFoundError) Error() string {
	return fmt.Sprintf("not found %s object", n.Object)
}

type UniqueError struct {
	Name  string
	Value string
}

func (u UniqueError) Error() string {
	return fmt.Sprintf("value[%q] of [%s] already exists", u.Value, u.Name)
}

type InvalidOffsetError struct {
	Offset int
}

func (i InvalidOffsetError) Error() string {
	return fmt.Sprintf("invalid offset: %d (offset should be between -24 and 24)", i.Offset)
}

type UnexpectedStateError struct {
	Object string
	Reason string
}

func (e UnexpectedStateError) Error() string {
	return e.Object + "is in invalid state: " + e.Reason
}

type NotBelongToUserError struct {
	ObjType          string
	ObjID            int
	objUserID        int
	InteractedUserID int
}

func (e NotBelongToUserError) Error() string {
	return fmt.Sprintf("%s %d[userID=%d] doesn't belong to user %d", e.ObjType, e.ObjID, e.objUserID, e.InteractedUserID)
}

func NewNotBelongToUserError(objType string, objID, objUserID, interactedUserID int) NotBelongToUserError {
	return NotBelongToUserError{
		ObjType:          objType,
		ObjID:            objID,
		objUserID:        objUserID,
		InteractedUserID: interactedUserID,
	}
}
