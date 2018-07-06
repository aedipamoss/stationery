package config

import "testing"

func TestConfig(t *testing.T) {
	config := &Config{
		Source: "src",
		Output: "out",
	}

	if config.Source != "src" {
		t.Error("no source dir was specified")
	}
}

func TestConfigFile(t *testing.T) {
	if ConfigFile != ".station.yml" {
		t.Error("invalid config file name")
	}
}

func TestAssets(t *testing.T) {
	data := `
source: path/to/src
output: path/to/out
assets:
  css:
    - site.css
`
	config := Config{}
	parsed, err := parse(config, []byte(data))
	if err != nil {
		t.Error(err)
	}
	if parsed.Assets.CSS == nil {
		t.Error("unable to parse CSS assets")
	}
}

func TestParse(t *testing.T) {
	data := `
source: path/to/source
output: path/to/output
`

	config := Config{}
	parsed, err := parse(config, []byte(data))
	if err != nil {
		t.Error(err)
	}
	if parsed.Source != "path/to/source" {
		t.Error("no source dir was specified")
	}
}
