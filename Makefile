.PHONY: test all build clean test test-full update_version cmds install build-and-upload

GO_FILES=$(shell find . -name \*.go)
GOMAPNIK_CONFIG=Godeps/_workspace/src/github.com/omniscale/go-mapnik/mapnik_config.go

BUILD_DATE=$(shell date +%Y%m%d)
BUILD_REV=$(shell git rev-parse --short HEAD)
BUILD_VERSION=dev-$(BUILD_DATE)-$(BUILD_REV)
VERSION_LDFLAGS=-X github.com/omniscale/magnacarto.buildVersion $(BUILD_VERSION)

GO=godep go

uname_S := $(shell sh -c 'uname -s 2>/dev/null || echo not' | tr '[:upper:]' '[:lower:]')
uname_M := $(shell sh -c 'uname -m 2>/dev/null || echo not' | tr '[:upper:]' '[:lower:]')

all: build test

CMDS=magnacarto magnacarto-mapnik magnaserv

$(GOMAPNIK_CONFIG):
	$(GO) generate github.com/omniscale/go-mapnik

build: $(GO_FILES) $(GOMAPNIK_CONFIG)
	$(GO) build -ldflags "$(VERSION_LDFLAGS)" -v ./...

magnacarto: $(GO_FILES)
	$(GO) build -ldflags "$(VERSION_LDFLAGS)" ./cmd/magnacarto

magnaserv: $(GO_FILES)
	$(GO) build -ldflags "$(VERSION_LDFLAGS)" ./cmd/magnaserv

magnacarto-mapnik: $(GO_FILES)
	$(GO) build -ldflags "$(VERSION_LDFLAGS)" ./render/magnacarto-mapnik

install: $(GO_FILES)
	$(GO) install -ldflags "$(VERSION_LDFLAGS)" ./cmd/...

cmds: build $(CMDS)

clean:
	rm -f $(CMDS)
	rm -rf Godeps/_workspace/pkg/

VERSION = $(shell ./$(firstword $(CMDS)) -version)
BIN_VERSION = $(VERSION)$(DISTRIBUTION)-$(uname_S)-$(uname_M)

dist: cmds
	mkdir -p dist/
	cp magnacarto dist/magnacarto-$(BIN_VERSION)
	cp magnaserv dist/magnaserv-$(BIN_VERSION)
	cp magnacarto-mapnik dist/magnacarto-mapnik-$(BIN_VERSION)

test:
	$(GO) test -i ./...
	$(GO) test -test.short -parallel 4 ./...

test-full:
	$(GO) test ./... -i
	export PATH=$(shell pwd):$$PATH; $(GO) test -parallel 4 ./...

