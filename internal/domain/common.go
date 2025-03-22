package domain

import "fmt"

type TaskType string

type NotBelongToUserError struct {
	objType          string
	objID            int
	objUserID        int
	interactedUserID int
}

func (e NotBelongToUserError) Error() string {
	return fmt.Sprintf("%s %d[userID=%d] doesn't belong to user %d", e.objType, e.objID, e.objUserID, e.interactedUserID)
}

func NewNotBelongToUserError(objType string, objID, objUserID, interactedUserID int) NotBelongToUserError {
	return NotBelongToUserError{
		objType:          objType,
		objID:            objID,
		objUserID:        objUserID,
		interactedUserID: interactedUserID,
	}
}
