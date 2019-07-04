SHELL:=/bin/bash
.PHONY: test all build install cmds clean dist test test-full

DEPS:=$(shell find . -name \*.go)

BUILD_DATE=$(shell date +%Y%m%d)
BUILD_REV=$(shell git rev-parse --short HEAD)
BUILD_BRANCH=$(shell git rev-parse --abbrev-ref HEAD)
TAG=$(shell git name-rev --tags --name-only $(BUILD_REV))
ifeq ($(TAG),undefined)
	BUILD_VERSION=$(BUILD_BRANCH)-$(BUILD_DATE)-$(BUILD_REV)
else
	# use TAG but strip v of v1.2.3
	BUILD_VERSION=$(TAG:v%=%)
endif

VERSION_LDFLAGS=-X github.com/omniscale/magnacarto.Version=$(BUILD_VERSION)
MAPNIK_LDFLAGS=-X github.com/omniscale/go-mapnik.fontPath=$(shell mapnik-config --fonts)
	-X github.com/omniscale/go-mapnik.pluginPath=$(shell mapnik-config --plugins)

MAPNIK_CGO_LDFLAGS = $(shell mapnik-config --libs) -lboost_system
MAPNIK_CGO_CXXFLAGS = $(shell mapnik-config --cxxflags --includes --dep-includes | tr '\n' ' ')

GO:=$(if $(shell go version |grep 'go1.5'),GO15VENDOREXPERIMENT=1,) go

uname_S = $(shell sh -c 'uname -s 2>/dev/null || echo not' | tr '[:upper:]' '[:lower:]')
uname_M = $(shell sh -c 'uname -m 2>/dev/null || echo not' | tr '[:upper:]' '[:lower:]')

all: build test

CMDS=magnacarto magnaserv magnacarto-mapnik

build: $(DEPS)
	$(GO) build -ldflags "$(VERSION_LDFLAGS)" -v ./...

magnacarto: $(DEPS)
	$(GO) build -ldflags "$(VERSION_LDFLAGS)" ./cmd/magnacarto

magnaserv: $(DEPS)
	$(GO) build -ldflags "$(VERSION_LDFLAGS)" ./cmd/magnaserv

magnacarto-mapnik: export CGO_CXXFLAGS = $(MAPNIK_CGO_CXXFLAGS)
magnacarto-mapnik: export CGO_LDFLAGS = $(MAPNIK_CGO_LDFLAGS)
magnacarto-mapnik: $(DEPS)
	$(GO) build -ldflags "$(VERSION_LDFLAGS) $(MAPNIK_LDFLAGS)" ./render/magnacarto-mapnik || echo "WARNING: failed to build mapnik plugin"

install: export CGO_CXXFLAGS = $(MAPNIK_CGO_CXXFLAGS)
install: export CGO_LDFLAGS = $(MAPNIK_CGO_LDFLAGS)
install: $(DEPS)
	$(GO) install -ldflags "$(VERSION_LDFLAGS)" ./cmd/...
	$(GO) install -ldflags "$(VERSION_LDFLAGS) $(MAPNIK_LDFLAGS)" ./render/magnacarto-mapnik || echo "WARNING: failed to build mapnik plugin"

cmds: build $(CMDS)

clean:
	rm -f $(CMDS)

BIN_VERSION = $(BUILD_VERSION)$(DISTRIBUTION)-$(uname_S)-$(uname_M)

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

