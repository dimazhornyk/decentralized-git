package common

import "github.com/ipfs/go-cid"

type User struct {
	Wallet        string
	ActionToken   string
	EncryptionKey string
	Repos         map[string][]cid.Cid // repoName -> versions
}

type NonUTFFile struct {
	Content []byte
}

type UTFFile struct {
	DeletedLines  []int
	InsertedLines map[int]string // key: start line, value: content
}
