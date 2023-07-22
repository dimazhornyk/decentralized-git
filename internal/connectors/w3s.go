package connectors

import (
	"git-test/internal/common"
	"github.com/web3-storage/go-w3s-client"
)

type storage struct {
	client w3s.Client
}

func NewStorage(cfg *common.Config) (Storage, error) {
	c, err := w3s.NewClient(w3s.WithToken(cfg.StorageToken))
	if err != nil {
		return nil, err
	}

	return storage{
		client: c,
	}, nil
}

func (s storage) Upload() {
	//s.client.Put()
}
