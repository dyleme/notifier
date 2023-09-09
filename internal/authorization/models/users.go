package models

type User struct {
	ID           int
	Email        string
	PasswordHash []byte
	TGID         int
}
