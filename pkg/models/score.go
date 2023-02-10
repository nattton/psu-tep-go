package models

import "time"

type Score struct {
	ID         uint      `gorm:"primarykey" json:"-"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	Answer1    float32   `json:"answer1"`
	Answer2    float32   `json:"answer2"`
	Answer3    float32   `json:"answer3"`
	ExamineeID uint      `json:"-"`
	UserID     uint      `json:"-"`
	User       User      `json:"user"`
}
