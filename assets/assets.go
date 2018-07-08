package assets

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
