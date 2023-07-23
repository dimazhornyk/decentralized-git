package logic

import (
	"archive/zip"
	"fmt"
	"git-test/siwe"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
)

func (s *service) GetRepos(c *gin.Context) {
	wallet, exists := c.Get("wallet")
	if !exists {
		c.String(http.StatusInternalServerError, "no wallet in request context")
		return
	}

	user, err := s.repo.GetUser(wallet.(string))
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("can't get user by wallet: %v", err.Error()))
	}

	var repos []string
	for repo := range user.Repos {
		repos = append(repos, repo)
	}

	c.JSON(http.StatusOK, gin.H{
		"repos": repos,
	})
}

func (s *service) DownloadRepo(c *gin.Context) {
	wallet, exists := c.Get("wallet")
	if !exists {
		c.String(http.StatusInternalServerError, "no wallet in request context")
		return
	}

	walletStr := wallet.(string)
	repoName := c.Param("repo")
	user, err := s.repo.GetUser(walletStr)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("can't get user: %s", err.Error()))
		return
	}

	files, err := s.getOldFiles(walletStr, repoName, user.EncryptionKey)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("error getting old files: %s", err.Error()))
		return
	}

	path := fmt.Sprintf("%s/%s.zip", walletStr, repoName)
	file, err := os.Create(path)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("error creating an archive: %s", err.Error()))
		return
	}
	defer file.Close()

	w := zip.NewWriter(file)
	defer w.Close()

	for filename, b := range files {
		f, err := w.Create(filename)
		if err != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("error adding file to an archive: %s", err.Error()))
			return
		}

		if _, err := f.Write(b); err != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("error writing to file in an archive: %s", err.Error()))
			return
		}
	}

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", "attachment; filename=out.zip")
	c.Header("Content-Type", "application/octet-stream")
	c.File(path)
}

func (s *service) GetNonce(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"nonce": siwe.GenerateNonce(),
	})
}
