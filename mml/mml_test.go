package mml

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	r, err := os.Open("tests/001-json.mml")
	assert.NoError(t, err)
	mml, err := Parse(r)
	assert.NoError(t, err)
	r.Close()

	assert.Equal(t, mml.Name, "JSON MML")
	assert.Equal(t, mml.Stylesheets[0], "style.mss")
	assert.Equal(t, mml.Layers[0].ID, "testlayer")
	assert.Equal(t, mml.Layers[0].Classes[0], "testclass")
	assert.Equal(t, mml.Layers[0].SRS, "+proj=merc +a=6378137 +b=6378137 +lat_ts=0.0 +lon_0=0.0 +x_0=0.0 +y_0=0.0 +k=1.0 +units=m +nadgrids=@null +wktext +no_defs +over")
	assert.Equal(t, mml.Layers[0].Type, "Polygon")
	ds := mml.Layers[0].Datasource.(Shapefile)
	assert.Equal(t, ds.Filename, "test.shp")

	r, err = os.Open("tests/002-yaml.mml")
	assert.NoError(t, err)
	mml, err = Parse(r)
	assert.NoError(t, err)
	r.Close()

	assert.Equal(t, mml.Name, "YAML MML")
	assert.Equal(t, mml.Stylesheets[0], "style.mss")
	assert.Equal(t, mml.Layers[0].ID, "testlayer")
	assert.Equal(t, mml.Layers[0].Classes[0], "testclass")
	assert.Equal(t, mml.Layers[0].SRS, "+proj=merc +a=6378137 +b=6378137 +lat_ts=0.0 +lon_0=0.0 +x_0=0.0 +y_0=0.0 +k=1.0 +units=m +nadgrids=@null +wktext +no_defs +over")
	assert.Equal(t, mml.Layers[0].Type, "Polygon")
	ds = mml.Layers[0].Datasource.(Shapefile)
	assert.Equal(t, ds.Filename, "test.shp")
}
