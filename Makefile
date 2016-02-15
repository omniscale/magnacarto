SHELL:=/bin/bash
.PHONY: test all build install cmds clean dist test test-full

GO_FILES:=$(shell find . -name \*.go)
GOMAPNIK_CONFIG:=vendor/github.com/omniscale/go-mapnik/mapnik_config.go
DEPS:=$(GO_FILES) $(GOMAPNIK_CONFIG)

BUILD_DATE=$(shell date +%Y%m%d)
BUILD_REV=$(shell git rev-parse --short HEAD)
BUILD_VERSION=dev-$(BUILD_DATE)-$(BUILD_REV)
VERSION_LDFLAGS=-X github.com/omniscale/magnacarto.buildVersion=$(BUILD_VERSION)

GO:=$(if $(shell go version |grep 'go1.5'),GO15VENDOREXPERIMENT=1,) go

uname_S = $(shell sh -c 'uname -s 2>/dev/null || echo not' | tr '[:upper:]' '[:lower:]')
uname_M = $(shell sh -c 'uname -m 2>/dev/null || echo not' | tr '[:upper:]' '[:lower:]')

all: build test

CMDS=magnacarto magnaserv magnacarto-mapnik

$(GOMAPNIK_CONFIG):
	$(GO) generate ./vendor/github.com/omniscale/go-mapnik

build: $(DEPS)
	$(GO) build -ldflags "$(VERSION_LDFLAGS)" -v ./...

magnacarto: $(DEPS)
	$(GO) build -ldflags "$(VERSION_LDFLAGS)" ./cmd/magnacarto

magnaserv: $(DEPS)
	$(GO) build -ldflags "$(VERSION_LDFLAGS)" ./cmd/magnaserv

magnacarto-mapnik: $(DEPS)
	$(GO) build -ldflags "$(VERSION_LDFLAGS)" ./render/magnacarto-mapnik || echo "WARNING: failed to build mapnik plugin"

install: $(DEPS)
	$(GO) install -ldflags "$(VERSION_LDFLAGS)" ./cmd/...
	$(GO) install -ldflags "$(VERSION_LDFLAGS)" ./render/magnacarto-mapnik || echo "WARNING: failed to build mapnik plugin"

cmds: build $(CMDS)

clean:
	rm -f $(CMDS)

VERSION = $(shell ./$(firstword $(CMDS)) -version)
BIN_VERSION = $(VERSION)$(DISTRIBUTION)-$(uname_S)-$(uname_M)

dist: cmds
	mkdir -p dist/
	cp magnacarto dist/magnacarto-$(BIN_VERSION)
	cp magnaserv dist/magnaserv-$(BIN_VERSION)
	cp magnacarto-mapnik dist/magnacarto-mapnik-$(BIN_VERSION)

# exclude render and regression packages in non-full tests
SHORT_TEST_PACKAGES=$(shell $(GO) list ./... | grep -Ev '/render|/regression|/vendor')
FULL_TEST_PACKAGES=$(shell $(GO) list ./... | grep -Ev '/vendor')

test:
	$(GO) test -i $(SHORT_TEST_PACKAGES)
	$(GO) test -test.short -parallel 4 $(SHORT_TEST_PACKAGES)

test-full:
	$(GO) test -i $(FULL_TEST_PACKAGES)
	export PATH=$(shell pwd):$$PATH; $(GO) test -parallel 4 $(FULL_TEST_PACKAGES)

comma:= ,
empty:=
space:= $(empty) $(empty)
COVER_IGNORE:='/vendor|/regression|/render|/cmd'
COVER_PACKAGES:= $(shell $(GO) list ./... | grep -Ev $(COVER_IGNORE))
COVER_PACKAGES_LIST:=$(subst $(space),$(comma),$(COVER_PACKAGES))

test-coverage:
	mkdir -p .coverprofile
	rm -f .coverprofile/*
	$(GO) list -f '{{if gt (len .TestGoFiles) 0}}"$(GO) test -covermode count -coverprofile ./.coverprofile/{{.Name}}-$$$$.coverprofile -coverpkg $(COVER_PACKAGES_LIST) {{.ImportPath}}"{{end}}' ./... \
		| grep -Ev $(COVER_IGNORE) \
		| xargs -n 1 bash -c
	$(GOPATH)/bin/gocovmerge .coverprofile/*.coverprofile > overalls.coverprofile
	rm -rf .coverprofile

test-coverage-html: test-coverage
	$(GO) tool cover -html overalls.coverprofile

