package logic

import (
	"fmt"
	"git-test/internal/common"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

type tokenManager struct {
	jwtKey      string
	jwtDuration time.Duration
}

func NewTokenManager(cfg *common.Config) TokenManager {
	return &tokenManager{}
}

func (t *tokenManager) GenerateToken(wallet string) (string, error) {
	claims := jwt.MapClaims{}
	claims["authorized"] = true
	claims["wallet"] = wallet
	claims["exp"] = time.Now().Add(time.Hour * t.jwtDuration).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(t.jwtKey))

}

func (t *tokenManager) IsTokenValid(c *gin.Context) error {
	tokenString := t.ExtractToken(c)
	if _, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(t.jwtKey), nil
	}); err != nil {
		return err
	}

	return nil
}

func (t *tokenManager) ExtractToken(c *gin.Context) string {
	bearerToken := c.Request.Header.Get("Authorization")
	if len(strings.Split(bearerToken, " ")) == 2 {
		return strings.Split(bearerToken, " ")[1]
	}

	return ""
}

func (t *tokenManager) ExtractTokenWallet(c *gin.Context) (string, error) {
	tokenString := t.ExtractToken(c)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(t.jwtKey), nil
	})
	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", fmt.Errorf("invalid token")
	}

	wallet, ok := claims["wallet"]
	if !ok {
		return "", fmt.Errorf("no wallet in token")
	}

	walletStr, ok := wallet.(string)
	if !ok {
		return "", fmt.Errorf("invalid wallet in token")
	}

	return walletStr, nil
}
