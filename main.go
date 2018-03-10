package main

import (
	"bufio"
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"

	"github.com/aedipamoss/stationery/config"
	blackfriday "gopkg.in/russross/blackfriday.v2"
)

type Page struct {
	Content template.HTML
	Title   string
	Config  config.Config
}

func findTitle(content []byte) (title string) {
	reader := bytes.NewReader(content)
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		text := scanner.Text()
		matched, _ := regexp.MatchString("^# [[:alpha:]]+", text)
		if matched {
			return string(text[2:])
		}
	}
	log.Fatal("no title found")
	return ""
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
		src := config.Source + "/" + file.Name()
		base := filepath.Ext(src)
		name := file.Name()[0 : len(file.Name())-len(base)]
		path := config.Output + "/" + name + ".html"
		page := &Page{}
		page.Config = config // pass config to page struct

		content, err := ioutil.ReadFile(src)
		if err != nil {
			log.Fatal(err)
		}

		f, err := os.Create(path)
		if err != nil {
			log.Fatal(err)
		}
		w := bufio.NewWriter(f)

		parsed := blackfriday.Run(content)

		page.Content = template.HTML(parsed[:]) // TODO: add err checks
		page.Title = findTitle(content)         // here too

		t, err := template.New("page").Parse(string(tmpl))
		if err != nil {
			log.Fatal(err)
		}
		_, err = t.Parse(AssetsTemplate)
		if err != nil {
			log.Fatal(err)
		}

		err = t.Execute(w, page)

		if err != nil {
			log.Fatal(err)
		}
		w.Flush()

		fmt.Println("Wrote: ", path)
	}

	fmt.Println("Done!")
}

func main() {
	Stationery()
	os.Exit(0)
}
