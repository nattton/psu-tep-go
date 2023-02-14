package models

type Examinee struct {
	ID        uint    `gorm:"primarykey" json:"id"`
	Code      string  `gorm:"unique" json:"code"`
	Firstname string  `json:"firstname"`
	Lastname  string  `json:"lastname"`
	Answer0   string  `json:"answer0"`
	Answer1   string  `json:"answer1"`
	Answer2   string  `json:"answer2"`
	Answer3   string  `json:"answer3"`
	Finish    bool    `json:"finish"`
	Scores    []Score `json:"scores"`
}
