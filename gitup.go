package gitup

import (
	"io"
	"os"

	"github.com/alecthomas/kong"
	"github.com/cmj0121/gitup/config"

	log "github.com/sirupsen/logrus"
)

// the GitUp instance
type GitUp struct {
	logger io.WriteCloser
	*config.Config

	*CLI
}

// create new GitUp instance
func New() *GitUp {
	return &GitUp{
		// default log writer
		logger: os.Stderr,
		// default config
		Config: &config.Config{
			Project: Version(),
			Author:  AUTHOR,

			Render: config.Render{
				Brand: Version(),
			},
		},
		// create the CLI with default value
		CLI: &CLI{},
	}
}

// parse the passed argument and execute command
func (gitup *GitUp) Run() {
	// parse the passed argument and may exit immediately
	ctx := kong.Parse(gitup.CLI)

	if err := gitup.prologue(); err != nil {
		// cannot run prepare steps
		panic(err)
	}
	defer gitup.epilogue()

	// run the command
	err := ctx.Run(gitup.Config)
	ctx.FatalIfErrorf(err)
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
		if err != nil {
			// cannot open log file
			return
		}
	}

	verbose := int(log.ErrorLevel) + gitup.CLI.Verbose
	log.SetLevel(log.Level(verbose))
	log.SetOutput(gitup.logger)

	log.Trace("setup the log sub-system")

	if gitup.Settings != "" {
		// load external settings
		log.WithFields(log.Fields{
			"settings": gitup.Settings,
		}).Debug("load external config")
		gitup.Config.Load(gitup.Settings)
	}
	return
}

func (gitup *GitUp) epilogue() {
	gitup.logger.Close()
}
