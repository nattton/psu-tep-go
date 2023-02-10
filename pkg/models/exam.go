package models

type Exam struct {
	ID    uint   `gorm:"primarykey" json:"id"`
	Prop1 string `json:"prop1"`
	Prop2 string `json:"prop2"`
	Prop3 string `json:"prop3"`
}
