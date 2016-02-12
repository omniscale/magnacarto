go-mapnik
=========

Description
-----------

Small wrapper for the Mapnik API to render beautiful maps from Go.

Features:

* Render to `[]byte`, `image.Image`, or file.
* Set scale denominator or scale factor.
* Enable/disable single layers.


Installation
------------

This package can be installed with the go get command:

    go get github.com/omniscale/go-mapnik
    go generate github.com/omniscale/go-mapnik

This package requires [Mapnik](http://mapnik.org/) (`libmapnik-dev` on Ubuntu/Debian, `mapnik --with-postgresql` in Homebrew).
Make sure `mapnik-config` is in your `PATH`.

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
