package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestStationery(t *testing.T) {
	if os.Getenv("BE_STATIONERY") == "1" {
		Stationery()
		return
	}

	tmpConfig := `
source: src
output: out`
	tmpProject, err := ioutil.TempDir("", "stationery")

	if err != nil {
		t.Fatalf("unable to setup temporary working dir")
	}
	defer os.RemoveAll(tmpProject)

	// write temp config to temp project dir
	if err = ioutil.WriteFile(filepath.Join(tmpProject, ".station.yml"), []byte(tmpConfig), 0666); err != nil {
		t.Fatalf("unable to setup temporary project config")
	}

	if _, err = os.Stat(filepath.Join(tmpProject, "src")); err != nil && os.IsNotExist(err) {
		os.Mkdir(filepath.Join(tmpProject, "src"), 0777)
	}

	tmpPost := `
# zomg

this is my temp post!`

	if err = ioutil.WriteFile(filepath.Join(tmpProject, "src", "zomg.md"), []byte(tmpPost), 0666); err != nil {
		t.Fatalf("unable to create temporary post for testing")
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestStationery")
	cmd.Dir = tmpProject
	cmd.Env = append(os.Environ(), "BE_STATIONERY=1")
	err = cmd.Run()
	if err != nil {
		t.Fatalf("command finished with error %v", err)
	}

	content, err := ioutil.ReadFile(filepath.Join(tmpProject, "out", "zomg.html"))
	if err != nil {
		t.Fatalf("unable to read temporary post after parsing")
	}

	if !strings.Contains(string(content), "<h1>zomg</h1>") {
		t.Errorf("content = %q, wanted <h1>zomg</h1>")
	}
}
