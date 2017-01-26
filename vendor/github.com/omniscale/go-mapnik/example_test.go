package mapnik_test

import (
	"log"

	"github.com/omniscale/go-mapnik"
)

func Example() {
	m := mapnik.New()
	if err := m.Load("test/map.xml"); err != nil {
		log.Fatal(err)
	}
	m.Resize(1000, 500)
	m.ZoomTo(-180, -90, 180, 90)
	opts := mapnik.RenderOpts{Format: "png32"}
	if err := m.RenderToFile(opts, "/tmp/go-mapnik-example.png"); err != nil {
		log.Fatal(err)
	}
	// Output:
}

func ExampleMap_SelectLayers_function() {
	m := mapnik.New()
	if err := m.Load("test/map.xml"); err != nil {
		log.Fatal(err)
	}

	selector := func(layername string) mapnik.Status {
		if layername == "labels" {
			return mapnik.Exclude
		}
		return mapnik.Default
	}
	m.SelectLayers(mapnik.SelectorFunc(selector))
	opts := mapnik.RenderOpts{Format: "png32"}
	if err := m.RenderToFile(opts, "/tmp/go-mapnik-example.png"); err != nil {
		log.Fatal(err)
	}
	m.ResetLayers()
	// Output:
}
