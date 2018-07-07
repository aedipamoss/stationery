OS ?= $(shell uname)
GOFILES = $(shell find . -name '*.go' -not -path './vendor/*')

.PHONY: build
build:
ifeq ($(OS), Darwin)
	GOOS=darwin GOARCH=amd64 go build -o build/darwin/amd64/stationery
else ifeq ($(OS), Linux)
	GOOS=linux GOARCH=amd64 go build -o build/linux/amd64/stationery
endif

.PHONY: build/*
$(BUILD_DIR)/%/amd64/$(EXECUTABLE): $(GOFILES)
	GOOS=$* GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o build/$*/amd64/stationery .

.PHONY: build-all
build-all: build/darwin/amd64/stationery build/linux/amd64/stationery

.PHONY: clean
clean:
	go clean
	rm -rf build

.PHONY: doc
doc:
	godoc -http=:6060

.PHONY: test
test:
	go test -v -short ./...

.PHONY: lint
lint:
	gometalinter --deadline 60s --vendor ./...
