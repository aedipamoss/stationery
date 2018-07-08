package page

import (
	"bufio"
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"regexp"

	"github.com/aedipamoss/stationery/assets"
	blackfriday "gopkg.in/russross/blackfriday.v2"
	yaml "gopkg.in/yaml.v2"
)

// Page contains everything needed to build a page and write it.
type Page struct {
	Assets  *assets.List  // assets available to this page
	Content template.HTML // parsed content into HTML
	Data    struct {      // extracted meta-data from the file
		Title string
	}
	Destination string      // path to write this page out to
	FileInfo    os.FileInfo // original source file info
	Raw         []byte      // raw markdown in bytes
	Source      string      // path to the original source file
	Template    string      // template used for this page
}

// Timestamp is a member function made available in the page template.
// So you can write `{{ .Timestamp "2018-03-24" }}`;
// In the resulting HTML will get an anchor tag to that timestamp.
func (page Page) Timestamp(timestamp string) string {
	return fmt.Sprint("[@ ", timestamp, "](#", timestamp, ")")
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

// Used in Generate()
func (page *Page) load() error {
	content, err := ioutil.ReadFile(page.Source)
	if err != nil {
		return err
	}
	err = page.parseFrontMatter(content)
	return err
}

// Generate does exactly what the name implies.
//
// Given a page this function will parse it's content from markdown to HTML,
// including the template and it's assets into a file on disk.
func (page Page) Generate() error {
	err := page.load()
	if err != nil {
		return err
	}

	f, err := os.Create(page.Destination)
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

	tmpl, err := ioutil.ReadFile(page.Template)
	if err != nil {
		log.Fatal(err)
	}

	t, err := template.New("page").Parse(string(tmpl))
	if err != nil {
		return err
	}
	t, err = t.Parse(assets.Template)
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
