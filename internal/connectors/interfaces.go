package connectors

import "git-test/internal/common"

type Storage interface {
}

type Repository interface {
	GetUser(wallet string) (common.User, error)
	GetUserByActionToken(token string) (common.User, error)
	CreateUser(wallet string) (string, string, error)
}
