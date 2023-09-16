package tgwf

import (
	"fmt"

	"github.com/go-telegram/bot/models"
)

func GetMessage(u *models.Update) (*models.Message, error) {
	if u.Message == nil {
		return nil, fmt.Errorf("it is not message")
	}

	return u.Message, nil
}
