package common

type User struct {
	Wallet        string
	ActionToken   string
	EncryptionKey string
	Repos         []string
}

type NonUTFFile struct {
	Content []byte
}

type UTFFile struct {
	DeletedLines  []int
	InsertedLines map[int]string // key: start line, value: content
}
