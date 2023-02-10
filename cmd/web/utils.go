package main

import (
	"github.com/gin-gonic/gin"
)

func getCurrentPath(c *gin.Context) string {
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	return scheme + "://" + c.Request.Host

}
