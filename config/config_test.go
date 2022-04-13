package config

import (
	"strings"
	"testing"
)

var test_config = `
---
`

func TestLoadConfig(t *testing.T) {
	conf := Config{}

	conf.Load(".")
	conf.Load("DOES_NOT_EXISTS")
}

func TestLoadFromReader(t *testing.T) {
	conf := Config{}

	reader := strings.NewReader(test_config)
	conf.LoadFromReader(reader)
}
