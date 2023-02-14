package forms

type Score struct {
	ExamineeID int     `json:"examinee_id"`
	Task1      float32 `json:"task1"`
	Task2      float32 `json:"task2"`
	Task3      float32 `json:"task3"`
}
