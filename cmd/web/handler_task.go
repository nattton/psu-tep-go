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

func (h *Handler) getTaskHandler(c *gin.Context) {
	var task models.Task
	h.db.First(&task)
	currentPath := getCurrentPath(c)
	task0, task1, task2, task3 := "", "", "", ""
	if task.Task0 != "" {
		task0 = currentPath + task.Task0
	}
	if task.Task1 != "" {
		task1 = currentPath + task.Task1
	}
	if task.Task2 != "" {
		task2 = currentPath + task.Task2
	}
	if task.Task3 != "" {
		task3 = currentPath + task.Task3
	}
	c.JSON(http.StatusOK, gin.H{
		"task0": task0,
		"task1": task1,
		"task2": task2,
		"task3": task3,
	})
}

func (h *Handler) saveTaskHandler(c *gin.Context) {
	idString := c.Param("id")
	print("idString " + idString)
	id, err := strconv.Atoi(idString)
	if err != nil || id < 0 || id > 3 {
		c.AbortWithStatus(http.StatusBadRequest)
	}
	file, err := c.FormFile("task")
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
	}

	var task models.Task
	h.db.First(&task)
	taskPath := fmt.Sprintf("/%s/", taskDir)
	if err := os.MkdirAll(h.storePath+taskPath, os.ModePerm); err != nil {
		log.Fatal(err)
	}

	filename := fmt.Sprintf("%d_%d_%s", id, time.Now().Unix(), file.Filename)
	if err := c.SaveUploadedFile(file, h.storePath+taskPath+filename); err != nil {
		log.Fatal(err)
	}

	h.db.Model(&task).Update("task"+idString, taskPath+filename)

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
	sheetName := "Sheet1"
	_, err := f.NewSheet(sheetName)
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
	f.SetSheetRow(sheetName, cell, &[]interface{}{"Test Taker ID", "Firstname", "Lastname", "Task1 / rater1", "Task1 / rater2", "Task2 / rater1", "Task2 / rater2", "Task3 / rater1", "Task3 / rater2"})

	for _, ex := range examinees {
		rowNum++
		cell, _ := excelize.CoordinatesToCellName(1, rowNum)
		f.SetSheetRow(sheetName, cell, &[]interface{}{ex.Code, ex.Firstname, ex.Lastname})
		for _, score := range ex.Scores {
			if score.UserID == 2 {
				cell, _ := excelize.CoordinatesToCellName(4, rowNum)
				f.SetCellFloat(sheetName, cell, float64(score.Task1), 2, 32)
				cell, _ = excelize.CoordinatesToCellName(6, rowNum)
				f.SetCellFloat(sheetName, cell, float64(score.Task2), 2, 32)
				cell, _ = excelize.CoordinatesToCellName(8, rowNum)
				f.SetCellFloat(sheetName, cell, float64(score.Task2), 2, 32)
			} else if score.UserID == 3 {
				cell, _ := excelize.CoordinatesToCellName(5, rowNum)
				f.SetCellFloat(sheetName, cell, float64(score.Task1), 2, 32)
				cell, _ = excelize.CoordinatesToCellName(7, rowNum)
				f.SetCellFloat(sheetName, cell, float64(score.Task2), 2, 32)
				cell, _ = excelize.CoordinatesToCellName(9, rowNum)
				f.SetCellFloat(sheetName, cell, float64(score.Task3), 2, 32)
			}
		}
	}
	fileDir := h.storePath + "/temp"
	os.MkdirAll(fileDir, os.ModePerm)
	filePath := fmt.Sprintf("%s/Score.xlsx", fileDir)
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
	score.Task1 = form.Task1
	score.Task2 = form.Task2
	score.Task3 = form.Task3
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
