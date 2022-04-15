package config

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"strings"

	_ "embed"
	log "github.com/sirupsen/logrus"
)

var (
	// the unique key of the blog's template
	KEY_BLOG_TMPL = "tmpl_blog_html"

	//go:embed assets/blog.htm
	TMPL_HTML string
	//go:embed assets/blog.css
	TMPL_STYLE string
)

// the customized template of the HTML
type Render struct {
	// the brand of the HTML page
	Brand string

	// the template of the HTML page
	Html string `yaml:",omitempty"`

	// the style of the HTML page
	Style string `yaml:",omitempty"`
}

// get the HTML template
func (render Render) Template() (tmpl *template.Template, err error) {
	var text string
	if text, err = render.html(); err != nil {
		// cannot get the template text
		return
	}

	tmpl, err = template.New(KEY_BLOG_TMPL).Funcs(template.FuncMap{
		"safe": func(text string) template.HTML {
			return template.HTML(text)
		},
		"indent": func(num_indent int, text interface{}) template.HTML {
			indent := "\n" + strings.Repeat(" ", num_indent)

			switch text := text.(type) {
			case string:
				return template.HTML(strings.Replace(text, "\n", indent, -1))
			case template.HTML:
				return template.HTML(strings.Replace(string(text), "\n", indent, -1))
			default:
				return template.HTML(strings.Replace(fmt.Sprintf("%v", text), "\n", indent, -1))
			}
		},
	}).Parse(text)
	return
}

// get the html template text
func (render Render) html() (text string, err error) {
	switch render.Html {
	case "":
		text = TMPL_HTML
	default:
		var data []byte

		if data, err = ioutil.ReadFile(render.Html); err != nil {
			log.WithFields(log.Fields{
				"path":  render.Html,
				"error": err,
			}).Warn("cannot read HTMP template")
		}

		text = string(data)
	}

	return
}
