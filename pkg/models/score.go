package models

import "time"

type Score struct {
	ID         uint      `gorm:"primarykey" json:"-"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	Task1      float32   `json:"task1"`
	Task2      float32   `json:"task2"`
	Task3      float32   `json:"task3"`
	ExamineeID uint      `json:"-"`
	UserID     uint      `json:"-"`
	User       User      `json:"user"`
}
