package domain

import "github.com/Dyleme/Notifier/internal/domain/apperr"

type Tag struct {
	ID     int
	UserID int
	Name   string
}

func (t Tag) BelongsTo(userID int) error {
	if t.UserID == userID {
		return nil
	}

	return apperr.NewNotBelongToUserError("tag", t.ID, t.UserID, userID)
}
