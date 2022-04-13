package clone

import (
	"net/url"

	"github.com/cmj0121/gitup/config"
)

// the clone instance
type Clone struct {
	// the remote repository URI
	Repo *url.URL `arg:"" help:"the remote repository"`
}

// clone the repository and generate the webpage
func (clone *Clone) Run(conf *config.Config) (err error) {
	return
}
