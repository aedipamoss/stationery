package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestStationery(t *testing.T) {
	if os.Getenv("BE_STATIONERY") == "1" {
		main()
		return
	}

	tmpProject, err := tmpProjectSetup(`
source: src
output: out
template: template.html
assets:
  css:
    - style.css`)
	if err != nil {
		t.Fatalf("unable to setup temporary working dir")
	}
	defer os.RemoveAll(tmpProject)

	err = mkdir(filepath.Join(tmpProject, "src"))
	if err != nil {
		t.Fatalf("unable to setup temp project src dir")
	}

	err = tmpPostSetup(filepath.Join(tmpProject, "src", "zomg.md"), `
---
title: zomg is a thing
---

# zomg
{{ .Timestamp "2018-03-24T12:43:03" }}

this is my temp post!`)
	if err != nil {
		t.Fatalf("unable to create temporary post")
	}

	err = execCommandWithProject(tmpProject)
	if err != nil {
		t.Fatalf("command finished with error %v", err)
	}

	page, err := readTmpPost(filepath.Join(tmpProject, "out", "zomg.html"))
	if err != nil {
		t.Fatalf("unable to read temporary post after parsing")
	}

	mustContain(t, page, "<h1>zomg</h1>")
	mustContain(t, page, "<title>zomg is a thing</title>")
	mustNotContain(t, page, "<h2>title: zomg is a thing</h2>")
}

func TestSingleFileSource(t *testing.T) {
	if os.Getenv("BE_STATIONERY") == "1" {
		main()
		return
	}

	tmpProject, err := tmpProjectSetup(`
source: src.md
output: out
template: template.html
assets:
  css:
    - style.css`)
	if err != nil {
		t.Fatalf("unable to setup temporary working dir")
	}
	defer os.RemoveAll(tmpProject)

	err = tmpPostSetup(filepath.Join(tmpProject, "src.md"), `
---
title: log of all zomg
---

# zomg all the things

this is my temp post!`)
	if err != nil {
		t.Fatalf("unable to create temporary post")
	}

	err = execCommandWithProject(tmpProject)
	if err != nil {
		t.Fatalf("command finished with error %v", err)
	}

	page, err := readTmpPost(filepath.Join(tmpProject, "out", "src.html"))
	if err != nil {
		t.Fatalf("unable to read temporary post after parsing")
	}

	mustContain(t, page, "<h1>zomg all the things</h1>")
	mustContain(t, page, "<title>log of all zomg</title>")
	mustNotContain(t, page, "<h2>title: log of all zomg</h2>")
}

func TestGenerateIndex(t *testing.T) {
	if os.Getenv("BE_STATIONERY") == "1" {
		main()
		return
	}

	tmpProject, err := tmpProjectSetup(`
source: src
output: out
template: template.html
assets:`)
	if err != nil {
		t.Fatalf("unable to setup temporary working dir")
	}
	defer os.RemoveAll(tmpProject)

	err = mkdir(filepath.Join(tmpProject, "src"))
	if err != nil {
		t.Fatalf("unable to setup temp project src dir")
	}

	err = tmpPostSetup(filepath.Join(tmpProject, "src", "zomg.md"), `
---
title: zomg is a thing
---

# zomg
{{ .Timestamp "2018-03-24T12:43:03" }}

this is my temp post!`)
	if err != nil {
		t.Fatalf("unable to create temporary post")
	}

	err = tmpPostSetup(filepath.Join(tmpProject, "src", "two.md"), `
---
title: my second post
---

# two
{{ .Timestamp "2018-03-24T12:43:03" }}

wow, so easy!`)
	if err != nil {
		t.Fatalf("unable to create temporary post")
	}

	err = tmpPostSetup(filepath.Join(tmpProject, "src", "three.md"), `
# three

look, i have no data!`)
	if err != nil {
		t.Fatalf("unable to create temporary post")
	}

	err = execCommandWithProject(tmpProject)
	if err != nil {
		t.Fatalf("command finished with error %v", err)
	}

	index, err := readTmpPost(filepath.Join(tmpProject, "out", "index.html"))
	if err != nil {
		t.Fatalf("unable to read temporary post after parsing")
	}

	mustContain(t, index, `<a href="zomg.html">zomg is a thing</a>`)
	mustContain(t, index, `<a href="three.html">three</a>`)
	mustContain(t, index, `
<html>
<head>`)
	mustContain(t, index, `<div id="index">`)
}

func TestTitles(t *testing.T) {
	if os.Getenv("BE_STATIONERY") == "1" {
		main()
		return
	}

	tmpProject, err := tmpProjectSetup(`
source: src
output: out
template: template.html
title: my blog
assets:`)
	if err != nil {
		t.Fatalf("unable to setup temporary working dir")
	}
	defer os.RemoveAll(tmpProject)

	err = mkdir(filepath.Join(tmpProject, "src"))
	if err != nil {
		t.Fatalf("unable to setup temp project src dir")
	}

	err = tmpPostSetup(filepath.Join(tmpProject, "src", "one.md"), `
---
title: first!
---

# one

this is one!`)
	if err != nil {
		t.Fatalf("unable to create temporary post")
	}

	err = tmpPostSetup(filepath.Join(tmpProject, "src", "two.md"), `
---
title: second!
---

# two

this is two!`)
	if err != nil {
		t.Fatalf("unable to create temporary post")
	}

	err = tmpPostSetup(filepath.Join(tmpProject, "src", "three.md"), `
# three

no title!`)
	if err != nil {
		t.Fatalf("unable to create temporary post")
	}

	err = execCommandWithProject(tmpProject)
	if err != nil {
		t.Fatalf("command finished with error %v", err)
	}

	index, err := readTmpPost(filepath.Join(tmpProject, "out", "index.html"))
	if err != nil {
		t.Fatalf("unable to read temporary post after parsing")
	}

	mustContain(t, index, `<a href="one.html">first!</a>`)
	mustContain(t, index, `<a href="three.html">three</a>`)
	mustContain(t, index, `
<html>
<head>
<title>my blog</title>`)

	page, err := readTmpPost(filepath.Join(tmpProject, "out", "three.html"))
	if err != nil {
		t.Fatalf("unable to read temporary post after parsing")
	}

	mustContain(t, page, `<title>three</title>`)
}

func mustContain(t *testing.T, page string, expected string) {
	if !strings.Contains(page, expected) {
		t.Errorf("content = %q, expected %s", page, expected)
	}
}

func mustNotContain(t *testing.T, page string, unexpected string) {
	if strings.Contains(page, unexpected) {
		t.Errorf("content = %q, unexpected %s", page, unexpected)
	}
}

func execCommandWithProject(tmpProject string) error {
	var out bytes.Buffer
	var stderr bytes.Buffer

	cmd := exec.Command(os.Args[0], "-test.run=TestStationery")
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	cmd.Dir = tmpProject
	cmd.Env = append(os.Environ(), "BE_STATIONERY=1")
	err := cmd.Run()
	if err != nil {
		fmt.Printf("%s: \n%s\n\n", err, stderr.String())
		return err
	}

	return nil
}

func readTmpPost(path string) (string, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return path, err
	}

	return string(content), err
}

func tmpProjectSetup(tmpConfig string) (string, error) {
	tmpProject, err := ioutil.TempDir("", "stationery")
	if err != nil {
		return tmpProject, err
	}

	tmpTemplate := `
<html>
<head>
<title>{{ .Title }}</title>
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
	err = mkdir(filepath.Join(tmpProject, "assets", "css"))
	if err != nil {
		return tmpProject, err
	}

	// write temp config to temp project dir
	if err = ioutil.WriteFile(filepath.Join(tmpProject, ".station.yml"), []byte(tmpConfig), 0666); err != nil {
		return tmpProject, err
	}

	// write temp template to temp project dir
	if err = ioutil.WriteFile(filepath.Join(tmpProject, "template.html"), []byte(tmpTemplate), 0666); err != nil {
		return tmpProject, err
	}

	// write temp css asset for temp project
	err = ioutil.WriteFile(filepath.Join(tmpProject, "assets", "css", "style.css"), []byte(tmpStyle), 0666)

	return tmpProject, err
}

func tmpPostSetup(tmpPath string, tmpPost string) error {
	return ioutil.WriteFile(tmpPath, []byte(tmpPost), 0666)
}

func mkdir(path string) error {
	return os.MkdirAll(path, 0777)
}
