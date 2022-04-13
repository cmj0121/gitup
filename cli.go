package gitup

import (
	"fmt"
	"os"

	"github.com/cmj0121/gitup/blog"
	"github.com/cmj0121/gitup/clone"
	"github.com/cmj0121/gitup/config"
	"gopkg.in/yaml.v2"
)

type version bool

func (ver version) BeforeApply() (err error) {
	fmt.Printf("%v (v%d.%d.%d)\n", PROJ_NAME, MAJOR, MINOR, MACRO)
	os.Exit(0)
	return
}

// the dummy struct to dump the config
type conf_render struct{}

// show the config as YAML format
func (*conf_render) Run(conf *config.Config) (err error) {
	var text []byte

	if text, err = yaml.Marshal(conf); err == nil {
		fmt.Println(string(text))
		return
	}

	return
}

// the command-line interface of GitUp
type CLI struct {
	// show the version info
	Version version `short:"V" help:"Show version info"`

	// the log verbose level (error, warn, info, debug, trace)
	Verbose int `short:"v" type:"counter" help:"the log verbose level"`

	// the log file
	LogFile string `type:"path" name:"log-file" help:"the log file destination (default: STDERR)"`

	// the sub-command and config settings
	Settings string       `short:"s" name:"setting" type:"file" help:"the global settings of the gitup"`
	Blog     *blog.Blog   `cmd:"" help:"generate the HTML by single blog/markdown"`
	Clone    *clone.Clone `cmd:"" help:"clone the repository and generate HTML webpages"`
	Config   *conf_render `name:"config" cmd:"" help:"dump the config settings"`
}
