package generate

import (
	"bufio"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/aedipamoss/stationery/config"
	"github.com/aedipamoss/stationery/page"
)

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

func load(files []os.FileInfo, cfg config.Config) (pages []*page.Page, err error) {
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
		fmt.Println("Wrote: ", page.FileInfo.Name())
	}

	return nil
}

var IndexTemplate = `
{{ define "index" }}
  {{ range . }}
    {{ .Data.Title }}
  {{ end }}
{{ end }}
`

func generateIndex(pages []*page.Page, cfg config.Config) error {
	tmpl, err := template.New("index").Parse(IndexTemplate)
	if err != nil {
		return err
	}

	dest := filepath.Join(cfg.Output, "index.html")
	f, err := os.Create(dest)
	if err != nil {
		return err
	}

	buf := bufio.NewWriter(f)

	err = tmpl.Execute(buf, pages)
	if err != nil {
		return err
	}
	err = buf.Flush()
	return err
}

// Run is the main entrypoint to this program.
// It's caller is main() and logs any errors that occur during file generation.
func Run() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	err = os.MkdirAll(cfg.Output, 0700)
	if err != nil {
		log.Fatal(err)
	}

	err = cfg.Assets.Generate(cfg.Output)
	if err != nil {
		log.Fatal(err)
	}

	files, err := sources(cfg.Source)
	if err != nil {
		log.Fatal(err)
	}

	pages, err := load(files, cfg)
	if err != nil {
		log.Fatal(err)
	}

	err = generateHTML(pages)
	if err != nil {
		log.Fatal(err)
	}

	err = generateIndex(pages, cfg)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Done!")
}
