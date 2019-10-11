go-mapnik
=========

Description
-----------

Small wrapper for the Mapnik 3 API to render beautiful maps from Go.

Features:

* Render to `[]byte`, `image.Image`, or file.
* Set scale denominator or scale factor.
* Enable/disable single layers.


Installation
------------

This package requires [Mapnik](http://mapnik.org/) (`libmapnik-dev` on Ubuntu/Debian, `mapnik` in Homebrew).
Make sure `mapnik-config` is in your `PATH`.

You need to set the `CGO_LDFLAGS` and `CGO_CXXFLAGS` environment variables for successful compilation and linking with Mapnik.
Refer to the Makefile how `mapnik-config` can be used to extract the required `CGO_LDFLAGS` and `CGO_CXXFLAGS` values. Use `-ldflags` to overwrite the default location of the input plugins and default fonts.


Documentation
-------------

API documentation can be found here: <http://godoc.org/github.com/omniscale/go-mapnik>


License
-------

MIT, see LICENSE file.

Author
------

Oliver Tonnhofer, [Omniscale](http://maps.omniscale.com)

Thanks
------

This package is inspired/based on [`mapnik-c-api`](https://github.com/springmeyer/mapnik-c-api) by Dane Springmeyer and [`go-mapnik`](https://github.com/fawick/go-mapnik) by Fabian Wickborn.
