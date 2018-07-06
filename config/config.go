package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// Config is structure containing the current blog's configuration
type Config struct {
	Source   string
	Output   string
	Template string
	Assets   struct {
		CSS    []string
		JS     []string
		Images []string
	}
}

// ConfigFile is the default name for configuration file used by stationery.
const ConfigFile string = ".station.yml"

func read(ConfigFile string) (content []byte, err error) {
	content, err = ioutil.ReadFile(ConfigFile)
	return
}

// Load will attempt to load the ConfigFile from disk and parse it.
func Load() (config Config, error error) {
	content, error := read(ConfigFile)
	if error != nil {
		return config, error
	}

	config, error = parse(Config{}, content)

	return config, error
}

func parse(config Config, content []byte) (parsed Config, err error) {
	err = yaml.Unmarshal(content, &config)
	return config, err
}
