package clone

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/cmj0121/gitup/blog"
	"github.com/cmj0121/gitup/config"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
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
func (clone *Clone) Run(config *config.Config) (err error) {
	clone.tempdir = fmt.Sprintf("%v/gitup.%d", os.TempDir(), os.Getpid())
	clone.tempdir = filepath.Clean(clone.tempdir)

	defer func() {
		if clone.Purge {
			os.RemoveAll(clone.tempdir) // nolint
		}
	}()

	var repo *git.Repository
	if repo, err = clone.Clone(); err != nil {
		log.WithFields(log.Fields{
			"repository": clone.Repo,
			"error":      err,
		}).Warn("clone repository")
		return
	}

	// load the customized config from repo
	config.Load(clone.tempdir)

	for _, dir := range config.Workdir {
		if err = clone.Process(config, dir); err != nil {
			log.WithFields(log.Fields{
				"path":  dir,
				"error": err,
			}).Error("generate blog fail")
		}
	}

	err = clone.Generate(config, repo)
	return
}

// clone the repo to local temporary folder
func (clone *Clone) Clone() (repo *git.Repository, err error) {
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
	if repo, err = git.PlainClone(clone.tempdir, false, &options); err != nil {
		// cannot clone from remote to local
		return
	}

	return
}

// process and generate HTML from specified folder
func (clone *Clone) Process(config *config.Config, dir string) (err error) {
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

			var md_blog *blog.Blog
			if md_blog, err = clone.process(config, md_path); err != nil {
				// parse the blog/markdown fail
				return
			}

			clone.blogs = append(clone.blogs, md_blog)
		}
	}

	return
}

// generate the final webpage
func (clone *Clone) Generate(config *config.Config, repo *git.Repository) (err error) {
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
	if err = clone.find_first_commit(repo, clone.blogs); err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Warn("cannot find the blogs first commit time")
		return
	}

	summary := clone.blogs.SummaryByYear()
	for _, blog := range clone.blogs {
		basename := filepath.Base(filepath.Clean(blog.Path))
		basename = basename[:len(basename)-len(filepath.Ext(basename))]

		dest_path := fmt.Sprintf("%v/%v-%v.htm", clone.Output, blog.UID(), basename)
		dest_path = filepath.Clean(dest_path)
		if dest_path[:len(clone.Output)] != clone.Output {
			err = fmt.Errorf("invalid desc path: %v", dest_path)
			return
		}

		blog.Output = dest_path
		blog.Link = blog.Output[len(clone.Output)+1:]
	}

	for _, blog := range clone.blogs {
		if err = blog.Write(config, summary); err != nil {
			// cannot write to description
			return
		}
	}

	err = clone.generate_default_pages(config, summary)
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
func (clone *Clone) process(config *config.Config, path string) (md_blog *blog.Blog, err error) {
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

	if md_blog, err = blog.New(file); err != nil {
		log.WithFields(log.Fields{
			"path":  path,
			"error": err,
		}).Info("cannot gen blog/markdown")
		return
	}
	// only record the related path of the blog/markdown
	md_blog.Path = path[len(clone.tempdir)+1:]
	if _, err = md_blog.RenderHTML(); err != nil {
		// cannot render HTML from blog
		return
	}

	return
}

// find the blog first commit date
func (clone *Clone) find_first_commit(repo *git.Repository, blogs blog.Blogs) (err error) {
	md_path_idx_map := map[string]int{}

	for idx, blog := range blogs {
		md_path_idx_map[blog.Path] = idx
	}

	options := git.LogOptions{
		// only trace the file with the blog list
		PathFilter: func(path string) (ok bool) {
			_, ok = md_path_idx_map[path]
			return
		},
	}

	var commit_iter object.CommitIter
	if commit_iter, err = repo.Log(&options); err != nil {
		log.WithFields(log.Fields{
			"repo":  repo,
			"error": err,
		}).Warn("cannot process git-log")

		return
	}

	err = commit_iter.ForEach(func(commit *object.Commit) (err error) {
		var stats object.FileStats

		if stats, err = commit.Stats(); err != nil {
			log.WithFields(log.Fields{
				"commit": commit,
				"error":  err,
			}).Warn("cannot get commit status")
			return
		}

		for _, st := range stats {
			idx, ok := md_path_idx_map[st.Name]
			if !ok {
				log.WithFields(log.Fields{
					"commit": commit,
					"file":   st.Name,
				}).Warn("find file in commit but not in blog list")
				continue
			}

			if blogs[idx].UpdatedAt.IsZero() {
				// only setup the updated if not beed set
				blogs[idx].UpdatedAt = commit.Author.When
			}
			blogs[idx].CreatedAt = commit.Author.When
		}
		return
	})

	return
}

// generate the default pages
func (clone *Clone) generate_default_pages(config *config.Config, summary blog.Summary) (err error) {
	sort.Sort(clone.blogs)

	// render the newest post as the index.htm
	blog := clone.blogs[0].Dup()
	blog.Output = fmt.Sprintf("%v/index.htm", clone.Output)
	blog.Output = filepath.Clean(blog.Output)
	blog.Link = "index.htm"
	if err = blog.Write(config, summary); err != nil {
		// cannot write index.htm
		return
	}

	// render the post-list
	path := fmt.Sprintf("%v/post-list.htm", clone.Output)
	path = filepath.Clean(path)
	if err = summary.Write(config, path); err != nil {
		// cannot write the summary page
		return
	}

	if config.Settings.AboutMe != "" {
		err = clone.generate_default_page(config, summary, config.Settings.AboutMe, "about-me.htm")
		if err != nil {
			// cannot render about-me.htm
			return
		}
	}
	if config.Settings.License != "" {
		err = clone.generate_default_page(config, summary, config.Settings.License, "license.htm")
		if err != nil {
			// cannot render about-me.htm
			return
		}
	}

	err = clone.generate_favicon(config)
	return
}

func (clone *Clone) generate_default_page(config *config.Config, summary blog.Summary, src, dest string) (err error) {
	var md_blog *blog.Blog

	md_path := fmt.Sprintf("%v/%v", clone.tempdir, src)
	md_path = filepath.Clean(md_path)
	if md_blog, err = clone.process(config, md_path); err != nil {
		// cannot load about-me.htm
		return
	}

	md_blog.Output = fmt.Sprintf("%v/%v", clone.Output, dest)
	md_blog.Output = filepath.Clean(md_blog.Output)
	md_blog.Link = dest
	err = md_blog.Write(config, summary)

	return
}

func (clone *Clone) generate_favicon(conf *config.Config) (err error) {
	var favicon []byte

	switch conf.Favicon {
	case "":
		favicon = config.DEFAULT_FAVICON
	default:
		src := fmt.Sprintf("%v/%v", clone.tempdir, conf.Favicon)
		src = filepath.Clean(src)

		if favicon, err = ioutil.ReadFile(src); err != nil {
			log.WithFields(log.Fields{
				"path":  src,
				"error": err,
			}).Warn("cannot open favicon")
			return
		}
	}

	path := fmt.Sprintf("%v/%v", clone.Output, conf.FaviconLink())
	path = filepath.Clean(path)

	err = ioutil.WriteFile(path, favicon, 0640)
	return
}
