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
	"gopkg.in/yaml.v2"
)

// Page contains everything needed to build a page and write it.
// This includes the following attributes:
//   * SourceFile: original source file
//   * Raw: raw markdown in bytes
//   * Data: data extracted from the file
//   * Content: parsed content into HTML
//   * Config: supporting config for this project
type Page struct {
	Content    template.HTML
	Config     config.Config
	Data       Data
	Raw        []byte
	SourceFile os.FileInfo
}

// Data contains the extracted meta-data from the original source.
// It's pulled from the raw content before parsing, and is then
// parsed separately as markdown into this struct.
//
// Any fields you want to add to the front-matter data should go here.
type Data struct {
	Title string
}

func (page *Page) source() string {
	file, err := os.Stat(page.Config.Source)
	if err != nil {
		return page.SourceFile.Name()
	}

	if !file.IsDir() {
		return page.SourceFile.Name()
	}

	return page.Config.Source + "/" + page.SourceFile.Name()
}

func (page *Page) basename() string {
	return filepath.Ext(page.SourceFile.Name())
}

func (page *Page) filename() string {
	return page.SourceFile.Name()[0 : len(page.SourceFile.Name())-len(page.basename())]
}

func (page *Page) outfile() string {
	return page.Config.Output + "/" + page.filename() + ".html"
}

func (page *Page) load() (ok bool, err error) {
	content, err := ioutil.ReadFile(page.source())
	if err != nil {
		return false, err
	}
	page.Data, err = parseFrontMatter(content)
	if err != nil {
		return false, err
	}
	r := regexp.MustCompile(FrontMatterRegex)
	raw := r.ReplaceAllString(string(content), "")
	page.Raw = []byte(raw)

	return true, nil
}

// FrontMatterRegex is a regular expression inspired by Jekyll.
// They have a constant YAML_FRONT_MATTER_REGEX, which is here:
//   https://github.com/jekyll/jekyll/blob/a944dd9/lib/jekyll/document.rb#L13
//
// We use this to pull out meta-data from the page content before parsing.
// The first use-case for this was a page title.
const FrontMatterRegex = `(?s)(---\s*\n.*?\n?)(---\s*\n?)`

func parseFrontMatter(content []byte) (data Data, err error) {
	r, err := regexp.Compile(FrontMatterRegex)
	if err != nil {
		return data, err
	}

	matches := r.FindAllStringSubmatch(string(content), -1)

	if len(matches) > 0 && len(matches[0]) > 0 {
		err = yaml.Unmarshal([]byte(matches[0][1]), &data)
		if err != nil {
			return data, err
		}
	}
	return data, nil
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
func (page Page) Generate() (ok bool, err error) {
	_, err = page.load()
	if err != nil {
		return false, err
	}

	f, err := os.Create(page.outfile())
	if err != nil {
		return false, err
	}
	w := bufio.NewWriter(f)

	tpl := template.New("content")
	tpl, err = tpl.Parse(string(page.Raw))
	if err != nil {
		log.Fatalf("got +%v", page.Raw)
		return false, err
	}

	buf := new(bytes.Buffer)
	err = tpl.Execute(buf, page)
	if err != nil {
		return false, err
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
		return false, err
	}
	t, err = t.Parse(AssetsTemplate)
	if err != nil {
		return false, err
	}

	err = t.Execute(w, page)

	if err != nil {
		return false, err
	}
	err = w.Flush()
	return true, err
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

func deferClose(closer io.Closer) {
	err := closer.Close()
	if err != nil {
		panic(err)
	}
}

func generateCSS(cssFiles []string, cssDir string) error {
	for _, file := range cssFiles {
		path := cssDir + "/" + file
		src := "assets/css/" + file

		from, err := os.Open(src)
		if err != nil {
			return err
		}
		defer deferClose(from)

		to, er := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0600)
		if er != nil {
			return err
		}
		defer deferClose(to)

		_, er = io.Copy(to, from)
		if er != nil {
			return err
		}
	}

	return nil
}

func setupCSSDir(output string) error {
	_, err := os.Stat(output)
	if err != nil && os.IsNotExist(err) {
		fmt.Printf("Making assets css output dir: %v\n", output)
		e := os.Mkdir(output, 0700)
		if e != nil {
			return e
		}
	}

	return nil
}

func generateImages(imgFiles []string, imgDir string) error {
	for _, file := range imgFiles {
		path := imgDir + "/" + file
		src := "assets/images/" + file

		from, err := os.Open(src)
		if err != nil {
			return err
		}
		defer deferClose(from)

		to, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0600)
		if err != nil {
			return err
		}
		defer deferClose(to)

		_, err = io.Copy(to, from)
		if err != nil {
			return err
		}
	}

	return nil
}

func setupImgDir(output string) error {
	_, err := os.Stat(output)
	if err != nil && os.IsNotExist(err) {
		fmt.Printf("Making assets images output dir: %v\n", output)
		e := os.Mkdir(output, 0700)
		if e != nil {
			return e
		}
	}

	return nil
}

func generateAssets(config config.Config) error {
	// generate css
	cssDir := config.Output + "/css"
	err := setupCSSDir(cssDir)
	if err != nil {
		return err
	}

	err = generateCSS(config.Assets.CSS, cssDir)
	if err != nil {
		return err
	}

	// generate images
	imgDir := config.Output + "/images"

	err = setupImgDir(imgDir)
	if err != nil {
		return err
	}

	err = generateImages(config.Assets.Images, imgDir)

	return err
}

func sourceFiles(source string) (files []os.FileInfo, err error) {
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

func setupOutputDir(output string) error {
	_, err := os.Stat(output)
	if err != nil && os.IsNotExist(err) {
		fmt.Printf("Making output dir: %v\n", output)
		e := os.Mkdir(output, 0700)
		if e != nil {
			return e
		}
	}

	return nil
}

func generateFiles(config config.Config) error {
	var err error

	files, err := sourceFiles(config.Source)
	if err != nil {
		return err
	}

	for _, file := range files {
		page := &Page{}
		page.Config = config
		page.SourceFile = file

		_, err = page.Generate()
		if err != nil {
			return err
		}
		fmt.Println("Wrote: ", page)
	}

	return err
}

// Stationery is the main entrypoint to this program.
// It's the original caller from inside main() and logs any errors that occur during file generation.
func Stationery() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	err = setupOutputDir(cfg.Output)
	if err != nil {
		log.Fatal(err)
	}

	err = generateAssets(cfg)
	if err != nil {
		log.Fatal(err)
	}

	err = generateFiles(cfg)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Done!")
}

func main() {
	Stationery()
	os.Exit(0)
}
