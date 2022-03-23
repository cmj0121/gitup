package gitup

import (
	"github.com/alecthomas/kong"
)

// the GitUp instance
type GitUp struct {
	*CLI
}

// create new GitUp instance
func New() *GitUp {
	return &GitUp{
		// create the CLI with default value
		CLI: &CLI{},
	}
}

func (gitup *GitUp) Run() {
	// parse the passed argument and may exit immediately
	kong.Parse(gitup.CLI)
}
