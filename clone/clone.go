package clone

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/cmj0121/gitup/config"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"

	log "github.com/sirupsen/logrus"
)

const (
	SUFFIX_MD       = ".md"
	SUFFIX_MARKDOWN = ".markdown"
)

// the clone instance
type Clone struct {
	// the remote repository URI
	Repo *url.URL `arg:"" help:"the remote repository"`

	// Auth with username/password
	Username string `short:"U" help:"the username used for auth"`
	Password string `short:"P" help:"the password used for auth"`

	// remove the temporary folder
	Purge bool `short:"p" negatable:"" default:"true" help:"purge the temporary repo cloned from remote"`
}

// clone the repository and generate the webpage
func (clone *Clone) Run(conf *config.Config) (err error) {
	tmpdir := clone.TempDir()
	log.WithFields(log.Fields{
		"path": tmpdir,
	}).Info("the local folder to store the repo")

	defer func() {
		if clone.Purge {
			if err := os.RemoveAll(tmpdir); err != nil {
				log.WithFields(log.Fields{
					"path": tmpdir,
				}).Info("cannot purge temporary folder")
			}
		}
	}()

	if err = clone.Clone(tmpdir); err != nil {
		log.WithFields(log.Fields{
			"repository": clone.Repo,
			"error":      err,
		}).Warn("clone repository")
		return
	}

	// load the customized config from repo
	conf.Load(tmpdir)

	for _, dir := range conf.Workdir {
		if err = clone.Process(dir); err != nil {
			log.WithFields(log.Fields{
				"path":  dir,
				"error": err,
			}).Error("generate blog fail")
		}
	}
	return
}

// clone the repo to local temporary folder
func (clone *Clone) Clone(tmpdir string) (err error) {
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

// process and generate HTML from specified folder
func (clone *Clone) Process(dir string) (err error) {
	working_space := clone.TempDir()
	path := filepath.Clean(fmt.Sprintf("%v/%v", working_space, dir))

	if path[:len(working_space)] != working_space {
		err = fmt.Errorf("invalid folder path: %v", path)
		return
	}

	var files []os.DirEntry
	if files, err = os.ReadDir(path); err != nil {
		log.WithFields(log.Fields{
			"path":  path,
			"error": err,
		}).Warn("cannot list blog")
		return
	}

	for _, file := range files {
		name := file.Name()

		switch {
		case name[0] == '.':
			// the hidden file, skip
		case strings.HasSuffix(name, SUFFIX_MD) || strings.HasSuffix(name, SUFFIX_MARKDOWN):
			// parse the blog/markdown
			if err = clone.process(name); err != nil {
				// parse the blog/markdown fail
				return
			}
		}
	}

	return
}

// parse the single blog/markdown by path
func (clone *Clone) process(path string) (err error) {
	log.WithFields(log.Fields{
		"path": path,
	}).Trace("process the blog/markdown")
	return
}
