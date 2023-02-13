package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
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
	idString := c.Param("id")
	id, err := strconv.Atoi(idString)
	if err != nil || id == 0 || id > 3 {
		c.AbortWithStatus(http.StatusBadRequest)
	}
	file, err := c.FormFile("quiz")
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
	}

	var quiz models.Quiz
	h.db.First(&quiz)
	quizPath := fmt.Sprintf("/%s/", quizDir)
	if err := os.MkdirAll(h.storePath+quizPath, os.ModePerm); err != nil {
		log.Fatal(err)
	}

	filename := fmt.Sprintf("%d_%d_%s", id, time.Now().Unix(), file.Filename)
	if err := c.SaveUploadedFile(file, h.storePath+quizPath+filename); err != nil {
		log.Fatal(err)
	}

	h.db.Model(&quiz).Update("quiz"+idString, quizPath+filename)

	c.JSON(http.StatusOK, gin.H{
		"message": "save file",
	})
}

func (h *Handler) listExamineeByAdminHandler(c *gin.Context) {
	userClaim, _ := h.decodeToken(c)
	var user models.User

	if err := h.db.First(&user, userClaim.ID).Error; err != nil {
		c.AbortWithStatus(http.StatusNotModified)
		return
	}

	var examinees []models.Examinee
	h.db.Preload("Scores.User").Preload("Scores", func(db *gorm.DB) *gorm.DB {
		return db.Order("scores.user_id ASC")
	}).Find(&examinees)
	for i := 0; i < len(examinees); i++ {
		examinees[i] = addPathToAnswer(c, examinees[i])
	}
	c.JSON(http.StatusOK, gin.H{
		"examinees": examinees,
	})
}

func (h *Handler) exportScores(c *gin.Context) {
	userClaim, _ := h.decodeToken(c)
	var user models.User

	if err := h.db.First(&user, userClaim.ID).Error; err != nil {
		c.AbortWithStatus(http.StatusNotModified)
		return
	}

	var examinees []models.Examinee
	h.db.Preload("Scores.User").Preload("Scores", func(db *gorm.DB) *gorm.DB {
		return db.Order("scores.user_id ASC")
	}).Find(&examinees)
	for i := 0; i < len(examinees); i++ {
		examinees[i] = addPathToAnswer(c, examinees[i])
	}

	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()
	_, err := f.NewSheet("Sheet1")
	if err != nil {
		fmt.Println(err)
		return
	}
	rowNum := 1
	cell, err := excelize.CoordinatesToCellName(1, rowNum)
	if err != nil {
		fmt.Println(err)
		return
	}
	f.SetSheetRow("Sheet1", cell, &[]interface{}{"Code", "Firstname", "Lastname", "Rate by", "Answer1", "Answer2", "Answer3", "Total"})

	for _, ex := range examinees {
		for _, score := range ex.Scores {
			rowNum++
			cell, err := excelize.CoordinatesToCellName(1, rowNum)
			if err != nil {
				fmt.Println(err)
				return
			}
			sumScore := score.Answer1 + score.Answer2 + score.Answer3
			f.SetSheetRow("Sheet1", cell, &[]interface{}{ex.Code, ex.Firstname, ex.Lastname, score.User.Name, score.Answer1, score.Answer2, score.Answer3, sumScore})
		}
	}

	filePath := fmt.Sprintf("%s/%s/Score.xlsx", h.storePath, quizDir)
	if err := f.SaveAs(filePath); err != nil {
		fmt.Println(err)
	}
	c.File(filePath)
}

func (h *Handler) listExamineeByRaterHandler(c *gin.Context) {
	userClaim, _ := h.decodeToken(c)
	var user models.User

	if err := h.db.First(&user, userClaim.ID).Error; err != nil {
		c.AbortWithStatus(http.StatusNotModified)
		return
	}

	var examinees []models.Examinee
	h.db.Preload("Scores.User").Preload("Scores", "user_id = ?", user.ID).Find(&examinees)
	for i := 0; i < len(examinees); i++ {
		examinees[i] = addPathToAnswer(c, examinees[i])
	}
	c.JSON(http.StatusOK, gin.H{
		"examinees": examinees,
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

	userClaim, _ := h.decodeToken(c)

	var examinee models.Examinee
	var user models.User

	if err := h.db.First(&examinee, form.ExamineeID).Error; err != nil {
		c.AbortWithStatus(http.StatusNotModified)
		return
	}

	if err := h.db.First(&user, userClaim.ID).Error; err != nil {
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
		c.JSON(http.StatusOK, gin.H{
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

func (h *Handler) downloadAnswers(c *gin.Context) {
	filePath, err := zipAnswerWriter(h.storePath)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
	}

	c.File(filePath)
}
