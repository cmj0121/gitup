package blog

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"os"
	"regexp"
	"sort"
	"time"

	"github.com/cmj0121/gitup/config"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"

	log "github.com/sirupsen/logrus"
)

// the blog/post instance
type Blog struct {
	// the source blog/markdown filepath
	Path string `short:"p" arg:"" type:"existingfile" help:"the blog/markdown filepath"`

	// the output file path
	Output string `short:"o" type:"path" default:"test.htm" help:"the destinate folder of the generated webpage"`
	// the customized title
	Title string `short:"t" help:"the customized title"`
	// the description of the blog
	Description string `kong:"-"`

	// the blogs timestamp
	CreatedAt time.Time
	UpdatedAt time.Time

	md   []byte // the raw markdown context
	html []byte // the raw HTML page
}

// create the blog from the open file
func New(reader io.Reader) (blog *Blog, err error) {
	var buff bytes.Buffer

	if _, err = io.Copy(&buff, reader); err != nil {
		// cannot read and save to buffer
		log.WithFields(log.Fields{
			"error": err,
		}).Warn("cannot read blog")
		return
	}

	blog = &Blog{
		md: buff.Bytes(),
	}

	return
}

// generate the blog via passwd arguments
func (blog *Blog) Run(config *config.Config) (err error) {
	var reader io.Reader

	if reader, err = os.Open(blog.Path); err != nil {
		log.WithFields(log.Fields{
			"path":  blog.Path,
			"error": err,
		}).Warn("cannot open blog file")
		return
	}

	var buff bytes.Buffer
	if _, err = io.Copy(&buff, reader); err != nil {
		// cannot read and save to buffer
		log.WithFields(log.Fields{
			"path":  blog.Path,
			"error": err,
		}).Warn("cannot read blog")
		return
	}

	blog.md = buff.Bytes()
	err = blog.Write(config, nil)
	return
}

func (blog *Blog) Dup() (dup *Blog) {
	dup = &Blog{
		Path: blog.Path,

		Output:      blog.Output,
		Title:       blog.Title,
		Description: blog.Description,

		CreatedAt: blog.CreatedAt,
		UpdatedAt: blog.UpdatedAt,

		md:   blog.md,
		html: blog.html,
	}
	return
}

// render the blog from markdown to HTML page
func (blog *Blog) Render(config *config.Config) (text []byte, err error) {
	if _, err = blog.RenderHTML(); err != nil {
		// cannot get the HTML page
		return
	}

	text = blog.html
	return
}

// render the raw HTML from markdown
func (blog *Blog) RenderHTML() (text []byte, err error) {
	if text = blog.html; len(text) == 0 {
		// the parser settings
		extensions := parser.CommonExtensions | parser.AutoHeadingIDs
		extensions |= parser.Titleblock
		extensions |= parser.Footnotes
		extensions |= parser.SuperSubscript
		extensions |= parser.Mmark

		parser := parser.NewWithExtensions(extensions)

		// the render settings
		htmlFlags := html.CommonFlags | html.HrefTargetBlank | html.TOC | html.LazyLoadImages
		htmlFlags |= html.NofollowLinks | html.NoreferrerLinks | html.NoopenerLinks

		opts := html.RendererOptions{Flags: htmlFlags}
		render := html.NewRenderer(opts)

		text = markdown.ToHTML(blog.md, parser, render)
		blog.html = text

		// find the post title
		if blog.Title == "" {
			RE_TITLE := regexp.MustCompile(`<h1 id=.*?>([\s\S]+?)</h1>`)

			if text := blog.html; RE_TITLE.Match(text) {
				// find the title
				for _, matched := range RE_TITLE.FindAllSubmatch(text, -1) {
					if blog.Title = string(matched[1]); blog.Title != "" {
						// found the first <h1> tag
						break
					}
				}
			}
		}

		RE_DESC := regexp.MustCompile(`<blockquote>\s*(:?<.*?>)*\s*([^<]+?)\s*(:?<.*?>)*\s*</blockquote>`)
		if RE_DESC.Match(text) {
			// find the description
			blog.Description = string(RE_DESC.FindAllSubmatch(text, -1)[0][2])
		}
	}

	return
}

// write blog to destination
func (blog *Blog) Write(conf *config.Config, summary Summary) (err error) {
	var writer io.Writer

	switch blog.Output {
	case "", "-":
		writer = os.Stdout
	default:
		var file *os.File
		file, err = os.OpenFile(blog.Output, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0640)
		if err != nil {
			log.WithFields(log.Fields{
				"path":  blog.Output,
				"error": err,
			}).Warn("cannot write as HTML")
			return
		}
		defer file.Close()
		writer = file
	}

	if _, err = blog.RenderHTML(); err != nil {
		log.WithFields(log.Fields{
			"path":  blog.Output,
			"error": err,
		}).Warn("cannot write as HTML")
		return
	}

	var tmpl *template.Template
	if tmpl, err = conf.Template(); err != nil {
		// cannot get the template from the config
		return
	}
	err = tmpl.Execute(writer, struct {
		*config.Config
		*Blog
		Summary
		Style template.CSS

		// the extra meta
		UTCNow time.Time
	}{
		Config:  conf,
		Blog:    blog,
		Summary: summary,
		Style:   conf.CSS(),

		UTCNow: time.Now().UTC(),
	})

	return
}

// the unique ID of the blog
func (blog Blog) UID() (uid string) {
	// the unique ID is the created at as micro seconds based on UTC+0
	uid = fmt.Sprintf("%v", blog.CreatedAt.UTC().UnixMicro()/1000000)
	return
}

// the rendered HTML
func (blog Blog) HTML() (html string) {
	html = string(blog.html)
	return
}

// the sort.Interface of []Blog
type Blogs []*Blog

// the number of elements in the collection.
func (blogs Blogs) Len() (size int) {
	size = len(blogs)
	return
}

// reports whether the element with index i must sort
// before the element with index j.
func (blogs Blogs) Less(i, j int) (less bool) {
	less = blogs[i].CreatedAt.UnixNano() > blogs[j].CreatedAt.UnixNano()
	return
}

// swaps the elements with indexes i and j.
func (blogs Blogs) Swap(i, j int) {
	blogs[i], blogs[j] = blogs[j], blogs[i]
}

// the summary via the year
func (blogs Blogs) SummaryByYear() (summary Summary) {
	years_category := map[string]Blogs{}

	for _, blog := range blogs {
		year := fmt.Sprintf("%v", blog.CreatedAt.UTC().Year())

		switch _, ok := years_category[year]; ok {
		case true:
			years_category[year] = append(years_category[year], blog)
		case false:
			years_category[year] = Blogs{blog}
		}
	}

	years := []string{}
	for year := range years_category {
		years = append(years, year)
	}

	sort.Sort(sort.Reverse(sort.StringSlice(years)))

	for _, year := range years {
		sort.Sort(years_category[year])

		summary = append(summary, &Category{
			Key:   year,
			Blogs: years_category[year],
		})
	}

	return
}
