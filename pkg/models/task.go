package models

type Task struct {
	ID    uint   `gorm:"primarykey" json:"id"`
	Task0 string `json:"task0"`
	Task1 string `json:"task1"`
	Task2 string `json:"task2"`
	Task3 string `json:"task3"`
}
