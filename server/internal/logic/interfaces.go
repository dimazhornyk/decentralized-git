package logic

import "github.com/gin-gonic/gin"

type Service interface {
	Login(c *gin.Context)
	Register(c *gin.Context)
	UploadArchive(c *gin.Context)

	GetRepos(c *gin.Context)
	DownloadRepo(c *gin.Context)

	JwtAuthMiddleware() gin.HandlerFunc
}

type TokenManager interface {
	GenerateToken(wallet string) (string, error)
	IsTokenValid(c *gin.Context) error
	ExtractToken(c *gin.Context) string
	ExtractTokenWallet(c *gin.Context) (string, error)
}
