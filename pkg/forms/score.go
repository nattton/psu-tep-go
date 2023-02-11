package forms

type Score struct {
	ExamineeID int     `json:"examinee_id"`
	Answer1    float32 `json:"answer1"`
	Answer2    float32 `json:"answer2"`
	Answer3    float32 `json:"answer3"`
}
