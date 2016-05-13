package mapserver

import (
	"testing"

	"github.com/omniscale/magnacarto/color"

	"github.com/omniscale/magnacarto/config"
	"github.com/omniscale/magnacarto/mml"
	"github.com/omniscale/magnacarto/mss"

	"github.com/stretchr/testify/assert"
)

var locator = config.LookupLocator{}

func TestNoLayers(t *testing.T) {
	m := New(&locator)
	m.SetNoMapBlock(true)

	assert.Empty(t, m.String())
}

func TestLineStringLayer(t *testing.T) {
	m := New(&locator)
	m.SetNoMapBlock(true)

	m.AddLayer(mml.Layer{ID: "test", SRS: "4326", Type: mml.LineString},
		[]mss.Rule{
			{Layer: "test", Properties: mss.NewProperties(
				"line-width", 1.0,
				"line-color", color.MustParse("red"),
				"line-opacity", 0.5,
				"line-dasharray", []mss.Value{3.0, 5.0},
			)},
		})
	result := m.String()
	assert.Contains(t, result, "WIDTH 1\n")
	assert.Contains(t, result, "COLOR \"#ff0000\"\n")
	assert.Contains(t, result, "OPACITY 50\n")
	assert.Regexp(t, `PATTERN\s+3\s+5\s+END`, result)
}

func TestScaledLineStringLayer(t *testing.T) {
	m := New(&locator)
	m.SetNoMapBlock(true)

	m.AddLayer(mml.Layer{ID: "test", SRS: "4326", Type: mml.LineString, ScaleFactor: 2.0},
		[]mss.Rule{
			{Layer: "test", Properties: mss.NewProperties(
				"line-width", 3.0,
				"line-opacity", 0.2,
				"line-dasharray", []mss.Value{2.0, 7.0},
			)},
		})
	m.AddLayer(mml.Layer{ID: "test", SRS: "4326", Type: mml.LineString},
		[]mss.Rule{
			{Layer: "test", Properties: mss.NewProperties(
				"line-width", 1.0,
				"line-color", color.MustParse("red"),
				"line-opacity", 0.5,
				"line-dasharray", []mss.Value{3.0, 5.0},
			)},
		})
	result := m.String()
	assert.Contains(t, result, "WIDTH 6\n")
	assert.Contains(t, result, "OPACITY 20\n")
	assert.Regexp(t, `PATTERN\s+4\s+14\s+END`, result)

	assert.Contains(t, result, "WIDTH 1\n")
	assert.Contains(t, result, "COLOR \"#ff0000\"\n")
	assert.Contains(t, result, "OPACITY 50\n")
	assert.Regexp(t, `PATTERN\s+3\s+5\s+END`, result)
}

func TestPolygonLayer(t *testing.T) {
	m := New(&locator)
	m.SetNoMapBlock(true)

	m.AddLayer(mml.Layer{ID: "test", SRS: "4326", Type: mml.Polygon},
		[]mss.Rule{
			{Layer: "test", Properties: mss.NewProperties(
				"line-width", 1.0,
				"line-color", color.MustParse("red"),
				"line-opacity", 0.5,
				"line-dasharray", []mss.Value{3.0, 5.0},
				"polygon-fill", color.MustParse("blue"),
				"polygon-opacity", 0.2,
				"text-size", 10.0,
				"text-name", []mss.Value{mss.Field("name")},
			)},
		})
	result := m.String()
	assert.Contains(t, result, "WIDTH 1\n")
	assert.Contains(t, result, "OUTLINECOLOR \"#ff000080\"\n")
	assert.Regexp(t, `PATTERN\s+3\s+5\s+END`, result)
	assert.Contains(t, result, "COLOR \"#0000ff\"\n")
	assert.Contains(t, result, "OPACITY 20\n")
	assert.Regexp(t, `LABEL\s+ SIZE 7.4\d+`, result)
	assert.Regexp(t, `TEXT 'name'`, result)
}

func TestItem(t *testing.T) {
	assert.Equal(t, `KEY "str"`, Item{"key", quote("str")}.String())
	assert.Equal(t, `"str"`, Item{"", quote("str")}.String())
	assert.Equal(t, `42`, Item{"", 42}.String())
	assert.Equal(t, `KEY 42`, Item{"key", 42}.String())

	assert.Equal(t, `FOO ON`, Item{"foo", "ON"}.String())

	// TODO
	// assert.Equal(t, `"quote\""`, Item{"", "quote\""}.String())
}

func TestBlock(t *testing.T) {
	assert.Equal(t, `KEY "str"`, Block{"", []Item{{"key", quote("str")}}}.String())
	assert.Equal(t,
		`KEY "str"
KEY "str"`,
		Block{"", []Item{{"key", quote("str")}, {"key", quote("str")}}}.String())

	assert.Equal(t,
		`CLASS
  KEY "str"
  KEY "str"
END`,
		Block{"CLASS", []Item{{"key", quote("str")}, {"key", quote("str")}}}.String())

	assert.Equal(t,
		`CLASS
  KEY "str"
  LABEL
    FOO 42
  END
END`,
		Block{"CLASS",
			[]Item{
				{"key", quote("str")},
				{"", Block{
					"label",
					[]Item{
						{"foo", 42},
					},
				}},
			},
		}.String())
}

func TestMarkerTransformWithRotationPoint(t *testing.T) {
	m := New(&locator)
	m.SetNoMapBlock(true)
	m.AddLayer(mml.Layer{ID: "test", SRS: "4326", Type: mml.Point},
		[]mss.Rule{
			{Layer: "test", Properties: mss.NewProperties(
				"marker-file", "/foo/bar.svg",
				"marker-transform", "translate(0.000000, -10.000000) scale(0.500000) rotate(345.000000, 0.000000, 20.000000)",
				"marker-width", 40.0,
				"marker-height", 40.0,
			)},
		})
	result := m.String()

	assert.Contains(t, result, `SYMBOL "anchor-0-5-1-foo-bar-svg"`)
	assert.Contains(t, result, `ANCHORPOINT 0.5 1`)
	assert.Contains(t, result, `ANGLE -345`)
	assert.Contains(t, result, `SIZE 20`)
}
