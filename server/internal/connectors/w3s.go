package connectors

import (
	"context"
	"git-test/internal/common"
	"github.com/ipfs/go-cid"
	"github.com/web3-storage/go-w3s-client"
	"io/fs"
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

func (s storage) Upload(f fs.File) (cid.Cid, error) {
	return s.client.Put(context.Background(), f)
}

func (s storage) GetRepoFiles(id cid.Cid) (fs.FS, error) {
	resp, err := s.client.Get(context.Background(), id)
	if err != nil {
		return nil, err
	}

	_, fsys, err := resp.Files()
	if err != nil {
		return nil, err
	}

	return fsys, nil
}
