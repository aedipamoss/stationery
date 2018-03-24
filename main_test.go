package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestTimestamp(t *testing.T) {
	page := Page{}
	stamp := "2018-03-22"
	expected := "[@ 2018-03-22](#2018-03-22)"
	if expected != page.Timestamp(stamp) {
		t.Errorf("expected %v, got %v", expected, page.Timestamp(stamp))
	}
}

func TestStationery(t *testing.T) {
	if os.Getenv("BE_STATIONERY") == "1" {
		Stationery()
		return
	}

	tmpConfig := `
source: src
output: out
template: template.html
assets:
  css:
    - style.css`
	tmpProject, err := ioutil.TempDir("", "stationery")
	if err != nil {
		t.Fatalf("unable to setup temporary working dir")
	}
	defer os.RemoveAll(tmpProject)

	tmpTemplate := `
<html>
<head>
<title>{{ .Data.Title }}</title>
{{ template "assets" . }}
</head>
<body>
  {{ .Content }}
</body>
</html>
`
	tmpStyle := `
html {
    background: #3c3c3c;
    color: #65cdad;
    font-family: mono;
}`

	// setup temp assets dir
	err = os.Mkdir(filepath.Join(tmpProject, "assets"), 0777)
	if err != nil {
		t.Fatalf("unable to setup temp project assets dir")
	}
	err = os.Mkdir(filepath.Join(tmpProject, "assets", "css"), 0777)
	if err != nil {
		t.Fatalf("unable to setup temp project assets css dir")
	}

	// write temp config to temp project dir
	if err = ioutil.WriteFile(filepath.Join(tmpProject, ".station.yml"), []byte(tmpConfig), 0666); err != nil {
		t.Fatalf("unable to setup temporary project config")
	}

	// write temp template to temp project dir
	if err = ioutil.WriteFile(filepath.Join(tmpProject, "template.html"), []byte(tmpTemplate), 0666); err != nil {
		t.Fatalf("unable to setup temporary project template")
	}

	// write temp css asset for temp project
	if err = ioutil.WriteFile(filepath.Join(tmpProject, "assets", "css", "style.css"), []byte(tmpStyle), 0666); err != nil {
		t.Fatalf("unable to setup temporary project css assets")
	}

	err = os.Mkdir(filepath.Join(tmpProject, "src"), 0777)
	if err != nil {
		t.Fatalf("unable to setup temp project src dir")
	}

	tmpPost := `
---
title: zomg is a thing
---

# zomg
{{ .Timestamp "2018-03-24T12:43:03" }}

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
	page := string(content)

	if !strings.Contains(page, "<h1>zomg</h1>") {
		t.Errorf("content = %q, wanted <h1>zomg</h1>", page)
	}

	if !strings.Contains(string(page), "<title>zomg is a thing</title>") {
		t.Errorf("expected content to have title: %q", page)
	}

	if strings.Contains(string(page), "<h2>title: zomg is a thing</h2>") {
		t.Errorf("meta-data is bleeding into content body", page)
	}
}

func TestSingleFileSource(t *testing.T) {
	if os.Getenv("BE_STATIONERY") == "1" {
		Stationery()
		return
	}

	tmpConfig := `
source: src.md
output: out
template: template.html
assets:
  css:
    - style.css`
	tmpProject, err := ioutil.TempDir("", "stationery")
	if err != nil {
		t.Fatalf("unable to setup temporary working dir")
	}
	defer os.RemoveAll(tmpProject)

	tmpTemplate := `
<html>
<head>
<title>{{ .Data.Title }}</title>
{{ template "assets" . }}
</head>
<body>
  {{ .Content }}
</body>
</html>
`
	tmpStyle := `
html {
    background: #3c3c3c;
    color: #65cdad;
    font-family: mono;
}`

	// setup temp assets dir
	err = os.Mkdir(filepath.Join(tmpProject, "assets"), 0777)
	if err != nil {
		t.Fatalf("unable to setup temp project assets dir")
	}
	err = os.Mkdir(filepath.Join(tmpProject, "assets", "css"), 0777)
	if err != nil {
		t.Fatalf("unable to setup temp project assets css dir")
	}

	// write temp config to temp project dir
	if err = ioutil.WriteFile(filepath.Join(tmpProject, ".station.yml"), []byte(tmpConfig), 0666); err != nil {
		t.Fatalf("unable to setup temporary project config")
	}

	// write temp template to temp project dir
	if err = ioutil.WriteFile(filepath.Join(tmpProject, "template.html"), []byte(tmpTemplate), 0666); err != nil {
		t.Fatalf("unable to setup temporary project template")
	}

	// write temp css asset for temp project
	if err = ioutil.WriteFile(filepath.Join(tmpProject, "assets", "css", "style.css"), []byte(tmpStyle), 0666); err != nil {
		t.Fatalf("unable to setup temporary project css assets")
	}

	tmpPost := `
---
title: log of all zomg
---

# zomg all the things

this is my temp post!`

	if err = ioutil.WriteFile(filepath.Join(tmpProject, "src.md"), []byte(tmpPost), 0666); err != nil {
		t.Fatalf("unable to create temporary post for testing")
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestStationery")
	cmd.Dir = tmpProject
	cmd.Env = append(os.Environ(), "BE_STATIONERY=1")
	err = cmd.Run()
	if err != nil {
		t.Fatalf("command finished with error %v", err)
	}

	content, err := ioutil.ReadFile(filepath.Join(tmpProject, "out", "src.html"))
	if err != nil {
		t.Fatalf("unable to read temporary post after parsing")
	}
	page := string(content)

	if !strings.Contains(page, "<h1>zomg all the things</h1>") {
		t.Errorf("content = %q, wanted <h1>zomg all the things</h1>", page)
	}

	if !strings.Contains(string(page), "<title>log of all zomg</title>") {
		t.Errorf("expected content to have title: %q", page)
	}

	if strings.Contains(string(page), "<h2>title: log of all zomg</h2>") {
		t.Errorf("meta-data is bleeding into content body", page)
	}
}
