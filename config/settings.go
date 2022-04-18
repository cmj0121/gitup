package config

import (
	"path/filepath"

	_ "embed"
)

var (
	//go:embed assets/favicon.png
	DEFAULT_FAVICON []byte

	DEFAULT_FAVICON_LINK = "favicon.png"
)

// the customized settings of the blog
type Settings struct {
	// the path of the about-me page
	AboutMe string `yaml:"abount_me,omitempty"`

	// the path of the license page
	License string `yaml:"license,omitempty"`

	// the FavIcon of the path
	Favicon string `yaml:"favicon,omitempty"`
}

func (settings Settings) FaviconLink() (link string) {
	switch settings.Favicon {
	case "":
		link = DEFAULT_FAVICON_LINK
	default:
		link = filepath.Base(settings.Favicon)
	}

	return
}
