package assets

import (
	"os"
	"path/filepath"

	"github.com/aedipamoss/stationery/fileutils"
)

// List is a struct containing all the CSS, JavaScript, and Images to be built.
type List struct {
	CSS    []string
	JS     []string
	Images []string
}

// Template defines a template which utilizes the Assets field of a page
// including the "CSS" array of stylesheets to include in the header.
const Template = `
{{ define "assets" }}
  {{ range .Assets.CSS }}
    <link type="text/css" rel="stylesheet" href="css/{{ . }}">
  {{ end }}
{{ end }}
`

func setupAndCopy(files []string, src string, dest string) error {
	from := filepath.Join("assets", src)
	to := filepath.Join(dest, src)

	err := os.MkdirAll(to, 0700)
	if err != nil {
		return err
	}

	return fileutils.CopyFiles(files, from, to)
}

// Generate will copy assets from each field, CSS, Images, and JS.
// It copies each file listed to the provided destination.
func (assets *List) Generate(dest string) error {
	err := setupAndCopy(assets.CSS, "css", dest)
	if err != nil {
		return err
	}

	return setupAndCopy(assets.Images, "images", dest)
}
