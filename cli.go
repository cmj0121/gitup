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
	Version version `short:"v" help:"Show version info"`
}
