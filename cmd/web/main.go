package main

import (
	"flag"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"gitlab.com/code-mobi/psu-tep/pkg/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const (
	taskDir   = "task"
	answerDir = "answer"
)

func main() {
	addr := flag.String("addr", ":4000", "HTTP Network Address")
	signedString := flag.String("secret", os.Getenv("SIGNED_STRING"), "Signed String")
	dsn := flag.String("dsn", os.Getenv("PSU_DSN"), "Database DSN")
	frontPath := flag.String("front-dir", os.Getenv("FRONT_PATH"), "Frontend Store Path")
	storePath := flag.String("store-dir", os.Getenv("STORE_PATH"), "File Store Path")
	db, err := gorm.Open(mysql.Open(*dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	handler := newHandler(db, *signedString, *storePath)

	db.AutoMigrate(&models.User{}, &models.Task{}, &models.Examinee{}, &models.Score{})

	handler.initUser()

	r := gin.Default()

	r.GET("/", func(ctx *gin.Context) {
		ctx.Redirect(http.StatusFound, "/index")
	})
	api := r.Group("/api")
	api.POST("/login", handler.loginHandler)
	api.POST("/login_examinee", handler.loginExamineeHandler)

	apiAdminProtected := r.Group("/api/admin", handler.authorizationMiddleware, handler.authorizationAdminMiddleware)
	apiAdminProtected.GET("/users", handler.listUserHandler)
	apiAdminProtected.PATCH("/user/:id", handler.updateUserHandler)
	apiAdminProtected.PATCH("/task/:id", handler.saveTaskHandler)
	apiAdminProtected.GET("/examinee/:id", handler.getExamineeHandler)
	apiAdminProtected.POST("/examinee", handler.createExamineeHandler)
	apiAdminProtected.PATCH("/examinee/:id", handler.updateExamineeHandler)
	apiAdminProtected.GET("/examinees", handler.listExamineeHandler)
	apiAdminProtected.POST("/examinees", handler.uploadExamineeHandler)
	apiAdminProtected.GET("/examinees/scores", handler.listExamineeByAdminHandler)
	apiAdminProtected.GET("/examinees/scores/download", handler.exportScores)
	apiAdminProtected.GET("/examinees/answers/download", handler.downloadAnswers)
	apiAdminProtected.POST("/clear_data", handler.clearDataHandler)

	apiRaterProtected := r.Group("/api/rater", handler.authorizationMiddleware, handler.authorizationRaterMiddleware)
	apiRaterProtected.GET("/examinees", handler.listExamineeByRaterHandler)
	apiRaterProtected.POST("/score", handler.rateExamineeHandler)

	apiUserProtected := r.Group("/api", handler.authorizationMiddleware, handler.authorizationUserMiddleware)
	apiUserProtected.GET("/refresh_token", handler.refreshTokenHandler)

	apiProtected := r.Group("/api", handler.authorizationMiddleware)
	apiProtected.GET("/task", handler.getTaskHandler)

	apiExamineeProtected := r.Group("/api/examinee", handler.authorizationMiddleware, handler.authorizationExamineeMiddleware)
	apiExamineeProtected.POST("answer/:ansNum", handler.sendAnswerHandler)

	r.Static("/index", *frontPath)
	r.Static("/"+taskDir, *storePath+"/"+taskDir)
	r.Static("/"+answerDir, *storePath+"/"+answerDir)
	r.Run(*addr)
}
