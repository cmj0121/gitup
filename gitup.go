package gitup

import (
	"os"

	"github.com/alecthomas/kong"
	log "github.com/sirupsen/logrus"
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

// parse the passed argument and execute command
func (gitup *GitUp) Run() {
	// parse the passed argument and may exit immediately
	kong.Parse(gitup.CLI)

	gitup.prologue()
}

// setup everything need before execute command
func (gitup *GitUp) prologue() {
	// setup the log sub-system
	formatter := log.TextFormatter{
		// can setup the color from environment
		EnvironmentOverrideColors: true,
		// show the RFC-3389 timestamp
		FullTimestamp: true,
	}

	log.SetFormatter(&formatter)
	log.SetOutput(os.Stderr)

	verbose := int(log.ErrorLevel) + gitup.CLI.Verbose
	log.SetLevel(log.Level(verbose))

	log.Trace("setup the log sub-system")
}
