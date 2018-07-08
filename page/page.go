package page

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
	yaml "gopkg.in/yaml.v2"
)

// Page contains everything needed to build a page and write it.
type Page struct {
	Content    template.HTML // parsed content into HTML
	Config     config.Config // supporting config for this project
	Data       Data          // extracted meta-data from the file
	Raw        []byte        // raw markdown in bytes
	SourceFile os.FileInfo   // original source file
}

// Data contains the extracted meta-data from the original source.
// It's pulled from the raw content before parsing, and is then
// parsed separately as markdown into this struct.
//
// Any fields you want to add to the front-matter data should go here.
type Data struct {
	Title string
}

// Only used in load()
func (page *Page) source() string {
	name := page.SourceFile.Name()
	file, err := os.Stat(page.Config.Source)
	if err != nil {
		return name
	}

	if !file.IsDir() {
		return name
	}

	return filepath.Join(page.Config.Source, name)
}

// Used in Generate()
func (page *Page) destination() string {
	name := page.SourceFile.Name()
	basename := filepath.Ext(name)
	filename := name[0 : len(name)-len(basename)]

	return filepath.Join(page.Config.Output, filename+".html")
}

// Used in Generate()
func (page *Page) load() error {
	content, err := ioutil.ReadFile(page.source())
	if err != nil {
		return err
	}
	err = page.parseFrontMatter(content)
	return err
}

// FrontMatterRegex is a regular expression inspired by Jekyll.
// They have a constant YAML_FRONT_MATTER_REGEX, which is here:
//   https://github.com/jekyll/jekyll/blob/a944dd9/lib/jekyll/document.rb#L13
//
// We use this to pull out meta-data from the page content before parsing.
// The first use-case for this was a page title.
const FrontMatterRegex = `(?s)(---\s*\n.*?\n?)(---\s*\n?)`

// Used in load()
func (page *Page) parseFrontMatter(content []byte) error {
	r := regexp.MustCompile(FrontMatterRegex)

	matches := r.FindAllStringSubmatch(string(content), -1)

	if len(matches) > 0 && len(matches[0]) > 0 {
		err := yaml.Unmarshal([]byte(matches[0][1]), &page.Data)
		if err != nil {
			return err
		}
	}

	raw := r.ReplaceAllString(string(content), "")
	page.Raw = []byte(raw)
	return nil
}

// Timestamp is a member function made available in the page template.
// So you can write `{{ .Timestamp "2018-03-24" }}`;
// In the resulting HTML will get an anchor tag to that timestamp.
func (page Page) Timestamp(timestamp string) string {
	return fmt.Sprint("[@ ", timestamp, "](#", timestamp, ")")
}

// Generate does exactly what the name implies.
//
// Given a page this function will parse it's content from markdown to HTML,
// including the template from config and it's assets into a file on disk.
func (page Page) Generate() error {
	err := page.load()
	if err != nil {
		return err
	}

	f, err := os.Create(page.destination())
	if err != nil {
		return err
	}
	w := bufio.NewWriter(f)

	tpl := template.New("content")
	tpl, err = tpl.Parse(string(page.Raw))
	if err != nil {
		log.Fatalf("got +%v", page.Raw)
		return err
	}

	buf := new(bytes.Buffer)
	err = tpl.Execute(buf, page)
	if err != nil {
		return err
	}

	parsed := blackfriday.Run(buf.Bytes())
	// nolint: gas
	page.Content = template.HTML(string(parsed[:]))

	tmpl, err := ioutil.ReadFile(page.Config.Template)
	if err != nil {
		log.Fatal(err)
	}

	t, err := template.New("page").Parse(string(tmpl))
	if err != nil {
		return err
	}
	t, err = t.Parse(AssetsTemplate)
	if err != nil {
		return err
	}

	err = t.Execute(w, page)

	if err != nil {
		return err
	}
	err = w.Flush()
	return err
}

// AssetsTemplate defines a template which utilizes the Config struct
// including the fields "Assets.CSS" as an array of stylesheet names
const AssetsTemplate = `
{{ define "assets" }}
  {{ range .Config.Assets.CSS }}
    <link type="text/css" rel="stylesheet" href="css/{{ . }}">
  {{ end }}
{{ end }}
`
