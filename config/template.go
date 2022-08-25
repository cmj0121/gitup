package config

import (
	"fmt"
	"html/template"
	"os"
	"strings"

	_ "embed"
	log "github.com/sirupsen/logrus"
)

var (
	// the unique key of the blog's template
	KEY_BLOG_TMPL = "tmpl_blog_html"

	//go:embed assets/blog.htm
	TMPL_HTML string
	//go:embed assets/list.htm
	TMPL_LIST_HTML string
	//go:embed assets/blog.css
	TMPL_STYLE string
)

// the customized template of the HTML
type Render struct {
	// the brand of the HTML page
	Brand string

	// the template of the HTML page
	Html string `yaml:",omitempty"`

	// the template of the post-list HTML page
	ListHtmp string `yaml:",omitempty"`

	// the style of the HTML page
	Style string `yaml:",omitempty"`
}

// get the HTML template
func (render Render) Template() (tmpl *template.Template, err error) {
	var text string
	if text, err = render.html(render.Html, TMPL_HTML); err != nil {
		// cannot get the template text
		return
	}

	tmpl, err = render.renderTemplate(text)
	return
}

// get the list/HTML template
func (render Render) ListTemplate() (tmpl *template.Template, err error) {
	var text string
	if text, err = render.html(render.ListHtmp, TMPL_LIST_HTML); err != nil {
		// cannot get the template text
		return
	}

	tmpl, err = render.renderTemplate(text)
	return
}

func (render Render) renderTemplate(text string) (tmpl *template.Template, err error) {
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
		"css": func(text interface{}) template.CSS {
			switch text := text.(type) {
			case string:
				return template.CSS(text)
			case template.HTML:
				return template.CSS(text)
			default:
				return template.CSS(fmt.Sprintf("%v", text))
			}
		},
	}).Parse(text)
	return
}

// get the html template text
func (render Render) html(filepath, default_html string) (text string, err error) {
	switch filepath {
	case "":
		text = default_html
	default:
		var data []byte

		if data, err = os.ReadFile(filepath); err != nil {
			log.WithFields(log.Fields{
				"path":  filepath,
				"error": err,
			}).Warn("cannot read HTMP template")
		}

		text = string(data)
	}

	return
}

// get the CSS style
func (render Render) CSS() (css template.CSS) {
	switch render.Style {
	case "":
		css = template.CSS(TMPL_STYLE)
	default:
		data, err := os.ReadFile(render.Html)
		if err != nil {
			log.WithFields(log.Fields{
				"path":  render.Html,
				"error": err,
			}).Warn("cannot read HTMP template")
			return
		}

		css = template.CSS(data)
	}

	return
}
