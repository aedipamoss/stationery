package main

import (
	"bufio"
	"bytes"
	"fmt"
	"html/template"
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

	for _, file := range files {
		src := config.Source + "/" + file.Name()
		base := filepath.Ext(src)
		name := file.Name()[0 : len(file.Name())-len(base)]
		path := config.Output + "/" + name + ".html"
		page := &Page{}

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

		page.Content = template.HTML(parsed[:])
		page.Title = findTitle(content)

		t, err := template.New("page").Parse(string(tmpl))
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
