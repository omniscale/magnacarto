go-mapnik
=========

Description
-----------

Small wrapper for the Mapnik 3/4 API to render beautiful maps from Go.

Features:

* Render to `[]byte`, `image.Image`, or file.
* Set scale denominator or scale factor.
* Enable/disable single layers.


Installation
------------

This package requires [Mapnik](http://mapnik.org/) (`libmapnik` on Ubuntu/Debian, `mapnik` in Homebrew).

Since v3 of this package you need to call `go generate .` to create a `build_config.go` file with all system dependent build flags.
Before v3, go-mapnik required setting `CGO_*` environment variables. The new method is easier, but it does not work when you `go get`/import go-mapnik.
It is recommended to manually vendorize the package into your repo (e.g via `git subtree`).

Example
-------

```
func Example() {
	m := mapnik.New()
	if err := m.Load("test/map.xml"); err != nil {
		log.Fatal(err)
	}
	m.Resize(1000, 500)
	m.ZoomTo(-180, -90, 180, 90)
	opts := mapnik.RenderOpts{Format: "png32"}
	if err := m.RenderToFile(opts, "/tmp/go-mapnik-example-1.png"); err != nil {
		log.Fatal(err)
	}
}
```

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
