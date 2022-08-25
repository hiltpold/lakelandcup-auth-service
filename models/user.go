package models

import (
	"github.com/google/uuid"
	"github.com/hiltpold/lakelandcup-auth-service/utils"
)

type User struct {
	Id       uuid.UUID `json:"id" gorm:"primaryKey"`
	Email    string    `json:"email"`
	Password string    `json:"password"`
}

// NewUser initializes a new user from an email, password and user data.
func NewUser(email string, password string) (*User, error) {
	id := uuid.New()

	pw, err := utils.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &User{
		Id:       id,
		Email:    email,
		Password: pw,
	}
	return user, nil
}
