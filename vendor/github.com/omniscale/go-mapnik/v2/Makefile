.PHONY: test build

export CGO_LDFLAGS = $(shell mapnik-config --libs)
export CGO_CXXFLAGS = $(shell mapnik-config --cxxflags --includes --dep-includes | tr '\n' ' ')

MAPNIK_LDFLAGS=-X github.com/omniscale/go-mapnik/v2.fontPath=$(shell mapnik-config --fonts) \
	-X github.com/omniscale/go-mapnik/v2.pluginPath=$(shell mapnik-config --input-plugins)

build:
	go build -ldflags "$(MAPNIK_LDFLAGS)"

test:
	go test -ldflags "$(MAPNIK_LDFLAGS)"
