package models

type Quiz struct {
	ID    uint   `gorm:"primarykey" json:"id"`
	Quiz1 string `json:"quiz1"`
	Quiz2 string `json:"quiz2"`
	Quiz3 string `json:"quiz3"`
}
