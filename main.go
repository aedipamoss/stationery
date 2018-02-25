package main

import (
	"bufio"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/aedipamoss/stationery/config"
	blackfriday "gopkg.in/russross/blackfriday.v2"
)

type Page struct {
	Content template.HTML
}

const Template = `
{{ define "Page" }}
<html>
<head>
  <meta charset="utf-8" />
  <title>a page</title>
</head>
<body>
  {{ .Content }}
</body>
</html>
{{ end }}
`

func main() {
	config := config.Config{
		Source: "src",
		Output: "out",
	}
	files, err := ioutil.ReadDir(config.Source)
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

		t, err := template.New("page").Parse(Template)
		err = t.ExecuteTemplate(w, "Page", page)

		//		err = ioutil.WriteFile(path, out, 0644)
		if err != nil {
			log.Fatal(err)
		}
		w.Flush()

		fmt.Println("Wrote: ", path)
		//fmt.Printf("File contents: %s\n", content)
	}

	fmt.Println("Done!")
}
