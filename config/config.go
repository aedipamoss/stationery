// Package config is for loading and facilitating project configuration
package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"

	"github.com/aedipamoss/stationery/assets"
)

// Config is structure containing the current blog's configuration
type Config struct {
	Assets  *assets.List
	Output  string
	Source  string
	Layouts []string
	SiteURL string `yaml:"site-url"`
	// RSS fields
	Title       string
	Description string
	Name        string
	Email       string
}

// ConfigFile is the default name for configuration file used by stationery.
const ConfigFile string = ".station.yml"

// Load will attempt to load the ConfigFile from disk and parse it.
func Load() (Config, error) {
	cfg := Config{}
	content, err := ioutil.ReadFile(ConfigFile)
	if err != nil {
		return cfg, err
	}

	err = cfg.parse(content)
	return cfg, err
}

func (cfg *Config) parse(content []byte) error {
	return yaml.Unmarshal(content, &cfg)
}
