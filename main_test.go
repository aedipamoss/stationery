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

	err = tmpPostSetup(filepath.Join(tmpProject, "src", "boom.wtf"), `
---
title: ":boom: goes the dynamite"
---

# boom
{{ .Timestamp "2018-08-13T23:20:49+09:00" }}

this file should be ignored!`)
	if err != nil {
		t.Fatalf("unable to create temporary post")
	}

	err = execCommandWithProject(tmpProject)
	if err != nil {
		t.Fatalf("command finished with error %v", err)
	}

	if _, err = os.Stat(filepath.Join(tmpProject, "out", "boom.html")); err == nil {
		t.Fatalf("file without .md extension was generated")
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
assets:`)
	if err != nil {
		t.Fatalf("unable to setup temporary working dir")
	}
	defer os.RemoveAll(tmpProject)

	tmpOut := filepath.Join(tmpProject, "out")

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

	mustContain(t, index, fmt.Sprintf(`<a href="%s/zomg.html">zomg is a thing</a>`, tmpOut))
	mustContain(t, index, fmt.Sprintf(`<a href="%s/three.html">three</a>`, tmpOut))
	mustContain(t, index, `
<html>
<head>`)
	mustContain(t, index, `<div id="index">`)
}

func TestGenerateTags(t *testing.T) {
	if os.Getenv("BE_STATIONERY") == "1" {
		main()
		return
	}

	tmpProject, err := tmpProjectSetup(`
source: src
output: out
assets:
  css:
    - style.css`)
	if err != nil {
		t.Fatalf("unable to setup temporary working dir")
	}
	defer os.RemoveAll(tmpProject)

	tmpOut := filepath.Join(tmpProject, "out")

	err = mkdir(filepath.Join(tmpProject, "src"))
	if err != nil {
		t.Fatalf("unable to setup temp project src dir")
	}

	err = tmpPostSetup(filepath.Join(tmpProject, "src", "zomg.md"), `
---
title: zomg is a thing
tags:
  - foo
  - bar
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
tags:
  - bar
---

# two
{{ .Timestamp "2018-03-24T12:43:03" }}

wow, so easy!`)
	if err != nil {
		t.Fatalf("unable to create temporary post")
	}

	err = execCommandWithProject(tmpProject)
	if err != nil {
		t.Fatalf("command finished with error %v", err)
	}

	index, err := readTmpPost(filepath.Join(tmpProject, "out", "tag", "bar.html"))
	if err != nil {
		t.Fatalf("unable to read temporary post after parsing")
	}

	mustContain(t, index, fmt.Sprintf(`<a href="%s/zomg.html">zomg is a thing</a>`, tmpOut))
	mustContain(t, index, fmt.Sprintf(`<a href="%s/two.html">my second post</a>`, tmpOut))
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
title: my blog
assets:`)
	if err != nil {
		t.Fatalf("unable to setup temporary working dir")
	}
	defer os.RemoveAll(tmpProject)

	tmpOut := filepath.Join(tmpProject, "out")

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

	mustContain(t, index, fmt.Sprintf(`<a href="%s/one.html">first!</a>`, tmpOut))
	mustContain(t, index, fmt.Sprintf(`<a href="%s/three.html">three</a>`, tmpOut))
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

func TestConfigIndexMetaData(t *testing.T) {
	if os.Getenv("BE_STATIONERY") == "1" {
		main()
		return
	}

	tmpProject, err := tmpProjectSetup(`
source: src
output: out
title: my blog
description: my default description
twitter: aedipamoss
image: images/avatar.jpg`)
	if err != nil {
		t.Fatalf("unable to setup temporary working dir")
	}
	defer os.RemoveAll(tmpProject)

	tmpOut := filepath.Join(tmpProject, "out")

	err = mkdir(filepath.Join(tmpProject, "src"))
	if err != nil {
		t.Fatalf("unable to setup temp project src dir")
	}

	err = tmpPostSetup(filepath.Join(tmpProject, "src", "config.md"), `
---
title: config inherited defaults!
tags:
  - mytag
---

this is config!`)
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

	mustContain(t, index, fmt.Sprintf(`<meta name="description" content="%s" />`, "my default description"))
	mustContain(t, index, fmt.Sprintf(`<meta property="og:description" content="%s" />`, "my default description"))
	mustContain(t, index, fmt.Sprintf(`<meta property="og:title" content="%s" />`, "my blog"))
	mustContain(t, index, fmt.Sprintf(`<meta property="og:image" content="%s" />`, filepath.Join(tmpOut, "images", "avatar.jpg")))
	mustContain(t, index, `<meta name="twitter:card" content="summary" />`)
	mustContain(t, index, `<meta name="twitter:creator" content="@aedipamoss" />`)
	mustContain(t, index, `<meta name="twitter:site" content="@aedipamoss" />`)

	tag, err := readTmpPost(filepath.Join(tmpProject, "out", "tag", "mytag.html"))
	if err != nil {
		t.Fatalf("unable to read temporary post after parsing")
	}

	mustContain(t, tag, fmt.Sprintf(`<meta property="og:title" content="%s" />`, "my blog"))
	mustContain(t, tag, fmt.Sprintf(`<meta property="og:image" content="%s" />`, filepath.Join(tmpOut, "images", "avatar.jpg")))
	mustContain(t, tag, `<meta name="twitter:card" content="summary" />`)
	mustContain(t, tag, `<meta name="twitter:creator" content="@aedipamoss" />`)
	mustContain(t, tag, `<meta name="twitter:site" content="@aedipamoss" />`)
}

func TestPageMetaData(t *testing.T) {
	if os.Getenv("BE_STATIONERY") == "1" {
		main()
		return
	}

	tmpProject, err := tmpProjectSetup(`
source: src
output: out
title: my blog
description: my default description
twitter: aedipamoss
image: images/avatar.jpg`)
	if err != nil {
		t.Fatalf("unable to setup temporary working dir")
	}
	defer os.RemoveAll(tmpProject)

	tmpOut := filepath.Join(tmpProject, "out")

	err = mkdir(filepath.Join(tmpProject, "src"))
	if err != nil {
		t.Fatalf("unable to setup temp project src dir")
	}

	err = tmpPostSetup(filepath.Join(tmpProject, "src", "config.md"), `
---
title: config inherited defaults!
tags:
  - mytag
---

this is config!`)
	if err != nil {
		t.Fatalf("unable to create temporary post")
	}

	err = execCommandWithProject(tmpProject)
	if err != nil {
		t.Fatalf("command finished with error %v", err)
	}

	config, err := readTmpPost(filepath.Join(tmpProject, "out", "config.html"))
	if err != nil {
		t.Fatalf("unable to read temporary post after parsing")
	}

	mustContain(t, config, fmt.Sprintf(`<meta name="description" content="%s" />`, "my default description"))
	mustContain(t, config, fmt.Sprintf(`<meta property="og:description" content="%s" />`, "my default description"))
	mustContain(t, config, fmt.Sprintf(`<meta property="og:title" content="%s" />`, "config inherited defaults!"))
	mustContain(t, config, fmt.Sprintf(`<meta property="og:image" content="%s" />`, filepath.Join(tmpOut, "images", "avatar.jpg")))
	mustContain(t, config, `<meta name="twitter:card" content="summary" />`)
	mustContain(t, config, `<meta name="twitter:creator" content="@aedipamoss" />`)
	mustContain(t, config, `<meta name="twitter:site" content="@aedipamoss" />`)
}

func TestOverideMetaData(t *testing.T) {
	if os.Getenv("BE_STATIONERY") == "1" {
		main()
		return
	}

	tmpProject, err := tmpProjectSetup(`
source: src
output: out
title: my blog
description: my default description
twitter: aedipamoss
image: images/avatar.jpg`)
	if err != nil {
		t.Fatalf("unable to setup temporary working dir")
	}
	defer os.RemoveAll(tmpProject)

	tmpOut := filepath.Join(tmpProject, "out")

	err = mkdir(filepath.Join(tmpProject, "src"))
	if err != nil {
		t.Fatalf("unable to setup temp project src dir")
	}

	err = tmpPostSetup(filepath.Join(tmpProject, "src", "overide.md"), `
---
title: config overridden!
twitter: forgetme
image: images/zomg.jpg
description: description overridden!
---

this is overridden!`)
	if err != nil {
		t.Fatalf("unable to create temporary post")
	}

	err = execCommandWithProject(tmpProject)
	if err != nil {
		t.Fatalf("command finished with error %v", err)
	}

	overridden, err := readTmpPost(filepath.Join(tmpProject, "out", "overide.html"))
	if err != nil {
		t.Fatalf("unable to read temporary post after parsing")
	}

	mustContain(t, overridden, fmt.Sprintf(`<meta name="description" content="%s" />`, "description overridden!"))
	mustContain(t, overridden, fmt.Sprintf(`<meta property="og:description" content="%s" />`, "description overridden!"))
	mustContain(t, overridden, fmt.Sprintf(`<meta property="og:title" content="%s" />`, "config overridden!"))
	mustContain(t, overridden, fmt.Sprintf(`<meta property="og:image" content="%s" />`, filepath.Join(tmpOut, "images", "zomg.jpg")))
	mustContain(t, overridden, `<meta name="twitter:card" content="summary" />`)
	mustContain(t, overridden, `<meta name="twitter:creator" content="@forgetme" />`)
	mustContain(t, overridden, `<meta name="twitter:site" content="@forgetme" />`)
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

	tmpPageTemplate := `
<html>
<head>
{{ .Headers }}
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

	tmpIndexTemplate := `
<html>
<head>
{{ .Headers }}
</head>
<body>
  <div id="index">
    {{ .Index }}
  </div>
</body>
</html>
`

	// setup temp assets dir
	err = mkdir(filepath.Join(tmpProject, "assets", "css"))
	if err != nil {
		return tmpProject, err
	}

	// write temp config to temp project dir
	if err = ioutil.WriteFile(filepath.Join(tmpProject, ".station.yml"), []byte(tmpConfig), 0666); err != nil {
		return tmpProject, err
	}

	// setup temp layouts dir
	err = mkdir(filepath.Join(tmpProject, "layouts"))
	if err != nil {
		return tmpProject, err
	}

	// write temp page template to temp project dir
	if err = ioutil.WriteFile(filepath.Join(tmpProject, "layouts", "page.html"), []byte(tmpPageTemplate), 0666); err != nil {
		return tmpProject, err
	}

	// write temp index template to temp project dir
	if err = ioutil.WriteFile(filepath.Join(tmpProject, "layouts", "index.html"), []byte(tmpIndexTemplate), 0666); err != nil {
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
