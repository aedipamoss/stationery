package generate

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/aedipamoss/stationery/config"
	"github.com/aedipamoss/stationery/page"
)

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
		page.Template = config.Template

		err = page.Load(config.Source, config.Output)
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

	err = generateFiles(cfg)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Done!")
}
