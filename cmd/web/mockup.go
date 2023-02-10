package main

import "gitlab.com/code-mobi/psu-tep/pkg/models"

func (h *Handler) initUser() {
	var user models.User
	err := h.db.First(&user, 1).Error
	if err != nil {
		user.ID = 1
		user.Name = "admin"
		user.NewPassword = "12345678"
		user.Role = "admin"
		h.db.Create(&user)

		var rater1 models.User
		rater1.ID = 2
		rater1.Name = "rater1"
		rater1.NewPassword = "rater@1"
		rater1.Role = "rater"
		h.db.Create(&rater1)

		var rater2 models.User
		rater2.ID = 3
		rater2.Name = "rater2"
		rater2.NewPassword = "rater@2"
		rater2.Role = "rater"
		h.db.Create(&rater2)

		var exam models.Exam
		exam.Prop1 = "/" + examDir + "/RestaurantConversation.mp4"
		exam.Prop2 = "/" + examDir + "/Clothes.mp4"
		exam.Prop3 = "/" + examDir + "/DailyRoutines.mp4"
		h.db.Create(&exam)

		h.mockUpExaminee()
	}
}

func (h *Handler) mockUpExaminee() {
	var examinee models.Examinee
	examinee.Code = "11111"
	examinee.Firstname = "ทดสอบ"
	examinee.Lastname = "ระบบ"
	h.db.Create(&examinee)
}
