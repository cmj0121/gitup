package blog

import (
	"bytes"
	"io"
	"os"

	"github.com/cmj0121/gitup/config"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"

	log "github.com/sirupsen/logrus"
)

// the blog/post instance
type Blog struct {
	Path string `short:"p" arg:"" type:"existingfile" help:"the blog/markdown filepath"`

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
func (blog *Blog) Run(conf *config.Config) (err error) {
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

	var text []byte
	if text, err = blog.Render(conf); err != nil {
		// cannot render the blog to HTML
		return
	}

	os.Stdout.Write(text)
	return
}

// render the blog from markdown to HTML page
func (blog *Blog) Render(conf *config.Config) (text []byte, err error) {
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
	}

	return
}
