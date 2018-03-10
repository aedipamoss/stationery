package config

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/go-yaml/yaml"
)

type Config struct {
	Source   string
	Output   string
	Template string
}

const ConfigFile string = ".station.yml"

func read(ConfigFile string) (content []byte, err error) {
	content, err = ioutil.ReadFile(ConfigFile)
	return
}

func Load() {
	content, err := read(ConfigFile)
	if err != nil {
		log.Fatal("Unable to read config")
	}

	config, err := parse(Config{}, content)
	if err != nil {
		log.Fatal("Unable to parse config")
	}
	fmt.Printf("Config is %v", config)
	return
}

func parse(config Config, content []byte) (parsed Config, err error) {
	err = yaml.Unmarshal(content, &config)
	if err != nil {
		return config, err
	}

	return config, nil
}
