package config

// the customized settings of the blog
type Settings struct {
	// the path of the about-me page
	AboutMe string `yaml:"abount_me,omitempty"`

	// the path of the license page
	License string `yaml:"license,omitempty"`
}
