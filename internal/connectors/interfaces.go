package connectors

import (
	"git-test/internal/common"
	"github.com/ipfs/go-cid"
	"io/fs"
)

type Storage interface {
	Upload(f fs.File) (cid.Cid, error)
	GetRepoFiles(id cid.Cid) (fs.FS, error)
}

type Repository interface {
	GetUser(wallet string) (common.User, error)
	GetUserByActionToken(token string) (common.User, error)
	CreateUser(wallet string) (string, string, error)
	SaveRepoVersion(wallet, repoName string, id cid.Cid) error
	GetRepoIDs(wallet, repoName string) ([]cid.Cid, error)
}
