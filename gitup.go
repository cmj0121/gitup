package gitup

import (
	"bytes"
	"io"
	"os"

	"github.com/alecthomas/kong"
	"github.com/cmj0121/gitup/config"
	"gopkg.in/yaml.v2"

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

	err = gitup.loadSettings()
	return
}

func (gitup *GitUp) loadSettings() (err error) {
	var reader io.ReadCloser

	switch gitup.Settings {
	case "":
		// need not load the config
		return
	default:
		log.WithFields(log.Fields{
			"settings": gitup.Settings,
		}).Trace("load external config settings")

		if reader, err = os.Open(gitup.Settings); err != nil {
			log.WithFields(log.Fields{
				"settings": gitup.Settings,
				"error":    err,
			}).Warn("cannot open config settings")
			return
		}
	}

	defer reader.Close()
	var buff bytes.Buffer

	if _, err = io.Copy(&buff, reader); err != nil {
		log.WithFields(log.Fields{
			"settings": gitup.Settings,
			"error":    err,
		}).Warn("cannot copy text from config")
		return
	}

	gitup.Config = &config.Config{}
	if err = yaml.Unmarshal(buff.Bytes(), &gitup.Config); err != nil {
		log.WithFields(log.Fields{
			"settings": gitup.Settings,
			"error":    err,
		}).Warn("cannot read config as YAML")
		return
	}

	return
}

func (gitup *GitUp) epilogue() {
	gitup.logger.Close()
}
