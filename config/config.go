package config

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"

	log "github.com/sirupsen/logrus"
)

var ConfigPath = []string{
	".gitup.yml",
	".gitup.yaml",
}

type Config struct {
	// the storage folder of the blogs
	Workdir []string

	// the general project meta
	Project string `yaml:",omitempty"`
	Author  string `yaml:",omitempty"`

	Render
	Settings
}

// load the external config/configs and override the exists settings
func (config *Config) Load(path string) {
	info, err := os.Stat(path)
	switch {
	case err == nil && info.IsDir():
		log.WithFields(log.Fields{
			"path": path,
		}).Info("load config from folder")

		for _, config_path := range ConfigPath {
			config_path = fmt.Sprintf("%v/%v", path, config_path)
			config_path = filepath.Clean(config_path)

			if config_path[:len(path)] != path {
				log.WithFields(log.Fields{
					"path":        path,
					"config_path": config_path,
				}).Info("invalid config path")
				continue
			}

			config.Load(config_path)
		}
	case err == nil:
		switch file, err := os.Open(path); err {
		case nil:
			defer file.Close()

			log.WithFields(log.Fields{
				"path": path,
			}).Debug("load config from file")

			config.LoadFromReader(file)
		default:
			log.WithFields(log.Fields{
				"path":  path,
				"error": err,
			}).Warn("cannot open config")
		}
	default:
		log.WithFields(log.Fields{
			"path":  path,
			"error": err,
		}).Trace("cannot load config")
	}
}

// load the external config from io.Reader
func (config *Config) LoadFromReader(reader io.Reader) {
	var buff bytes.Buffer

	if _, err := io.Copy(&buff, reader); err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Warn("cannot copy text from reader")
		return
	}

	if err := yaml.Unmarshal(buff.Bytes(), &config); err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Warn("cannot read config as YAML")
		return
	}
}

// show the config as YAML format
func (config Config) String() (conf string) {
	if text, err := yaml.Marshal(config); err == nil {
		conf = string(text)
	}
	return
}
