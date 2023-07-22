package connectors

import (
	"errors"
	"git-test/internal/common"
)

type repository struct {
	users map[string]common.User
}

func NewRepository(cfg *common.Config) (Repository, error) {
	return &repository{}, nil
}

func (r *repository) GetUser(wallet string) (common.User, error) {
	user, ok := r.users[wallet]
	if !ok {
		return common.User{}, errors.New("user not found")
	}

	return user, nil
}

func (r *repository) GetUserByActionToken(token string) (common.User, error) {
	var res common.User
	for _, user := range r.users {
		if user.ActionToken == token {
			res = user
			break
		}
	}

	if res.Wallet == "" {
		return common.User{}, errors.New("user not found")
	}

	return res, nil
}

func (r *repository) CreateUser(wallet string) (string, string, error) {
	if _, ok := r.users[wallet]; ok {
		return "", "", errors.New("user already exists")
	}

	token, encryptionKey := common.RandStringRunes(64), common.RandStringRunes(64)
	user := common.User{
		Wallet:        wallet,
		ActionToken:   token,
		EncryptionKey: encryptionKey,
	}
	r.users[wallet] = user

	return token, encryptionKey, nil
}
