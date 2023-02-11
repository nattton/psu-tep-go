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
	quizDir   = "quiz"
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

	db.AutoMigrate(&models.User{}, &models.Quiz{}, &models.Examinee{}, &models.Score{})

	handler.initUser()

	r := gin.Default()

	r.GET("/", func(ctx *gin.Context) {
		ctx.Redirect(http.StatusFound, "/index")
	})
	api := r.Group("/api")
	api.POST("/login", handler.loginHandler)
	api.POST("/login_examinee", handler.loginExamineeHandler)

	apiAdminProtected := r.Group("/api", handler.authorizationMiddleware, handler.authorizationAdminMiddleware)
	apiAdminProtected.GET("/users", handler.listUserHandler)
	apiAdminProtected.PATCH("/user/:id", handler.updateUserHandler)
	apiAdminProtected.PATCH("/quiz", handler.saveQuizHandler)
	apiAdminProtected.GET("/admin/examinees", handler.listExamineeByAdminHandler)

	apiRaterProtected := r.Group("/api", handler.authorizationMiddleware, handler.authorizationRaterMiddleware)
	apiRaterProtected.GET("/rater/examinees", handler.listExamineeByRaterHandler)
	apiRaterProtected.POST("/rater/score", handler.rateExamineeHandler)

	apiUserProtected := r.Group("/api", handler.authorizationMiddleware, handler.authorizationUserMiddleware)
	apiUserProtected.GET("/examinees", handler.listExamineeHandler)

	apiProtected := r.Group("/api", handler.authorizationMiddleware)
	apiProtected.GET("/examinee/:id", handler.getExamineeHandler)
	apiProtected.POST("/examinee", handler.createExamineeHandler)
	apiProtected.PATCH("/examinee/:id", handler.updateExamineeHandler)

	apiProtected.GET("/quiz", handler.getQuizHandler)
	apiProtected.POST("/answer", handler.sendAnswerHandler)

	r.Static("/index", *frontPath)
	r.Static("/"+quizDir, *storePath+"/"+quizDir)
	r.Static("/"+answerDir, *storePath+"/"+answerDir)
	r.Run(*addr)
}
