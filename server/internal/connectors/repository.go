package connectors

import (
	"errors"
	"git-test/internal/common"
	"github.com/ipfs/go-cid"
)

// TODO: a proper db must be used here, but it's not a priority for demoable MVP
type repository struct {
	users map[string]common.User
}

func NewRepository(cfg *common.Config) (Repository, error) {
	return &repository{
		users: map[string]common.User{
			"0xC87CB8CEa4426c87Fb53e9BD64e0530BDe1cef2b": {
				Wallet:        "0xC87CB8CEa4426c87Fb53e9BD64e0530BDe1cef2b",
				ActionToken:   "nBjBwLvhZrjyrywCSMlkrOpoupuGPSfH",
				EncryptionKey: "HmsQxFWISzgMPGzBaNVchpeuoLhEBOUK",
				Repos:         map[string][]cid.Cid{},
			},
		},
	}, nil
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
		return common.User{}, errors.New("unknown token")
	}

	return res, nil
}

func (r *repository) CreateUser(wallet string) (string, string, error) {
	if _, ok := r.users[wallet]; ok {
		return "", "", errors.New("user already exists")
	}

	token, encryptionKey := common.RandStringRunes(32), common.RandStringRunes(32)
	user := common.User{
		Wallet:        wallet,
		ActionToken:   token,
		EncryptionKey: encryptionKey,
		Repos:         map[string][]cid.Cid{},
	}
	r.users[wallet] = user

	return token, encryptionKey, nil
}

func (r *repository) SaveRepoVersion(wallet, repoName string, id cid.Cid) error {
	if _, ok := r.users[wallet]; !ok {
		return errors.New("user does not exist")
	}

	if _, ok := r.users[wallet].Repos[repoName]; !ok {
		r.users[wallet].Repos[repoName] = []cid.Cid{}
	}

	r.users[wallet].Repos[repoName] = append(r.users[wallet].Repos[repoName], id)

	return nil
}

func (r *repository) GetRepoIDs(wallet, repoName string) ([]cid.Cid, error) {
	user, ok := r.users[wallet]
	if !ok {
		return nil, errors.New("user not found")
	}

	return user.Repos[repoName], nil
}
