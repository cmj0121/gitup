package gitup

import (
	"io"
	"os"

	"github.com/alecthomas/kong"
	log "github.com/sirupsen/logrus"
)

// the GitUp instance
type GitUp struct {
	logger io.WriteCloser

	*CLI
}

// create new GitUp instance
func New() *GitUp {
	return &GitUp{
		// default log writer
		logger: os.Stderr,
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
func (gitup *GitUp) prologue() (err error) {
	// setup the log sub-system
	formatter := log.TextFormatter{
		// can setup the color from environment
		EnvironmentOverrideColors: true,
		// show the RFC-3389 timestamp
		FullTimestamp: true,
	}

	log.SetFormatter(&formatter)
	switch gitup.CLI.LogFile {
	case "":
		gitup.logger = os.Stderr
	case "-":
		gitup.logger = os.Stdout
	default:
		gitup.logger, err = os.OpenFile(gitup.LogFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0640)
	}

	verbose := int(log.ErrorLevel) + gitup.CLI.Verbose
	log.SetLevel(log.Level(verbose))
	log.SetOutput(gitup.logger)

	log.Trace("setup the log sub-system")
	return
}

func (gitup *GitUp) epilogue() {
	gitup.logger.Close()
}
