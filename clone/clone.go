package clone

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/cmj0121/gitup/config"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"

	log "github.com/sirupsen/logrus"
)

// the clone instance
type Clone struct {
	// the remote repository URI
	Repo *url.URL `arg:"" help:"the remote repository"`

	// Auth with username/password
	Username string `short:"U" help:"the username used for auth"`
	Password string `short:"P" help:"the password used for auth"`
}

// clone the repository and generate the webpage
func (clone *Clone) Run(conf *config.Config) (err error) {
	if err = clone.Clone(); err != nil {
		log.WithFields(log.Fields{
			"repository": clone.Repo,
			"error":      err,
		}).Warn("clone repository")
		return
	}

	return
}

// clone the repo to local temporary folder
func (clone *Clone) Clone() (err error) {
	var auth transport.AuthMethod
	if auth, err = clone.auth_method(); err != nil {
		// cannot get the auth method
		return
	}

	// clone options
	options := git.CloneOptions{
		Auth: auth,
		URL:  clone.Repo.String(),
	}

	tmpdir := clone.TempDir()
	log.WithFields(log.Fields{
		"path": tmpdir,
	}).Info("the local folder to store the repo")

	if _, err = git.PlainClone(tmpdir, false, &options); err != nil {
		// cannot clone from remote to local
		return
	}

	return
}

// get the auth method from the provided URI
func (clone *Clone) auth_method() (auth transport.AuthMethod, err error) {
	switch scheme := clone.Repo.Scheme; scheme {
	case "http", "https":
		// generatl HTTP/HTTPS repository
		auth = &http.BasicAuth{
			Username: clone.Username,
			Password: clone.Password,
		}
	default:
		err = fmt.Errorf("not support scheme: %v", scheme)
		return
	}

	return
}

// the temporary folder
func (clone *Clone) TempDir() (folder string) {
	folder = fmt.Sprintf("%v/gitup.%d", os.TempDir(), os.Getpid())
	folder = filepath.Clean(folder)
	return
}
