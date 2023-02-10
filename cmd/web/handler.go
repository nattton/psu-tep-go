package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"gorm.io/gorm"
)

type Handler struct {
	db           *gorm.DB
	signedString string
	storePath    string
}

type UserClaim struct {
	Role string `json:"role"`
	ID   string `json:"id"`
	jwt.RegisteredClaims
}

func newHandler(db *gorm.DB, signedString string, storePath string) *Handler {
	return &Handler{db: db, signedString: signedString, storePath: storePath}
}

func (h *Handler) authorizationMiddleware(c *gin.Context) {
	s := c.Request.Header.Get("Authorization")

	token := strings.TrimPrefix(s, "Bearer ")

	if err := h.validateToken(token); err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
}

func (h *Handler) validateToken(token string) error {
	_, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}

		return []byte(h.signedString), nil
	})

	return err
}

func (h *Handler) decodeToken(c *gin.Context) (UserClaim, error) {
	s := c.Request.Header.Get("Authorization")
	token := strings.TrimPrefix(s, "Bearer ")
	var userClaim UserClaim

	_, err := jwt.ParseWithClaims(token, &userClaim, func(token *jwt.Token) (interface{}, error) {
		return []byte(h.signedString), nil
	})

	return userClaim, err
}

func (h *Handler) authorizationAdminMiddleware(c *gin.Context) {
	h.authorizationMiddleware(c)
	user, _ := h.decodeToken(c)
	if user.Role != "admin" {
		c.AbortWithStatus(http.StatusUnauthorized)
	}
}

func (h *Handler) authorizationRaterMiddleware(c *gin.Context) {
	h.authorizationMiddleware(c)
	user, _ := h.decodeToken(c)
	if user.Role != "rater" {
		c.AbortWithStatus(http.StatusUnauthorized)
	}
}

func (h *Handler) authorizationUserMiddleware(c *gin.Context) {
	h.authorizationMiddleware(c)
	user, _ := h.decodeToken(c)
	if !(user.Role == "admin" || user.Role == "rater") {
		c.AbortWithStatus(http.StatusUnauthorized)
	}
}
