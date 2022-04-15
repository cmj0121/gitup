package clone

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/cmj0121/gitup/blog"
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

	// the final destinate folder of the webpage
	Output string `short:"o" type:"path" default:"build" help:"the destinate folder of the generated webpage"`

	// Auth with username/password
	Username string `short:"U" help:"the username used for auth"`
	Password string `short:"P" help:"the password used for auth"`

	// remove the temporary folder
	Purge bool `short:"p" negatable:"" default:"true" help:"purge the temporary repo cloned from remote"`

	tempdir string     // the working space
	blogs   blog.Blogs // the processed blog instances
}

// clone the repository and generate the webpage
func (clone *Clone) Run(conf *config.Config) (err error) {
	clone.tempdir = fmt.Sprintf("%v/gitup.%d", os.TempDir(), os.Getpid())
	clone.tempdir = filepath.Clean(clone.tempdir)

	defer func() {
		if clone.Purge {
			os.RemoveAll(clone.tempdir) //nolint
		}
	}()

	if err = clone.Clone(); err != nil {
		log.WithFields(log.Fields{
			"repository": clone.Repo,
			"error":      err,
		}).Warn("clone repository")
		return
	}

	// load the customized config from repo
	conf.Load(clone.tempdir)

	for _, dir := range conf.Workdir {
		if err = clone.Process(conf, dir); err != nil {
			log.WithFields(log.Fields{
				"path":  dir,
				"error": err,
			}).Error("generate blog fail")
		}
	}

	err = clone.Generate()
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
	if _, err = git.PlainClone(clone.tempdir, false, &options); err != nil {
		// cannot clone from remote to local
		return
	}

	return
}

// process and generate HTML from specified folder
func (clone *Clone) Process(conf *config.Config, dir string) (err error) {
	path := filepath.Clean(fmt.Sprintf("%v/%v", clone.tempdir, dir))

	if path[:len(clone.tempdir)] != clone.tempdir {
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
			md_path := fmt.Sprintf("%v/%v", path, name)
			md_path = filepath.Clean(md_path)

			if md_path[:len(path)] != path {
				log.WithFields(log.Fields{
					"path": md_path,
				}).Info("invalid blog/markdown path")
				continue
			}

			if err = clone.process(conf, md_path); err != nil {
				// parse the blog/markdown fail
				return
			}
		}
	}

	return
}

// generate the final webpage
func (clone *Clone) Generate() (err error) {
	if _, err := os.Stat(clone.Output); err == nil {
		// always remove the description folder if exists
		os.RemoveAll(clone.Output) // nolint
	}

	if err = os.MkdirAll(clone.Output, 0750); err != nil {
		log.WithFields(log.Fields{
			"path":  clone.Output,
			"error": err,
		}).Warn("cannot create description folder")
		return
	}

	// sort by the blog
	sort.Sort(clone.blogs)

	for _, blog := range clone.blogs {
		basename := filepath.Base(filepath.Clean(blog.Path))
		basename = basename[:len(basename)-len(filepath.Ext(basename))]

		dest_path := fmt.Sprintf("%v/%v.htm", clone.Output, basename)
		dest_path = filepath.Clean(dest_path)
		if dest_path[:len(clone.Output)] != clone.Output {
			err = fmt.Errorf("invalid desc path: %v", dest_path)
			return
		}

		blog.Output = dest_path
		if err = blog.Write(); err != nil {
			// cannot write to description
			return
		}
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

// parse the single blog/markdown by path
func (clone *Clone) process(conf *config.Config, path string) (err error) {
	log.WithFields(log.Fields{
		"path": path,
	}).Trace("process the blog/markdown")

	var file *os.File
	if file, err = os.Open(path); err != nil {
		log.WithFields(log.Fields{
			"path":  path,
			"error": err,
		}).Info("cannot open blog/markdown")
		return
	}
	defer file.Close()

	var md_blog *blog.Blog
	if md_blog, err = blog.New(file); err != nil {
		log.WithFields(log.Fields{
			"path":  path,
			"error": err,
		}).Info("cannot gen blog/markdown")
		return
	}
	md_blog.Path = path

	clone.blogs = append(clone.blogs, md_blog)
	return
}
