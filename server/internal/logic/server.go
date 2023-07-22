package logic

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func NewGin(service Service) (*gin.Engine, error) {
	r := gin.Default()
	r.Use(cors.Default())

	r.POST("/upload", service.UploadArchive)
	r.POST("/login", service.Login)
	r.POST("/register", service.Register)

	protected := r.Group("/api")
	protected.Use(service.JwtAuthMiddleware())
	protected.POST("/getRepos", service.GetRepos)
	protected.POST("/downloadRepo", service.DownloadRepo)

	return r, nil
}
