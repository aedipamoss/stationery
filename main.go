package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/aedipamoss/stationery/config"
	"github.com/aedipamoss/stationery/page"
)

func deferClose(closer io.Closer) {
	err := closer.Close()
	if err != nil {
		panic(err)
	}
}

func dirExistOrMkdir(path string) error {
	_, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		fmt.Printf("Making dir: %v\n", path)
		e := os.Mkdir(path, 0700)
		if e != nil {
			return e
		}
	}

	return nil
}

func copyFilesToDest(files []string, src string, dest string) error {
	for _, file := range files {
		path := filepath.Join(src, file)
		src := filepath.Join(dest, file)

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

func generateAssets(config config.Config) error {
	// generate css
	cssDir := filepath.Join(config.Output, "css")
	err := dirExistOrMkdir(cssDir)
	if err != nil {
		return err
	}

	err = copyFilesToDest(config.Assets.CSS, cssDir, "assets/css/")
	if err != nil {
		return err
	}

	// generate images
	imgDir := filepath.Join(config.Output, "images")
	err = dirExistOrMkdir(imgDir)
	if err != nil {
		return err
	}

	err = copyFilesToDest(config.Assets.Images, imgDir, "assets/images/")

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

func source(path string, file os.FileInfo) string {
	name := file.Name()
	file, err := os.Stat(path)
	if err != nil {
		return name
	}

	if !file.IsDir() {
		return name
	}

	return filepath.Join(path, name)
}

func destination(path string, file os.FileInfo) string {
	name := file.Name()
	basename := filepath.Ext(name)
	filename := name[0 : len(name)-len(basename)]

	return filepath.Join(path, filename+".html")
}

func generateFiles(config config.Config) error {
	var err error

	files, err := sourceFiles(config.Source)
	if err != nil {
		return err
	}

	for _, file := range files {
		page := &page.Page{}
		page.Assets = config.Assets
		page.FileInfo = file
		page.Destination = destination(config.Output, file)
		page.Source = source(config.Source, file)
		page.Template = config.Template

		err = page.Load()
		if err != nil {
			return err
		}

		err = page.Generate()
		if err != nil {
			return err
		}
		fmt.Println("Wrote: ", file.Name())
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

	err = dirExistOrMkdir(cfg.Output)
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
