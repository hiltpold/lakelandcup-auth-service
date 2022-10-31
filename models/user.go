package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	Id        uuid.UUID `json:"id" gorm:"primaryKey"`
	FirstName string    `json:"firstName" gorm:"type:varchar(255);not null"`
	LastName  string    `json:"lastName" gorm:"type:varchar(255);not null"`
	Email     string    `json:"email" gorm:"type:varchar(255);unique;not null"`
	Role      string    `json:"role" gorm:"type:varchar(255)"`
	Confirmed bool      `json:"confirmed" gorm:"type:bool;default:false"`
	Password  string    `json:"password"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (user *User) BeforeCreate(db *gorm.DB) error {
	user.Id = uuid.New()
	user.CreatedAt = time.Now().Local()
	return nil
}

func (user *User) BeforeUpdate(db *gorm.DB) error {
	user.UpdatedAt = time.Now().Local()
	return nil
}
