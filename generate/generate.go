package generate

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/aedipamoss/stationery/config"
	"github.com/aedipamoss/stationery/page"

	"github.com/gorilla/feeds"
)

var cfg config.Config

func rootURI() string {
	if cfg.SiteURL != "" {
		return strings.TrimRight(cfg.SiteURL, "/") + "/"
	} else {
		path, err := filepath.Abs(cfg.Output)
		if err != nil {
			panic(err)
		}
		return strings.TrimRight(path, "/") + "/"
	}
}

// Returns a list of pages sorted by date
func load(source string) (pages []*page.Page, err error) {
	var files []os.FileInfo
	file, err := os.Stat(source)
	if err != nil {
		return nil, err
	}

	if !file.IsDir() {
		files = []os.FileInfo{file}
	} else {
		files, err = ioutil.ReadDir(source)
	}
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) != ".md" {
			continue
		}
		page := &page.Page{}
		page.Assets = cfg.Assets
		page.FileInfo = file
		page.Root = rootURI()
		page.Template = filepath.Join("layouts", "page.html")

		err := page.Load(cfg.Source, cfg.Output)
		if err != nil {
			return pages, err
		}
		pages = append(pages, page)
	}

	sort.Slice(pages[:], func(i, j int) bool {
		return pages[i].Date().After(pages[j].Date())
	})

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
		Link:        &feeds.Link{Href: cfg.SiteURL},
		Description: cfg.Description,
		Author:      &feeds.Author{Name: cfg.Name, Email: cfg.Email},
	}

	for _, page := range pages {
		feed.Add(&feeds.Item{
			Title:       page.Title(),
			Link:        &feeds.Link{Href: page.URL()},
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

func generateIndex(pages []*page.Page) error {
	index := &page.Page{}
	index.Assets = cfg.Assets
	index.Root = rootURI()
	index.Data.Title = cfg.Title
	index.Destination = filepath.Join(cfg.Output, "index.html")
	index.Template = filepath.Join("layouts", "index.html")
	index.Children = pages

	err := index.Generate()
	if err != nil {
		return err
	}

	fmt.Println("Wrote: ", index.Destination)
	return err
}

func buildTagsTree(pages []*page.Page) map[string][]*page.Page {
	tree := make(map[string][]*page.Page)
	for _, page := range pages {
		if len(page.Data.Tags) > 0 {
			for _, tag := range page.Data.Tags {
				tree[tag] = append(tree[tag], page)
			}
		}
	}

	return tree
}

func generateTags(pages []*page.Page) error {
	tree := buildTagsTree(pages)

	err := os.MkdirAll(filepath.Join(cfg.Output, "tag"), 0700)
	if err != nil {
		log.Fatal(err)
	}

	for tag, ps := range tree {
		p := &page.Page{}
		p.Assets = cfg.Assets
		p.Root = rootURI()
		p.Data.Title = cfg.Title
		p.Destination = filepath.Join(cfg.Output, "tag", fmt.Sprintf("%s.html", tag))
		p.Template = filepath.Join("layouts", "index.html")
		p.Children = ps

		err := p.Generate()
		if err != nil {
			return err
		}

		fmt.Println("Wrote: ", p.Destination)
	}

	return nil
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

	pages, err := load(cfg.Source)
	if err != nil {
		log.Fatal(err)
	}

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

	err = generateTags(pages)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Done!")
}
