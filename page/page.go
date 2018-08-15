package page

import (
	"bufio"
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"time"

	"github.com/aedipamoss/stationery/assets"
	"github.com/aedipamoss/stationery/fileutils"
	blackfriday "gopkg.in/russross/blackfriday.v2"
	yaml "gopkg.in/yaml.v2"
)

// Page contains everything needed to build a page and write it.
type Page struct {
	Assets   *assets.List  // assets available to this page
	Children []*Page       // children pages used for index templates
	Content  template.HTML // parsed content into HTML
	Data     struct {      // extracted meta-data from the file
		Description string
		Image       string
		Title       string
		Timestamp   string
		Tags        []string
		Twitter     string // twitter user handle who created this page
	}
	Destination string      // path to write this page out to
	FileInfo    os.FileInfo // original source file info
	Raw         string      // raw markdown after subbing data
	Root        string      // parent of this page, usually config.SiteURL
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
func (page Page) Slug() string {
	if page.FileInfo != nil {
		return fileutils.Basename(page.FileInfo)
	}

	stat, err := os.Stat(page.Destination)
	if err != nil {
		panic(err)
	}

	return fileutils.Basename(stat)
}

// Title is used when printing the index page as the anchor text currently in generate.IndexTemplate.
func (page Page) Title() string {
	if page.Data.Title != "" {
		return page.Data.Title
	}

	return page.Slug()
}

// Write a bunch of strings to a buffer then return a string
func toString(strs ...string) string {
	var buf bytes.Buffer
	var err error

	for _, str := range strs {
		_, err = buf.Write([]byte(str))
		if err != nil {
			panic(err)
		}
	}

	return buf.String()
}

// Tags will build a list of tags and their links in an HTML safe way.
func (page Page) Tags() template.HTML {
	var str string
	if len(page.Data.Tags) > 0 {
		str += "<br>"
		for _, tag := range page.Data.Tags {
			str += `<span class="tag">#`
			str += fmt.Sprintf(`<a href="%stag/%s.html">`, page.Root, tag)
			str += tag
			str += `</a>`
			str += `</span>`
		}
	}
	// nolint: gosec
	return template.HTML(str)
}

// MetaTags builds a list of meta tags for the header of a page.
func (page Page) MetaTags() template.HTML {
	var str string
	if page.Data.Description != "" {
		str += fmt.Sprintf(`<meta name="description" content="%s">`, page.Data.Description)
		str += fmt.Sprintf(`<meta property="og:description" content="%s" />`, page.Data.Description)
	}

	if page.Data.Twitter != "" {
		str += `<meta name="twitter:card" content="summary" />`
		str += fmt.Sprintf(`<meta name="twitter:site" content="@%s" />`, page.Data.Twitter)
		str += fmt.Sprintf(`<meta name="twitter:creator" content="@%s" />`, page.Data.Twitter)
	}

	str += fmt.Sprintf(`<meta property="og:url" content="%s" />`, page.URL())
	str += fmt.Sprintf(`<meta property="og:title" content="%s" />`, page.Title())

	if page.Data.Image != "" {
		str += fmt.Sprintf(`<meta property="og:image" content="%s%s" />`, page.Root, page.Data.Image)
	}

	// nolint: gosec
	return template.HTML(str)
}

// Link is used when printing a page's link inside generate.IndexTemplate
func (page Page) Link() template.HTML {
	str := toString(
		fmt.Sprintf(`<a href="%s%s.html">`, page.Root, page.Slug()),
		page.Title(),
		"</a>",
		"<br>",
		`<span class="page_date">`,
		page.DateString(),
		"</span>")

	str += string(page.Tags())

	// nolint: gosec
	return template.HTML(str)
}

// URL is used when generating the rss feed for the site.
func (page Page) URL() string {
	url, err := url.Parse("")
	if err != nil {
		panic(err)
	}

	url, err = url.Parse(page.Root)
	if err != nil {
		panic(err)
	}

	url.Path = path.Join(url.Path, page.Slug()+".html")
	return url.String()
}

// Description is used when generating the rss feed for the site.
func (page Page) Description() string {
	if page.Data.Description != "" {
		return page.Data.Description
	}

	return page.Title()
}

// DateString returns a string formatted date of the (*page).Date()
func (page Page) DateString() string {
	return page.Date().Format("Jan _2, 2006")
}

// Date creates a time.Time from the meta-data of a page's Data.Timestamp field by parsing it with time.RFC3339.
func (page Page) Date() time.Time {
	if page.Data.Timestamp != "" {
		t, err := time.Parse(time.RFC3339, page.Data.Timestamp)
		if err != nil {
			panic(err)
		}

		return t
	}

	if page.FileInfo != nil {
		return page.FileInfo.ModTime()
	}

	return time.Now()
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
	// nolint: gosec
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
