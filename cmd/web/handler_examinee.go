package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"gitlab.com/code-mobi/psu-tep/pkg/forms"
	"gitlab.com/code-mobi/psu-tep/pkg/models"
)

func (h *Handler) loginExamineeHandler(c *gin.Context) {
	var login forms.Examinee
	if err := c.ShouldBindJSON(&login); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	var examinee models.Examinee
	if err := h.db.Where(
		&models.Examinee{
			Code:      login.Code,
			Firstname: login.Firstname,
			Lastname:  login.Lastname}).
		First(&examinee).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "examinee not found",
		})
		return
	}

	claims := &UserClaim{
		"examinee",
		strconv.FormatUint(uint64(examinee.ID), 10),
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(6 * time.Hour)),
			Issuer:    "code-mobi.com",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	ss, err := token.SignedString([]byte(h.signedString))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"examinee": examinee,
		"token":    ss,
	})
}

func (h *Handler) listExamineeHandler(c *gin.Context) {
	var examinees []models.Examinee
	h.db.Order("code").Find(&examinees)
	c.JSON(http.StatusOK, gin.H{
		"store_path": getCurrentPath(c),
		"examinees":  examinees,
	})
}

func (h *Handler) getExamineeHandler(c *gin.Context) {
	id := c.Param("id")
	var examinee models.Examinee
	if err := h.db.First(&examinee, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"store_path": getCurrentPath(c),
		"examinee":   examinee,
	})
}

func (h *Handler) createExamineeHandler(c *gin.Context) {
	var form forms.Examinee
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	examinee := models.Examinee{
		Code:      form.Code,
		Firstname: form.Firstname,
		Lastname:  form.Lastname,
	}

	if err := h.db.Create(&examinee).Error; err != nil {
		c.AbortWithStatus(http.StatusNotModified)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "create examinee",
	})
}

func (h *Handler) updateExamineeHandler(c *gin.Context) {
	id := c.Param("id")
	var form forms.Examinee
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if len(form.Code) < 1 || len(form.Firstname) < 1 || len(form.Lastname) < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "data should not be empty",
		})
		return
	}

	var examinee models.Examinee
	h.db.First(&examinee, id)
	examinee.Code = form.Code
	examinee.Firstname = form.Firstname
	examinee.Lastname = form.Lastname

	if err := h.db.Save(&examinee).Error; err != nil {
		c.AbortWithStatus(http.StatusNotModified)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "update examinee",
	})
}

func (h *Handler) sendAnswerHandler(c *gin.Context) {
	claims, _ := h.decodeToken(c)
	if claims.Role != "examinee" {
		c.JSON((http.StatusUnauthorized), gin.H{
			"error": claims.Role + " unauthorized this function.",
		})
		return
	}

	println("ID", claims.ID)
	println("Role", claims.Role)
	var examinee models.Examinee
	h.db.First(&examinee, claims.ID)

	answerPath := fmt.Sprintf("/%s/%s_%s/", answerDir, strconv.FormatUint(uint64(examinee.ID), 10), examinee.Code)
	if err := os.MkdirAll(h.storePath+answerPath, os.ModePerm); err != nil {
		log.Fatal(err)
	}

	for i := 1; i < 4; i++ {
		seq := fmt.Sprint(i)
		file, err := c.FormFile("answer" + seq)
		if err == nil {
			filename := examinee.Code + "_" + seq + "_" + file.Filename
			if err := c.SaveUploadedFile(file, h.storePath+answerPath+filename); err != nil {
				log.Fatal(err)
			}

			h.db.Model(&examinee).Update("answer"+seq, answerPath+filename)
			if i == 3 {
				h.db.Model(&examinee).Update("finish", true)
			}

			c.JSON(http.StatusCreated, gin.H{
				"message": "save file",
			})
			return
		} else {
			println("answer" + seq + " not found")
		}
	}

	// c.AbortWithStatus(http.StatusNotModified)
}
