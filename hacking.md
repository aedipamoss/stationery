# Hacking

This project requires `make` and Go 1.10+ to build.

## Make tasks

By default `make` will only build the binary for your current OS, either Linux or Darwin.

You can build binaries for both using `make build-all`.

* `make clean`: Clean the build directory
* `make doc`: Run `godoc` on localhost:6060
* `make test`: Run `go test` with short mode, excluding vendor
* `make lint`: Runs `gometalinter` so make sure you have that installed

## Dependencies

Install [dep](https://github.com/golang/dep) and run `dep ensure`.

## Linting

Install [gometalinter](https://github.com/alecthomas/gometalinter) via:

```
$ go get -u github.com/alecthomas/gometalinter
$ gometalinter --install
```

Now `make lint` should work.

## Documentation

Install [godoc](https://godoc.org/golang.org/x/tools/cmd/godoc) via:

```
$ go get -u golang.org/x/tools/cmd/godoc
```

Run `make doc` and open up your browser to http://localhost:6060/pkg/github.com/aedipamoss/stationery/


