package main

import (
	"bufio"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"

	"github.com/aedipamoss/stationery/config"
	"github.com/go-yaml/yaml"
	blackfriday "gopkg.in/russross/blackfriday.v2"
)

type Page struct {
	Content    template.HTML
	Config     config.Config
	Data       Data
	Raw        []byte
	SourceFile os.FileInfo
}

type Data struct {
	Title string
}

func (page *Page) source() string {
	return page.Config.Source + "/" + page.SourceFile.Name()
}

func (page *Page) basename() string {
	return filepath.Ext(page.SourceFile.Name())
}

func (page *Page) filename() string {
	return page.SourceFile.Name()[0 : len(page.SourceFile.Name())-len(page.basename())]
}

func (page *Page) outfile() string {
	return page.Config.Output + "/" + page.filename() + ".html"
}

func (page *Page) load() (ok bool, err error) {
	content, err := ioutil.ReadFile(page.source())
	if err != nil {
		return false, err
	}
	data, err := parseFrontMatter(content)
	if err != nil {
		return false, err
	}
	r := regexp.MustCompile(FrontMatterRegex)
	raw := r.ReplaceAllString(string(content), "")
	page.Data = data
	page.Raw = []byte(raw)

	return true, nil
}

const FrontMatterRegex = `(?s)(---\s*\n.*?\n?)(---\s*\n?)`

func parseFrontMatter(content []byte) (data Data, err error) {
	r, err := regexp.Compile(FrontMatterRegex)
	if err != nil {
		return data, err
	}

	matches := r.FindAllStringSubmatch(string(content), -1)

	if len(matches) > 0 && len(matches[0]) > 0 {
		err = yaml.Unmarshal([]byte(matches[0][1]), &data)
		if err != nil {
			return data, err
		}
	}
	return data, nil
}

func (page Page) Generate(tmpl []byte) (ok bool, err error) {
	_, err = page.load()
	if err != nil {
		return false, err
	}

	f, err := os.Create(page.outfile())
	if err != nil {
		return false, err
	}
	w := bufio.NewWriter(f)

	parsed := blackfriday.Run(page.Raw)

	page.Content = template.HTML(parsed[:]) // TODO: add err checks

	t, err := template.New("page").Parse(string(tmpl))
	if err != nil {
		return false, err
	}
	_, err = t.Parse(AssetsTemplate)
	if err != nil {
		return false, err
	}

	err = t.Execute(w, page)

	if err != nil {
		return false, err
	}
	w.Flush()
	return true, nil
}

// assets expects a struct with access to the Config struct
// that includes "Assets.Css" fields with an array of stylesheet names
const AssetsTemplate = `
{{ define "assets" }}
  {{ range .Config.Assets.Css }}
    <link type="text/css" rel="stylesheet" href="css/{{ . }}">
  {{ end }}
{{ end }}
`

func generateAssets(config config.Config) (ok bool, error error) {
	// generate css
	cssDir := config.Output + "/css"

	_, err := os.Stat(cssDir)
	if err != nil && os.IsNotExist(err) {
		fmt.Printf("Making assets css output dir: %v\n", cssDir)
		e := os.Mkdir(cssDir, 0777)
		if e != nil {
			return false, e
		}
	}

	for _, file := range config.Assets.Css {
		path := cssDir + "/" + file
		src := "assets/css/" + file

		from, err := os.Open(src)
		if err != nil {
			return false, err
		}
		defer from.Close()

		to, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)
		if err != nil {
			return false, err
		}
		defer to.Close()

		_, err = io.Copy(to, from)
		if err != nil {
			return false, err
		}
	}
	return true, nil
}

func Stationery() {
	config, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	files, err := ioutil.ReadDir(config.Source)
	if err != nil {
		log.Fatal(err)
	}

	tmpl, err := ioutil.ReadFile(config.Template)
	if err != nil {
		log.Fatal(err)
	}

	_, err = os.Stat(config.Output)
	if err != nil && os.IsNotExist(err) {
		fmt.Printf("Making output dir: %v\n", config.Output)
		os.Mkdir(config.Output, 0777)
	}

	_, err = generateAssets(config)
	if err != nil {
		log.Fatal("Problem generating assets")
	}

	for _, file := range files {
		page := &Page{}
		page.Config = config
		page.SourceFile = file

		_, err := page.Generate(tmpl)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Wrote: ", page)

	}

	fmt.Println("Done!")
}

func main() {
	Stationery()
	os.Exit(0)
}
