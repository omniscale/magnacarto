Magnacarto
==========


Magnacarto is a CartoCSS map style processor that generates Mapnik XML and MapServer map files.


It is released as open source under the [Apache License 2.0][].

[Apache License 2.0]: http://www.apache.org/licenses/LICENSE-2.0.html


The development of Magnacarto is sponsored by [Omniscale](http://omniscale.com/) and development will continue as resources permit.
Please get in touch if you need commercial support or if you need specific features.


Features
--------

* Easy to use map viewer

![Magnaserv](./docs/magnaserv.png)


* Generate styles for Mapnik 2/3 and MapServer

#### MapServer
![OSM-Bright MapServer](./docs/osm-bright-mapserver.png)

#### Mapnik

![OSM-Bright Mapnik](./docs/osm-bright-mapnik.png)


Current status
--------------

- Supports nearly all features of CartoCSS
  - Attachments
  - Instances
  - Classes
  - Color functions
  - Expressions
  - etc.
- Can successfully convert complex styles (like the OSM Carto style)

### Missing ###

- Regexp filters
- Not all CartoCSS features are supported by the MapServer builder
- Improved configuration
- Support for new Mapnik 3 features
- ...

Installation
------------

### Binary

There are binary releases available for Windows, Linux and Mac OS X (Darwin): <http://download.omniscale.de/magnacarto/rel/>

### Source

#### Dependencies

You need [Go][]>1.5 and [git][].

[Go]: https://golang.org
[git]: https://git-scm.com/

#### Compiling

First create a `GOPATH` directory for all your Go code, if you don't have one already:

    mkdir -p go
    cd go
    export GOPATH=`pwd`

Then you need to enable GO15VENDOREXPERIMENT, if you are using Go 1.5. You can skip this if you are using 1.6 or higher:

    export GO15VENDOREXPERIMENT=1

Next you can fetch the source code and build/install it:

    go get -u github.com/omniscale/magnacarto
    cd $GOPATH/src/github.com/omniscale/magnacarto
    make install


#### Render Plugins

The web-frontend of Magnacarto can render map images with Mapserver and Mapnik.

##### Mapserver

The Mapserver plugin is already included in the default `magnaserv` installation and it requires the `mapserv` binary in your `PATH` on runtime.


##### Mapnik

The Mapnik plugin needs to be compiled as an additional binary (`magnacarto-mapnik`). You need to have Mapnik installed with all header files. It supports Mapnik 2.2 and 3.0. Make sure `mapnik-config` is in your `PATH`. Call `make install` to build the plugin binary.

Usage
-----

### magnacarto

`magnacarto` takes a single `-mml` file.

    magnacarto -mml project.mml > /tmp/magnacarto.xml

To build a MapServer map file:

    magnacarto -builder mapserver -mml project.mml > /tmp/magnacarto.map

See `magnacarto -help` for more options.

### magnaserv


Magnaserv is a web-frontend for Magnacarto. Make sure the `./app` directory is in your working directory or next to the `magnaserv` executable.


To start magnaserv on port 7070:

    magnaserv

Magnaserv will search for .mml files in the current working directory or in direct sub-directories.

To start magnaserv on port 8080 with the Mapserver plugin enabled:

    magnaserv -builder mapserver -listen 127.0.0.1:8080

You can configure the location of stylings, shapefiles or images, and database connection parameters with a configuration file. See `example-magnacarto.tml`:

    magnaserv -builder mapserver -config magnacarto.tml


Documentation
-------------

See [docs/examples](https://github.com/omniscale/magnacarto/tree/master/docs/examples) for example files and usage instructions.

Refer to the [Carto project](https://github.com/mapbox/carto) for documentation of the CartoCSS format.

- https://github.com/mapbox/carto/blob/master/docs/latest.md
- https://www.mapbox.com/tilemill/docs/crashcourse/styling/

Refer to the following CartoCSS projects for larger .mml and .mss examples.

- https://github.com/mapbox/osm-bright
- https://github.com/gravitystorm/openstreetmap-carto

Please note that openstreetmap-carto relies on a few advanced Mapnik features that are not supported by Mapserver. Future versions of Magnacarto might work around these limitations.


Support
-------

Please use GitHub for questions: <https://github.com/omniscale/magnacarto/issues>

For commercial support [contact Omniscale](http://omniscale.com/contact).

Development
-----------

The latest developer documentation can be found here: <http://godoc.org/github.com/omniscale/magnacarto>

The source code is available at: <https://github.com/omniscale/magnacarto/>

You can report any issues at: <https://github.com/omniscale/magnacarto/issues>

### Test ###

#### Unit tests ####

    go test -short ./...


#### Regression tests ####

There are regression tests that generate Mapnik and MapServer map files, renders images and compares them.
These tests require Image Magick (`compare`) and MapServer >=7 (`shp2img`).

    go test ./...
