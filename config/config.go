package config

import (
	"io/ioutil"

	"github.com/go-yaml/yaml"
)

type Config struct {
	Source   string
	Output   string
	Template string
	Assets   struct {
		Css []string
		Js  []string
	}
}

const ConfigFile string = ".station.yml"

func read(ConfigFile string) (content []byte, err error) {
	content, err = ioutil.ReadFile(ConfigFile)
	return
}

func Load() (config Config, error error) {
	content, error := read(ConfigFile)
	if error != nil {
		return config, error
	}

	config, error = parse(Config{}, content)
	if error != nil {
		return config, error
	}

	return config, nil
}

func parse(config Config, content []byte) (parsed Config, err error) {
	err = yaml.Unmarshal(content, &config)
	if err != nil {
		return config, err
	}

	return config, nil
}
