package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"gitlab.com/code-mobi/psu-tep/pkg/forms"
	"gitlab.com/code-mobi/psu-tep/pkg/models"
	"gorm.io/gorm"
)

func (h *Handler) getQuizHandler(c *gin.Context) {
	var quiz models.Quiz
	h.db.First(&quiz)
	currentPath := getCurrentPath(c)
	c.JSON(http.StatusOK, gin.H{
		"quiz1": currentPath + quiz.Quiz1,
		"quiz2": currentPath + quiz.Quiz2,
		"quiz3": currentPath + quiz.Quiz3,
	})
}

func (h *Handler) saveQuizHandler(c *gin.Context) {
	claims, _ := h.decodeToken(c)
	if claims.Role != "admin" {
		c.JSON((http.StatusUnauthorized), gin.H{
			"error": claims.Role + " unauthorized this function.",
		})
		return
	}

	println("ID", claims.ID)
	println("Role", claims.Role)
	var quiz models.Quiz
	h.db.First(&quiz)

	quizPath := fmt.Sprintf("/%s/", quizDir)
	if err := os.MkdirAll(h.storePath+quizPath, os.ModePerm); err != nil {
		log.Fatal(err)
	}

	for i := 1; i < 4; i++ {
		seq := fmt.Sprint(i)
		file, err := c.FormFile("quiz" + seq)
		if err == nil {
			filename := fmt.Sprintf("%s_%d_%s", seq, time.Now().Unix(), file.Filename)
			if err := c.SaveUploadedFile(file, h.storePath+quizPath+filename); err != nil {
				log.Fatal(err)
			}

			h.db.Model(&quiz).Update("quiz"+seq, quizPath+filename)

			c.JSON(http.StatusOK, gin.H{
				"message": "save file",
			})
			return
		} else {
			println("quiz" + seq + " not found")
		}
	}

	c.AbortWithStatus(http.StatusNotModified)
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
