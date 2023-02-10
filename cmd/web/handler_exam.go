package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.com/code-mobi/psu-tep/pkg/forms"
	"gitlab.com/code-mobi/psu-tep/pkg/models"
	"gorm.io/gorm"
)

func (h *Handler) getExamHandler(c *gin.Context) {
	var exam models.Exam
	h.db.First(&exam)
	currentPath := getCurrentPath(c)
	c.JSON(http.StatusOK, gin.H{
		"prop1": currentPath + exam.Prop1,
		"prop2": currentPath + exam.Prop2,
		"prop3": currentPath + exam.Prop3,
	})
}

func (h *Handler) rateExamineeHandler(c *gin.Context) {
	var form forms.Score
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	examineeID := c.Param("examinee_id")
	userClaim, _ := h.decodeToken(c)

	var examinee models.Examinee
	var user models.User

	if err := h.db.Find(&examinee, examineeID).Error; err != nil {
		c.AbortWithStatus(http.StatusNotModified)
		return
	}

	if err := h.db.Find(&user, userClaim.ID).Error; err != nil {
		c.AbortWithStatus(http.StatusNotModified)
		return
	}

	var score models.Score
	result := h.db.Where("examinee_id = ? AND user_id = ?", examinee.ID, user.ID).First(&score)
	score.ExamineeID = examinee.ID
	score.UserID = user.ID
	score.Answer1 = form.Answer1
	score.Answer2 = form.Answer2
	score.Answer3 = form.Answer3
	if result.Error == gorm.ErrRecordNotFound {
		h.db.Create(&score)
		c.JSON(http.StatusCreated, gin.H{
			"message": "create score",
		})
		return
	} else if result.Error == nil {
		h.db.Save(&score)
		c.JSON(http.StatusOK, gin.H{
			"message": "update score",
		})
		return
	}

	c.AbortWithStatus(http.StatusNotModified)
}
