package generate

import (
	"bufio"
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/aedipamoss/stationery/config"
	"github.com/aedipamoss/stationery/page"

	"github.com/gorilla/feeds"
)

var cfg config.Config

func sources(source string) (files []os.FileInfo, err error) {
	file, err := os.Stat(source)
	if err != nil {
		return nil, err
	}

	if !file.IsDir() {
		return []os.FileInfo{file}, nil
	}

	files, err = ioutil.ReadDir(source)
	return files, err
}

func load(files []os.FileInfo) (pages []*page.Page, err error) {
	for _, file := range files {
		page := &page.Page{}
		page.Assets = cfg.Assets
		page.FileInfo = file
		page.Template = cfg.Template

		err = page.Load(cfg.Source, cfg.Output)
		if err != nil {
			return pages, err
		}
		pages = append(pages, page)
	}

	return pages, nil
}

func generateHTML(pages []*page.Page) error {
	for _, page := range pages {
		err := page.Generate()
		if err != nil {
			return err
		}
		fmt.Println("Wrote: ", page.Destination)
	}

	return nil
}

func generateRSS(pages []*page.Page) error {
	feed := feeds.Feed{
		Title:       cfg.Title,
		Link:        &feeds.Link{Href: cfg.Link},
		Description: cfg.Description,
		Author:      &feeds.Author{Name: cfg.Name, Email: cfg.Email},
	}

	for _, page := range pages {
		feed.Add(&feeds.Item{
			Title:       page.Title(),
			Link:        &feeds.Link{Href: page.URL(cfg)},
			Description: page.Description(),
			Author:      &feeds.Author{Name: cfg.Name, Email: cfg.Email},
			Created:     page.Date(),
		})
	}
	rss, err := feed.ToRss()
	if err != nil {
		return err
	}

	dest := filepath.Join(cfg.Output, "index.rss")
	err = ioutil.WriteFile(dest, []byte(rss), 0644)

	fmt.Println("Wrote: ", dest)
	return err
}

// IndexTemplate is the text/template used for generating the index page.
var IndexTemplate = `
{{ define "index" }}
  <div id="index">
    <ul>
      {{ range . }}
        <li>{{ .Link }}</li>
      {{ end }}
    </ul>
  </div>
{{ end }}
`

func generateIndex(pages []*page.Page) error {
	index := &page.Page{}
	index.Destination = filepath.Join(cfg.Output, "index.html")
	index.Assets = cfg.Assets
	index.Template = cfg.Template

	var content bytes.Buffer
	buf := bufio.NewWriter(&content)

	tmpl, err := template.New("index").Parse(IndexTemplate)
	if err != nil {
		return err
	}

	err = tmpl.Execute(buf, pages)
	if err != nil {
		return err
	}
	err = buf.Flush()
	if err != nil {
		return err
	}

	// nolint: gosec
	index.Content = template.HTML(content.String())

	err = index.Generate()
	if err != nil {
		return err
	}

	fmt.Println("Wrote: ", index.Destination)
	return err
}

// Run is the main entrypoint to this program.
// It's caller is main() and logs any errors that occur during file generation.
func Run() {
	loaded, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	cfg = loaded

	err = os.MkdirAll(cfg.Output, 0700)
	if err != nil {
		log.Fatal(err)
	}

	if cfg.Assets != nil {
		err = cfg.Assets.Generate(cfg.Output)
		if err != nil {
			log.Fatal(err)
		}
	}

	files, err := sources(cfg.Source)
	if err != nil {
		log.Fatal(err)
	}

	pages, err := load(files)
	if err != nil {
		log.Fatal(err)
	}

	sort.Slice(pages[:], func(i, j int) bool {
		return pages[i].Date().After(pages[j].Date())
	})

	err = generateHTML(pages)
	if err != nil {
		log.Fatal(err)
	}

	err = generateRSS(pages)
	if err != nil {
		log.Fatal(err)
	}

	err = generateIndex(pages)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Done!")
}
