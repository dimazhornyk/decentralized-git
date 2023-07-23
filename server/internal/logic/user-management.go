package logic

import (
	"fmt"
	"git-test/siwe"
	"github.com/gin-gonic/gin"
	"net/http"
)

type walletRequestBody struct {
	Message   string
	Signature string
}

func (s *service) Login(c *gin.Context) {
	var req walletRequestBody
	if err := c.BindJSON(&req); err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("can't parse request:%s", err.Error()))
		return
	}

	address, err := s.getMessageAddress(req)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("can't get message address: %s", err.Error()))
		return
	}

	if _, err := s.repo.GetUser(address); err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("user does not exist"))
		return
	}

	token, err := s.tokenManager.GenerateToken(address)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("can't generate token: %s", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
	})
}

func (s *service) Register(c *gin.Context) {
	var req walletRequestBody
	if err := c.BindJSON(&req); err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("can't parse request:%s", err.Error()))
		return
	}

	address, err := s.getMessageAddress(req)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("can't get message address: %s", err.Error()))
		return
	}

	actionToken, encryptionKey, err := s.repo.CreateUser(address)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("can't create user: %s", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"action_token":   actionToken,
		"encryption_key": encryptionKey,
	})
}

func (s *service) getMessageAddress(req walletRequestBody) (string, error) {
	message, err := siwe.ParseMessage(req.Message)
	if err != nil {
		return "", fmt.Errorf("can't parse messageStr: %s", err.Error())
	}

	return message.GetAddress().String(), nil
}
