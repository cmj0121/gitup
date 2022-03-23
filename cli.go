package gitup

import (
	"fmt"
	"os"
)

type version bool

func (ver version) BeforeApply() (err error) {
	fmt.Printf("%v (v%d.%d.%d)\n", PROJ_NAME, MAJOR, MINOR, MACRO)
	os.Exit(0)
	return
}

// the command-line interface of GitUp
type CLI struct {
	// show the version info
	Version version `short:"V" help:"Show version info"`

	// the log verbose level (error, warn, info, debug, trace)
	Verbose int `short:"v" type:"counter" help:"the log verbose level"`
}
