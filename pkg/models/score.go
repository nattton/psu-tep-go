package models

import "gorm.io/gorm"

type Score struct {
	gorm.Model
	Answer1    float32
	Answer2    float32
	Answer3    float32
	ExamineeID uint
	UserID     uint
}
