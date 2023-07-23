package logic

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
)

func NewGin(service Service) (*gin.Engine, error) {
	r := gin.Default()
	r.Use(CORSMiddleware())
	r.Use(ginBodyLogMiddleware)

	r.GET("/nonce", service.GetNonce)
	r.POST("/upload", service.UploadArchive)
	r.POST("/login", service.Login)
	r.POST("/register", service.Register)

	protected := r.Group("/api")
	protected.Use(service.JwtAuthMiddleware())
	protected.GET("/getRepos", service.GetRepos)
	protected.GET("/downloadRepo", service.DownloadRepo)

	return r, nil
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func ginBodyLogMiddleware(c *gin.Context) {
	blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
	c.Writer = blw
	c.Next()
	statusCode := c.Writer.Status()
	if statusCode >= 400 {
		//ok this is an request with error, let's make a record for it
		// now print body (or log in your preferred way)
		fmt.Println("Response body: " + blw.body.String())
	}
}
