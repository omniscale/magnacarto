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
MAPNIK_LDFLAGS=-X github.com/omniscale/go-mapnik.fontPath=$(shell mapnik-config --fonts) \
	-X github.com/omniscale/go-mapnik.pluginPath=$(shell mapnik-config --input-plugins)

MAPNIK_CGO_LDFLAGS = $(shell mapnik-config --libs) -lboost_system
MAPNIK_CGO_CXXFLAGS = $(shell mapnik-config --cxxflags --includes --dep-includes | tr '\n' ' ')

uname_S = $(shell sh -c 'uname -s 2>/dev/null || echo not' | tr '[:upper:]' '[:lower:]')
uname_M = $(shell sh -c 'uname -m 2>/dev/null || echo not' | tr '[:upper:]' '[:lower:]')

all: build test

CMDS=magnacarto magnaserv magnacarto-mapnik

build: $(CMDS)

magnacarto: $(DEPS)
	go build -ldflags "$(VERSION_LDFLAGS)" ./cmd/magnacarto

magnaserv: $(DEPS)
	go build -ldflags "$(VERSION_LDFLAGS)" ./cmd/magnaserv

magnacarto-mapnik: export CGO_CXXFLAGS = $(MAPNIK_CGO_CXXFLAGS)
magnacarto-mapnik: export CGO_LDFLAGS = $(MAPNIK_CGO_LDFLAGS)
magnacarto-mapnik: $(DEPS)
	go build -ldflags "$(VERSION_LDFLAGS) $(MAPNIK_LDFLAGS)" ./render/magnacarto-mapnik || echo "WARNING: failed to build mapnik plugin"

install: export CGO_CXXFLAGS = $(MAPNIK_CGO_CXXFLAGS)
install: export CGO_LDFLAGS = $(MAPNIK_CGO_LDFLAGS)
install: $(DEPS)
	go install -ldflags "$(VERSION_LDFLAGS)" ./cmd/...
	go install -ldflags "$(VERSION_LDFLAGS) $(MAPNIK_LDFLAGS)" ./render/magnacarto-mapnik || echo "WARNING: failed to build mapnik plugin"

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
SHORT_TEST_PACKAGES=$(shell go list ./... | grep -Ev '/render|/regression|/vendor')
FULL_TEST_PACKAGES=$(shell go list ./... | grep -Ev '/vendor')

test:
	go test -i $(SHORT_TEST_PACKAGES)
	go test -test.short -parallel 4 $(SHORT_TEST_PACKAGES)

test-full: export CGO_CXXFLAGS = $(MAPNIK_CGO_CXXFLAGS)
test-full: export CGO_LDFLAGS = $(MAPNIK_CGO_LDFLAGS)
test-full:
	go test -i $(FULL_TEST_PACKAGES)
	export PATH=$(shell pwd):$$PATH; go test -parallel 4 $(FULL_TEST_PACKAGES)

test-coverage:
	go test -coverprofile magnacarto.coverprofile -coverpkg ./... -covermode count ./...
test-coverage-html: test-coverage
	go tool cover -html magnacarto.coverprofile

