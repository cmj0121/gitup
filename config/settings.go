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

	// the list of the hidden posts
	// it should be the related path in local repo
	Hidden []string `yaml:"hidden,omitempty"`

	// disabled the generated HTML file with timestamp as prefix
	DisabledTimestampPrefix bool `yaml:"disabled_timestamp_prefix,omitempty"`
}

// return the Favicon link
func (settings Settings) FaviconLink() (link string) {
	switch settings.Favicon {
	case "":
		link = DEFAULT_FAVICON_LINK
	default:
		link = filepath.Base(settings.Favicon)
	}

	return
}

// check the path is set as hidden or not
func (settings Settings) IsHidden(path string) (hidden bool) {
	for idx := range settings.Hidden {
		if settings.Hidden[idx] == path {
			hidden = true
		}
	}

	return
}
