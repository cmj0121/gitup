package gitup

import (
	"fmt"
)

const (
	// the project name
	PROJ_NAME = "gitup"
	AUTHOR    = "cmj <cmj@cmj.tw>"
	// the version meta
	MAJOR = 0
	MINOR = 3
	MACRO = 4
)

// the helper function to get the version
func Version() (ver string) {
	ver = fmt.Sprintf("%v (v%d.%d.%d)\n", PROJ_NAME, MAJOR, MINOR, MACRO)
	return
}
