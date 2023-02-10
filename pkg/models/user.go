package models

import (
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	Name        string    `json:"name"`
	Password    []byte    `json:"-"`
	NewPassword string    `gorm:"-:all" json:"-"`
	Role        string    `json:"role"` // admin, rater
	CreatedAt   time.Time `json:"-"`
	UpdatedAt   time.Time `json:"-"`
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.NewPassword == "" {
		return
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = hashedPassword
	return
}

func (u *User) BeforeSave(tx *gorm.DB) (err error) {
	return u.BeforeCreate(tx)
}

func (u *User) VerifyUser(password string) error {
	err := bcrypt.CompareHashAndPassword(u.Password, []byte(password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return fmt.Errorf("invalid user credentials")
	} else if err != nil {
		return err
	}

	return nil
}
