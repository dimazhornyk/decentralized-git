package logic

import (
	"archive/zip"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/gob"
	"fmt"
	"git-test/internal/common"
	"git-test/internal/connectors"
	"github.com/gin-gonic/gin"
	"github.com/ipfs/go-cid"
	w3fs "github.com/web3-storage/go-w3s-client/fs"
	"io"
	"io/fs"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

const (
	repoNameKey    = "repo_name"
	actionTokenKey = "action_token"
	archiveKey     = "archive"
)

type service struct {
	storage      connectors.Storage
	repo         connectors.Repository
	tokenManager TokenManager
}

func NewService(tokenManager TokenManager, repo connectors.Repository, storage connectors.Storage) (Service, error) {
	return &service{
		tokenManager: tokenManager,
		storage:      storage,
		repo:         repo,
	}, nil
}

func (s *service) UploadArchive(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("get form error: %s", err.Error()))
		return
	}

	if err := validateForm(form); err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("form validation error: %s", err.Error()))
		return
	}

	actionToken := form.Value[actionTokenKey][0]
	repo := form.Value[repoNameKey][0]
	archive := form.File[archiveKey][0]

	user, err := s.repo.GetUserByActionToken(actionToken)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("get user error: %s", err.Error()))
		return
	}

	oldFiles, err := s.getOldFiles(user.Wallet, repo, user.EncryptionKey)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("get old files error: %s", err.Error()))
		return
	}

	newFiles, err := readZip(archive)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	diffs, err := getDiffs(oldFiles, newFiles)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	encoded, err := encode(diffs, []byte(user.EncryptionKey))
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	contentID, err := s.uploadNewDiffs(user.Wallet, repo, encoded)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	if err := s.repo.SaveRepoVersion(user.Wallet, repo, contentID); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.String(http.StatusOK, "Files successfully uploaded!")
}

func (s *service) getOldFiles(walletAddress, repoFullName, encryptionKey string) (map[string][]byte, error) {
	ids, err := s.repo.GetRepoIDs(walletAddress, repoFullName)
	if err != nil {
		return nil, err
	}

	filesState := make(map[string][]byte)
	for _, id := range ids {
		fsys, err := s.storage.GetRepoFiles(id)
		if err != nil {
			return nil, err
		}

		files := make(map[string][]byte)
		if err := fs.WalkDir(fsys, "/", func(path string, d fs.DirEntry, err error) error {
			f, err := fsys.Open(d.Name())
			if err != nil {
				return err
			}

			buf := new(bytes.Buffer)
			if _, err := io.Copy(buf, f); err != nil {
				return err
			}

			decrypted, err := s.decryptFile(buf.Bytes(), []byte(encryptionKey))
			files[d.Name()] = decrypted

			return nil
		}); err != nil {
			return nil, err
		}

		if err := s.applyDiffs(filesState, files); err != nil {
			return nil, err
		}
	}

	return filesState, nil
}

func (s *service) applyDiffs(currState map[string][]byte, diffs map[string][]byte) error {
	buff := new(bytes.Buffer)
	dec := gob.NewDecoder(buff)
	for filename, f := range diffs {
		if _, err := buff.Write(f); err != nil {
			return fmt.Errorf("error writing to buffer: %w", err)
		}

		curr := currState[filename]

		var utfFile common.UTFFile
		if err := dec.Decode(&utfFile); err == nil {
			merged, err := s.mergeWithDiffs(curr, utfFile)
			if err != nil {
				return err
			}
			currState[filename] = merged
		} else {
			var nonUTFFile common.NonUTFFile
			if _, err := buff.Write(f); err != nil {
				return fmt.Errorf("error writing to buffer: %w", err)
			}

			if err := dec.Decode(&nonUTFFile); err != nil {
				return fmt.Errorf("can't decode a file: %w", err)
			}

			currState[filename] = nonUTFFile.Content
		}
	}

	return nil
}

func (s *service) mergeWithDiffs(bs []byte, file common.UTFFile) ([]byte, error) {
	if len(bs) == 0 {
		resStrs := make([]string, len(file.InsertedLines))
		for idx, line := range file.InsertedLines {
			resStrs[idx] = line
		}

		return []byte(strings.Join(resStrs, "\n")), nil
	}

	str := string(bs)
	lines := strings.SplitAfter(str, "\n")
	res := make([]string, 0)
	for i, line := range lines {
		var found bool
		for _, n := range file.DeletedLines {
			if i == n {
				found = true
			}
		}

		if !found {
			res = append(res, line)
		}
	}

	for i, line := range file.InsertedLines {
		res = append(res[:i+1], res[i:]...)
		res[i] = line
	}

	return []byte(strings.Join(res, "\n")), nil
}

func (s *service) decryptFile(b []byte, key []byte) ([]byte, error) {
	var res []byte
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	iv := b[:aes.BlockSize]
	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(b[aes.BlockSize:], res)

	return res, nil
}

func (s *service) uploadNewDiffs(wallet, repoName string, files map[string][]byte) (cid.Cid, error) {
	dir := fmt.Sprintf("%s/%s", wallet, repoName)
	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			fmt.Printf("cleanup error: %s\n", err.Error())
		}
	}()

	var openFiles []fs.File
	for name, f := range files {
		if strings.HasSuffix(name, "/") || strings.Contains(name, ".git") || strings.Contains(name, ".idea") {
			continue
		}

		filename := fmt.Sprintf("%s/%s/%s", wallet, repoName, name)
		if err := os.MkdirAll(filepath.Dir(filename), 0777); err != nil {
			return cid.Cid{}, err
		}
		if err := os.WriteFile(filename, f, 0644); err != nil {
			return cid.Cid{}, err
		}

		open, err := os.Open(filename)
		if err != nil {
			return cid.Cid{}, err
		}
		openFiles = append(openFiles, open)
	}

	defer func() {
		for _, f := range openFiles {
			f.Close()
		}
	}()

	newDir := w3fs.NewDir(wallet, openFiles)

	//f, err := os.Open(fmt.Sprintf("%s/%s/", wallet, repoName))
	//if err != nil {
	//	return cid.Cid{}, err
	//}

	id, err := s.storage.Upload(newDir)
	if err != nil {
		return cid.Cid{}, err
	}

	return id, nil
}

func getDiffs(oldFiles map[string][]byte, newFiles []*zip.File) (map[string]any, error) {
	res := make(map[string]any)
	for _, f := range newFiles {
		r, err := f.Open()
		if err != nil {
			return nil, fmt.Errorf("open file error: %s", err.Error())
		}

		b, err := io.ReadAll(r)
		if err != nil {
			return nil, fmt.Errorf("read file error: %s", err.Error())
		}

		if !utf8.Valid(b) {
			res[f.Name] = common.NonUTFFile{
				Content: b,
			}

			continue
		}

		if oldFile, ok := oldFiles[f.Name]; !ok {
			res[f.Name] = newFileToModel(string(b))
		} else {
			diffs := common.Diff("oldFile", oldFile, "newFile", b)
			if len(diffs) != 0 {
				res[f.Name] = parseDiffs(diffs)
			}
		}
	}

	return res, nil
}

func encode(files map[string]any, key []byte) (map[string][]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create cipher error: %s", err.Error())
	}

	var buff bytes.Buffer
	enc := gob.NewEncoder(&buff)
	res := make(map[string][]byte, len(files))
	for name, f := range files {
		if err := enc.Encode(f); err != nil {
			return nil, fmt.Errorf("encode file error: %s", err.Error())
		}

		b := buff.Bytes()
		encrypted := make([]byte, aes.BlockSize+len(b))
		iv := encrypted[:aes.BlockSize]
		if _, err := io.ReadFull(rand.Reader, iv); err != nil {
			panic(err)
		}
		stream := cipher.NewCFBEncrypter(block, iv)
		stream.XORKeyStream(encrypted[aes.BlockSize:], b)

		res[name] = encrypted
	}

	return res, nil
}

func newFileToModel(content string) common.UTFFile {
	lines := strings.SplitAfter(content, "\n")
	res := common.UTFFile{InsertedLines: make(map[int]string, len(lines))}

	for i, line := range lines {
		res.InsertedLines[i] = line
	}

	return res
}

func parseDiffs(diffs []byte) common.UTFFile {
	lines := strings.Split(string(diffs), "\n")
	res := common.UTFFile{InsertedLines: make(map[int]string), DeletedLines: make([]int, 0)}

	deletionIdx := 0
	insertionIdx := 0
	for _, line := range lines[4:] {
		switch {
		case strings.HasPrefix(line, "-"):
			res.DeletedLines = append(res.DeletedLines, deletionIdx)
			deletionIdx++
		case strings.HasPrefix(line, "+"):
			res.InsertedLines[insertionIdx] = line[1:]
			insertionIdx++
		case strings.HasPrefix(line, "\\"):
			continue
		default:
			deletionIdx++
			insertionIdx++
		}
	}

	return res
}

func readZip(zipHeader *multipart.FileHeader) ([]*zip.File, error) {
	f, err := zipHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("open zip error: %s", err.Error())
	}

	buf := bytes.NewBuffer(nil)
	if _, err := buf.ReadFrom(f); err != nil {
		return nil, fmt.Errorf("read zip error: %s", err.Error())
	}

	b := buf.Bytes()
	reader, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		return nil, fmt.Errorf("zip reader error: %s", err.Error())
	}

	return reader.File, nil
}

func validateForm(form *multipart.Form) error {
	if form == nil {
		return fmt.Errorf("form is empty")
	}

	if len(form.Value[repoNameKey]) == 0 {
		return fmt.Errorf("repo_name is empty")
	}

	if len(form.Value[actionTokenKey]) == 0 {
		return fmt.Errorf("wallet_address is empty")
	}

	if len(form.File[archiveKey]) == 0 {
		return fmt.Errorf("archive is empty")
	}

	return nil
}
