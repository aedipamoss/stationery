package page

import (
	"bufio"
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"regexp"

	"github.com/aedipamoss/stationery/assets"
	"github.com/aedipamoss/stationery/fileutils"
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
	Raw         string      // raw markdown after subbing data
	Source      string      // path to the original source file
	Template    string      // template used for this page
}

// Timestamp is a member function made available in the page template.
// So you can write `{{ .Timestamp "2018-03-24" }}`;
// In the resulting HTML will get an anchor tag to that timestamp.
func (page Page) Timestamp(timestamp string) string {
	return fmt.Sprint("[@ ", timestamp, "](#", timestamp, ")")
}

// Slug is used to reference the destination for a page without the extension.
// It's used both in generate.IndexTemplate and (*page.Page).setDestination()
// Be careful, this assumes there is a *os.FileInfo attached to the page.
func (page Page) Slug() string {
	basename := fileutils.Basename(page.FileInfo)
	return basename
}

// Title is used when printing the index page as the anchor text currently in generate.IndexTemplate.
func (page Page) Title() string {
	return page.Data.Title
}

// FrontMatterRegex is a regular expression inspired by Jekyll.
// They have a constant YAML_FRONT_MATTER_REGEX, which is here:
//   https://github.com/jekyll/jekyll/blob/a944dd9/lib/jekyll/document.rb#L13
//
// We use this to pull out meta-data from the page content before parsing.
// The first use-case for this was a page title.
const FrontMatterRegex = `(?s)(---\s*\n.*?\n?)(---\s*\n?)`

// Parses the front-matter data into the page and returns the content stripped of meta-data.
// This function is called directly by parseRaw().
func (page *Page) parseFrontMatter(content []byte) (string, error) {
	r := regexp.MustCompile(FrontMatterRegex)

	matches := r.FindAllStringSubmatch(string(content), -1)

	if len(matches) > 0 && len(matches[0]) > 0 {
		err := yaml.Unmarshal([]byte(matches[0][1]), &page.Data)
		if err != nil {
			return string(content), err
		}
	}

	return r.ReplaceAllString(string(content), ""), nil
}

// Reads the file from source and parses the meta-data into the page.
// This function also sets the raw data field after parsing.
// This function is called directly in Load().
func (page *Page) parseRaw() error {
	content, err := ioutil.ReadFile(page.Source)
	if err != nil {
		return err
	}

	raw, err := page.parseFrontMatter(content)
	if err != nil {
		return err
	}

	page.Raw = raw

	return err
}

// Execute the raw markdown into a generate text template.
// This function is called directly by parseContent().
//
// BUG(ae): there is probably a simpler way to do this without using template
func (page *Page) executeContent() ([]byte, error) {
	buf := new(bytes.Buffer)
	tpl := template.New("content")
	tpl, err := tpl.Parse(page.Raw)
	if err != nil {
		return buf.Bytes(), err
	}

	err = tpl.Execute(buf, page)
	if err != nil {
		return buf.Bytes(), err
	}

	return buf.Bytes(), err
}

// Set the page content after parsing the markdown.
// This function is called directly in Load().
func (page *Page) parseContent() error {
	buf, err := page.executeContent()
	if err != nil {
		return err
	}
	parsed := blackfriday.Run(buf)
	// nolint: gas
	page.Content = template.HTML(string(parsed[:]))

	return nil
}

func (page *Page) setSource(src string) error {
	name := page.FileInfo.Name()
	file, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !file.IsDir() {
		page.Source = name
		return nil
	}

	page.Source = filepath.Join(src, name)
	return nil
}

func (page *Page) setDestination(dest string) error {
	page.Destination = filepath.Join(dest, page.Slug()+".html")

	return nil
}

// Load reads the page from source and parses the content and front-matter into data.
func (page *Page) Load(src string, dest string) error {
	err := page.setSource(src)
	if err != nil {
		return err
	}

	err = page.setDestination(dest)
	if err != nil {
		return err
	}

	err = page.parseRaw()
	if err != nil {
		return err
	}

	err = page.parseContent()

	return err
}

// Create a buffered writer at the page destination.
// This function is called directly in Generate().
func (page *Page) createDestination() (*bufio.Writer, error) {
	f, err := os.Create(page.Destination)
	if err != nil {
		return nil, err
	}
	return bufio.NewWriter(f), err
}

// Parse the page template along with assets and return a template ready for execution.
// This function is called directly in Generate().
func (page *Page) parseTemplate() (*template.Template, error) {
	tmpl, err := ioutil.ReadFile(page.Template)
	if err != nil {
		return nil, err
	}

	t, err := template.New("page").Parse(string(tmpl))
	if err != nil {
		return t, err
	}
	t.Funcs(template.FuncMap{
		"exists": func(name string, data interface{}) bool {
			v := reflect.ValueOf(data)
			if v.Kind() == reflect.Ptr {
				v = v.Elem()
			}
			if v.Kind() != reflect.Struct {
				return false
			}
			return v.FieldByName(name).IsValid()
		},
	})
	t, err = t.Parse(assets.Template)
	return t, err
}

// Generate does exactly what the name implies.
//
// This function will execute the given template and content along with any assets into a file on disk.
//
// BUG(ae): should throw an error if content isn't loaded yet.
func (page Page) Generate() error {
	wrtr, err := page.createDestination()
	if err != nil {
		return err
	}

	tmpl, err := page.parseTemplate()
	if err != nil {
		return err
	}

	err = tmpl.Execute(wrtr, page)
	if err != nil {
		return err
	}
	err = wrtr.Flush()
	return err
}
