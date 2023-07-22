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
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	repoNameKey    = "repo_name"
	actionTokenKey = "action_token"
	archiveKey     = "archive"
)

type service struct {
	repo         connectors.Repository
	tokenManager TokenManager
}

func NewService(conf *common.Config) (Service, error) {
	return &service{}, nil
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

	oldFiles, err := s.getOldFiles(user.Wallet, repo)
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

	files, err := s.uploadNewDiffs(repo, encoded)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	defer func() {
		for _, f := range files {
			if err := os.Remove(f); err != nil {
				fmt.Printf("remove file error: %s\n", err.Error())
			}
		}
	}()

	c.String(http.StatusOK, "Files successfully uploaded!")
}

// TODO
func (s *service) getOldFiles(walletAddress, repoFullName string) (map[string][]byte, error) {
	return map[string][]byte{}, nil
}

func (s *service) uploadNewDiffs(repoName string, files map[string][]byte) ([]string, error) {
	finalNames := make([]string, 0, len(files))
	for name, f := range files {
		filename := fmt.Sprintf("%s/%s/%d", repoName, name, int(time.Now().Unix()))
		if err := os.WriteFile(filename, f, 0644); err != nil {
			return nil, err
		}
		finalNames = append(finalNames, filename)
	}

	return finalNames, nil
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
