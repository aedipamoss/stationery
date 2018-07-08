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

// Generate will copy assets from each field, CSS, Images, and JS.
// It copies each file listed to the provided destination.
func (assets *List) Generate(dest string) error {
	// generate css
	cssDir := filepath.Join(dest, "css")
	err := os.MkdirAll(cssDir, 0700)
	if err != nil {
		return err
	}

	err = fileutils.CopyFiles(assets.CSS, cssDir, "assets/css/")
	if err != nil {
		return err
	}

	// generate images
	imgDir := filepath.Join(dest, "images")
	err = os.MkdirAll(imgDir, 0700)
	if err != nil {
		return err
	}

	err = fileutils.CopyFiles(assets.Images, imgDir, "assets/images/")

	return err
}
