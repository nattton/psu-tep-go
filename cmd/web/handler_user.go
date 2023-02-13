package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"gitlab.com/code-mobi/psu-tep/pkg/forms"
	"gitlab.com/code-mobi/psu-tep/pkg/models"
)

func (h *Handler) loginHandler(c *gin.Context) {
	var login forms.Login
	if err := c.ShouldBindJSON(&login); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	var user models.User
	h.db.Where("name = ?", login.Username).First(&user)

	if user.ID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user or password not match",
		})
		return
	}

	if err := user.VerifyUser(login.Password); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	claims := &UserClaim{
		user.Role,
		strconv.FormatUint(uint64(user.ID), 10),
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			Issuer:    "code-mobi.com",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString([]byte(h.signedString))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"user":  user,
		"token": ss,
	})
}

func (h *Handler) refreshTokenHandler(c *gin.Context) {
	userClaim, _ := h.decodeToken(c)

	var user models.User
	if err := h.db.First(&user, userClaim.ID).Error; err != nil {
		c.AbortWithStatus(http.StatusNotModified)
		return
	}

	if user.ID == 0 {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	claims := &UserClaim{
		user.Role,
		strconv.FormatUint(uint64(user.ID), 10),
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			Issuer:    "code-mobi.com",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString([]byte(h.signedString))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"user":  user,
		"token": ss,
	})
}

func (h *Handler) listUserHandler(c *gin.Context) {
	var users []models.User
	h.db.Find(&users)
	c.JSON(http.StatusOK, gin.H{
		"users": users,
	})
}

func (h *Handler) updateUserHandler(c *gin.Context) {
	id := c.Param("id")
	var form forms.Login
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	var user models.User
	h.db.First(&user, id)
	user.Name = form.Username
	if form.Password != "" {
		user.NewPassword = form.Password
	}

	if err := h.db.Save(&user).Error; err != nil {
		c.AbortWithStatus(http.StatusNotModified)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "update user",
	})
}
