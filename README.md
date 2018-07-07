# stationery

A static blog generator written in the Go programming language.

## Examples

* .
* ..
* ...
* ....

## Getting Started

### Installation

### Setting up a blog

### Generating your site

## What's a blog?

* Blog format
  * Timestamps, etc
* Meta-data
  * Description, tags, etc
* Assets
  * JavaScript, CSS, Images
* RSS

## Development

This project requires `make` and Go 1.10+ to build.

### Make tasks

By default `make` will only build the binary for your current OS, either Linux or Darwin.

You can build binaries for both using `make build-all`.

* `make clean`: Clean the build directory
* `make doc`: Run `godoc` on localhost:6060
* `make test`: Run `go test` with short mode, excluding vendor
* `make lint`: Runs `gometalinter` so make sure you have that installed

### Dependencies

Install [dep](https://github.com/golang/dep) and run `dep ensure`.

### Linting

Install [gometalinter](https://github.com/alecthomas/gometalinter) via:

```
$ go get -u github.com/alecthomas/gometalinter
$ gometalinter --install
```

Now `make lint` should work.

### Documentation

Install [godoc](https://godoc.org/golang.org/x/tools/cmd/godoc) via:

```
$ go get -u golang.org/x/tools/cmd/godoc
```

Run `make doc` and open up your browser to http://localhost:6060/pkg/github.com/aedipamoss/stationery/

## License

Copyright © 2018 Ædipa Moss

Distributed under the MIT license, please see [LICENSE](LICENSE) for full copy.

