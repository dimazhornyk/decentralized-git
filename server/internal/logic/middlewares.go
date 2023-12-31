package logic

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (s *service) JwtAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		wallet, err := s.tokenManager.ExtractTokenWallet(c)
		if err != nil {
			fmt.Println(err.Error())
			c.String(http.StatusUnauthorized, "Unauthorized")
			c.Abort()

			return
		}

		c.Set("wallet", wallet)
		c.Next()
	}
}
