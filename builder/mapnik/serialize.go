// Package mapserver builds Mapnik .xml files.
package mapnik

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/omniscale/magnacarto/builder"
	"github.com/omniscale/magnacarto/builder/sql"
	"github.com/omniscale/magnacarto/color"
	"github.com/omniscale/magnacarto/config"
	"github.com/omniscale/magnacarto/mml"
	"github.com/omniscale/magnacarto/mss"
)

type Map struct {
	fontSets       map[string]string
	XML            *XMLMap
	locator        config.Locator
	autoTypeFilter bool
	mapnik2        bool
}

type maker struct {
	mapnik2 bool
}

func (m maker) Type() string       { return "mapnik" }
func (m maker) FileSuffix() string { return ".xml" }
func (m maker) New(locator config.Locator) builder.MapWriter {
	mm := New(locator)
	if m.mapnik2 {
		mm.SetMapnik2(true)
	}
	return mm
}

var Maker2 = maker{mapnik2: true}
var Maker3 = maker{}

func New(locator config.Locator) *Map {
	return &Map{
		fontSets: make(map[string]string),
		XML:      &XMLMap{SRS: "+init=epsg:3857"},
		locator:  locator,
	}
}

func (m *Map) SetAutoTypeFilter(enable bool) {
	m.autoTypeFilter = enable
}

func (m *Map) SetBackgroundColor(c color.Color) {
	m.XML.BgColor = fmtColor(c, true)
}

func (m *Map) SetMapnik2(enable bool) {
	m.mapnik2 = enable
}

func (m *Map) AddLayer(l mml.Layer, rules []mss.Rule) {
	styles := m.newStyles(rules)
	m.XML.Styles = append(m.XML.Styles, styles...)

	layer := Layer{}
	layer.SRS = &l.SRS
	layer.Name = l.ID
	if !l.Active {
		layer.Status = "off"
	}
	if l.GroupBy != "" {
		layer.GroupBy = l.GroupBy
	}
	z := mss.RulesZoom(rules)
	if z != mss.AllZoom {
		if l := z.First(); l > 0 {
			if m.mapnik2 {
				layer.MaxZoom = zoomRanges[l]
			} else {
				layer.MaxScaleDenom = zoomRanges[l]
			}
		}
		if l := z.Last(); l < 22 {
			if m.mapnik2 {
				layer.MinZoom = zoomRanges[l+1]
			} else {
				layer.MinScaleDenom = zoomRanges[l+1]
			}
		}
	}
	params := m.newDatasource(l.Datasource, rules)
	if params != nil {
		layer.Datasource = &params
	}
	for _, s := range styles {
		layer.StyleNames = append(layer.StyleNames, s.Name)
	}
	m.XML.Layers = append(m.XML.Layers, layer)
}

func (m *Map) Write(w io.Writer) error {
	e := xml.NewEncoder(w)
	e.Indent("", "  ")
	err := e.Encode(m.XML)
	return err
}

func (m *Map) WriteFiles(basename string) error {
	f, err := os.Create(basename)
	if err != nil {
		return err
	}
	defer f.Close()
	return m.Write(f)
}

func (m *Map) newDatasource(ds mml.Datasource, rules []mss.Rule) []Parameter {
	var params []Parameter
	switch ds := ds.(type) {
	case mml.PostGIS:
		ds = m.locator.PostGIS(ds)
		params = []Parameter{
			{Name: "host", Value: ds.Host},
			{Name: "port", Value: ds.Port},
			{Name: "geometry_field", Value: ds.GeometryField},
			{Name: "dbname", Value: ds.Database},
			{Name: "user", Value: ds.Username},
			{Name: "password", Value: ds.Password},
			{Name: "extent", Value: ds.Extent},
			{Name: "table", Value: pqSelectString(ds.Query, rules, m.autoTypeFilter)},
			{Name: "srid", Value: ds.SRID},
			{Name: "type", Value: "postgis"},
		}
	case mml.Shapefile:
		fname := m.locator.Shape(ds.Filename)
		params = []Parameter{
			{Name: "file", Value: fname},
			{Name: "type", Value: "shape"},
		}
	case mml.SQLite:
		fname := m.locator.SQLite(ds.Filename)
		params = []Parameter{
			{Name: "file", Value: fname},
			{Name: "srid", Value: ds.SRID},
			{Name: "extent", Value: ds.Extent},
			{Name: "geometry_field", Value: ds.GeometryField},
			{Name: "table", Value: ds.Query},
			{Name: "type", Value: "sqlite"},
		}
	case mml.OGR:
		params = []Parameter{
			{Name: "file", Value: ds.Filename},
			{Name: "srid", Value: ds.SRID},
			{Name: "extent", Value: ds.Extent},
			{Name: "layer", Value: ds.Layer},
			{Name: "type", Value: "ogr"},
		}
	case mml.GDAL:
		params = []Parameter{
			{Name: "file", Value: ds.Filename},
			{Name: "srid", Value: ds.SRID},
			{Name: "extent", Value: ds.Extent},
			{Name: "band", Value: ds.Band},
			{Name: "type", Value: "gdal"},
		}
	case nil:
		// datasource might be nil for exports withour mml
	default:
		panic(fmt.Sprintf("datasource not supported by Mapnik: %v", ds))
	}

	// drop empty parameters
	var result []Parameter
	for _, p := range params {
		if p.Value != "" {
			result = append(result, p)
		}
	}
	return result
}

func pqSelectString(query string, rules []mss.Rule, autoTypeFilter bool) string {
	if !autoTypeFilter {
		return query
	}
	filter := sql.FilterString(rules)
	return sql.WrapWhere(query, filter)
}

func (m *Map) newStyles(rules []mss.Rule) []Style {
	styles := []Style{}
	style := Style{FilterMode: "first"}

	for _, r := range rules {
		mr := m.newRule(r)

		styleName := r.Layer
		if r.Attachment != "" {
			styleName += "-" + r.Attachment
		}

		if style.Name != styleName {
			if len(style.Rules) > 0 {
				styles = append(styles, style)
			}
			style = Style{Name: styleName, FilterMode: "first"}
			// apply style-level properties
			for _, rr := range rules {
				if r.Attachment == rr.Attachment {
					if v, ok := r.Properties.GetString("comp-op"); ok {
						style.CompOp = &v
					}
					if v, ok := r.Properties.GetFloat("opacity"); ok {
						style.Opacity = &v
					}
				}
			}
		}
		style.Rules = append(style.Rules, *mr)
	}
	if len(style.Rules) > 0 {
		styles = append(styles, style)
	}

	return styles
}

func (m *Map) newRule(r mss.Rule) *Rule {
	result := &Rule{}

	if r.Zoom != mss.AllZoom {
		result.Zoom = r.Zoom.String()
	}
	if l := r.Zoom.First(); l > 0 {
		result.MaxScaleDenom = zoomRanges[l]
	}
	if l := r.Zoom.Last(); l < 22 {
		result.MinScaleDenom = zoomRanges[l+1]
	}

	result.Filter = fmtFilters(r.Filters)
	prefixes := mss.SortedPrefixes(r.Properties, []string{"line-", "polygon-", "polygon-pattern-", "text-", "shield-", "marker-", "point-", "building-", "raster-"})

	for _, p := range prefixes {
		r.Properties.SetDefaultInstance(p.Instance)
		switch p.Name {
		case "line-":
			m.addLineSymbolizer(result, r)
		case "polygon-":
			m.addPolygonSymbolizer(result, r)
		case "polygon-pattern-":
			m.addPolygonPatternSymbolizer(result, r)
		case "text-":
			m.addTextSymbolizer(result, r)
		case "shield-":
			m.addShieldSymbolizer(result, r)
		case "marker-":
			m.addMarkerSymbolizer(result, r)
		case "point-":
			m.addPointSymbolizer(result, r)
		case "building-":
			m.addBuildingSymbolizer(result, r)
		case "raster-":
			m.addRasterSymbolizer(result, r)
		default:
			log.Println("invalid prefix", p)
		}
	}
	r.Properties.SetDefaultInstance("")
	return result
}

func (m *Map) addLineSymbolizer(result *Rule, r mss.Rule) {
	if width, ok := r.Properties.GetFloat("line-width"); ok {
		symb := LineSymbolizer{}
		symb.Width = fmtFloat(width, true)
		symb.Clip = fmtBool(r.Properties.GetBool("line-clip"))
		symb.Color = fmtColor(r.Properties.GetColor("line-color"))
		symb.Dasharray = fmtPattern(r.Properties.GetFloatList("line-dasharray"))
		symb.Gamma = fmtFloat(r.Properties.GetFloat("line-gamma"))
		symb.GammaMethod = fmtString(r.Properties.GetString("line-gamma-method"))
		symb.Linecap = fmtString(r.Properties.GetString("line-cap"))
		symb.Miterlimit = fmtFloat(r.Properties.GetFloat("line-miterlimit"))
		symb.Linejoin = fmtString(r.Properties.GetString("line-join"))
		symb.Offset = fmtFloat(r.Properties.GetFloat("line-offset"))
		symb.Opacity = fmtFloat(r.Properties.GetFloat("line-opacity"))
		symb.Rasterizer = fmtString(r.Properties.GetString("line-rasterizer"))
		symb.Simplify = fmtFloat(r.Properties.GetFloat("line-simplify"))
		symb.SimplifyAlgorithm = fmtString(r.Properties.GetString("line-simplify-algorithm"))
		symb.Smooth = fmtFloat(r.Properties.GetFloat("line-smooth"))
		symb.CompOp = fmtString(r.Properties.GetString("line-comp-op"))
		symb.GeometryTransform = fmtString(r.Properties.GetString("line-geometry-transform"))
		result.Symbolizers = append(result.Symbolizers, &symb)
	}
}

func (m *Map) addLinePatternSymbolizer(result *Rule, r mss.Rule) {
	if patFile, ok := r.Properties.GetString("line-pattern-file"); ok {
		symb := LinePatternSymbolizer{}
		fname := m.locator.Image(patFile)
		symb.File = &fname
		symb.Offset = fmtFloat(r.Properties.GetFloat("line-pattern-offset"))
		symb.Opacity = fmtFloat(r.Properties.GetFloat("line-pattern-opacity"))
		symb.Clip = fmtBool(r.Properties.GetBool("line-pattern-clip"))
		symb.Simplify = fmtFloat(r.Properties.GetFloat("line-pattern-simplify"))
		symb.SimplifyAlgorithm = fmtString(r.Properties.GetString("line-pattern-simplify-algorithm"))
		symb.Smooth = fmtFloat(r.Properties.GetFloat("line-pattern-smooth"))
		symb.GeometryTransform = fmtString(r.Properties.GetString("line-pattern-geometry-transform"))
		symb.CompOp = fmtString(r.Properties.GetString("line-pattern-comp-op"))
		result.Symbolizers = append(result.Symbolizers, &symb)
	}
}

func (m *Map) addPolygonSymbolizer(result *Rule, r mss.Rule) {
	if fill, ok := r.Properties.GetColor("polygon-fill"); ok {
		symb := PolygonSymbolizer{}
		symb.Color = fmtColor(fill, true)
		symb.Opacity = fmtFloat(r.Properties.GetFloat("polygon-opacity"))
		symb.Gamma = fmtFloat(r.Properties.GetFloat("polygon-gamma"))
		symb.GammaMethod = fmtString(r.Properties.GetString("polygon-gamma-method"))
		symb.Clip = fmtBool(r.Properties.GetBool("polygon-clip"))
		symb.Simplify = fmtFloat(r.Properties.GetFloat("polygon-simplify"))
		symb.SimplifyAlgorithm = fmtString(r.Properties.GetString("polygon-simplify-algorithm"))
		symb.Smooth = fmtFloat(r.Properties.GetFloat("polygon-smooth"))
		symb.GeometryTransform = fmtString(r.Properties.GetString("polygon-geometry-transform"))
		symb.CompOp = fmtString(r.Properties.GetString("polygon-comp-op"))

		result.Symbolizers = append(result.Symbolizers, &symb)
	}
}

func (m *Map) addTextSymbolizer(result *Rule, r mss.Rule) {
	if size, ok := r.Properties.GetFloat("text-size"); ok {
		symb := TextSymbolizer{}
		symb.Size = fmtFloat(size, true)
		symb.Fill = fmtColor(r.Properties.GetColor("text-fill"))
		symb.Name = fmtField(r.Properties.GetFieldList("text-name"))
		symb.HaloFill = fmtColor(r.Properties.GetColor("text-halo-fill"))
		symb.HaloRadius = fmtFloat(r.Properties.GetFloat("text-halo-radius"))
		symb.HaloOpacity = fmtFloat(r.Properties.GetFloat("text-halo-opacity"))
		symb.HaloRasterizer = fmtString(r.Properties.GetString("text-halo-rasterizer"))
		symb.HaloTransform = fmtString(r.Properties.GetString("text-halo-transform"))
		symb.HaloCompOp = fmtString(r.Properties.GetString("text-halo-comp-op"))
		symb.Opacity = fmtFloat(r.Properties.GetFloat("text-opacity"))
		symb.WrapCharacter = fmtString(r.Properties.GetString("text-wrap-character"))
		symb.WrapBefore = fmtString(r.Properties.GetString("text-wrap-before"))
		symb.WrapWidth = fmtFloat(r.Properties.GetFloat("text-wrap-width"))
		symb.RepeatWrapCharacter = fmtBool(r.Properties.GetBool("text-repeat-wrap-characater"))
		symb.Ratio = fmtFloat(r.Properties.GetFloat("text-ratio"))
		symb.MaxCharAngleDelta = fmtFloat(r.Properties.GetFloat("text-max-char-angle-delta"))

		symb.Placement = fmtString(r.Properties.GetString("text-placement"))
		symb.PlacementType = fmtString(r.Properties.GetString("text-placement-type"))
		symb.Placements = fmtString(r.Properties.GetString("text-placements"))
		symb.LabelPositionTolerance = fmtFloat(r.Properties.GetFloat("text-label-position-tolerance"))
		symb.VerticalAlign = fmtString(r.Properties.GetString("text-vertical-alignment"))
		symb.HorizontalAlign = fmtString(r.Properties.GetString("text-horizontal-alignment"))
		symb.JustifyAlign = fmtString(r.Properties.GetString("text-justify-alignment"))
		symb.Margin = fmtFloat(r.Properties.GetFloat("text-margin"))
		symb.Simplify = fmtFloat(r.Properties.GetFloat("text-simplify"))
		symb.SimplifyAlgorithm = fmtString(r.Properties.GetString("text-simplify-algorithm"))
		symb.Smooth = fmtFloat(r.Properties.GetFloat("text-smooth"))
		symb.CompOp = fmtString(r.Properties.GetString("text-comp-op"))

		symb.Dx = fmtFloat(r.Properties.GetFloat("text-dx"))
		symb.Dy = fmtFloat(r.Properties.GetFloat("text-dy"))

		if v, ok := r.Properties.GetFloat("text-orientation"); ok {
			symb.Orientation = fmtFloat(v, true)
		} else if v, ok := r.Properties.GetFieldList("text-orientation"); ok {
			symb.Orientation = fmtField(v, true)
		}
		symb.RotateDisplacement = fmtBool(r.Properties.GetBool("text-rotate-displacement"))
		symb.Upright = fmtString(r.Properties.GetString("text-upgright"))

		symb.CharacterSpacing = fmtFloat(r.Properties.GetFloat("text-character-spacing"))
		symb.LineSpacing = fmtFloat(r.Properties.GetFloat("text-line-spacing"))

		symb.AllowOverlap = fmtBool(r.Properties.GetBool("text-allow-overlap"))

		symb.LargestBboxOnly = fmtBool(r.Properties.GetBool("text-largest-bbox-only"))

		// TODO see for issue/upcoming fixes with 3.0 https://github.com/mapnik/mapnik/issues/2362

		// spacing between repeated labels
		symb.Spacing = fmtFloat(r.Properties.GetFloat("text-spacing"))
		// min-distance to other label, does not work with placement-line
		symb.MinimumDistance = fmtFloat(r.Properties.GetFloat("text-min-distance"))
		symb.RepeatDistance = fmtFloat(r.Properties.GetFloat("text-repeat-distance"))
		// min-padding to map edge
		symb.MinimumPadding = fmtFloat(r.Properties.GetFloat("text-min-padding"))
		symb.MinPathLength = fmtFloat(r.Properties.GetFloat("text-min-path-length"))

		symb.Clip = fmtBool(r.Properties.GetBool("text-clip"))
		symb.TextTransform = fmtString(r.Properties.GetString("text-transform"))

		symb.FontFeatureSettings = fmtString(r.Properties.GetString("font-feature-settings"))
		if faceNames, ok := r.Properties.GetStringList("text-face-name"); ok {
			symb.FontsetName = m.fontSetName(faceNames)
		}

		if symb.Name != nil && *symb.Name != "" {
			result.Symbolizers = append(result.Symbolizers, &symb)
		}
	}
}

func (m *Map) addShieldSymbolizer(result *Rule, r mss.Rule) {
	if shieldFile, ok := r.Properties.GetString("shield-file"); ok {
		symb := ShieldSymbolizer{}

		fname := m.locator.Image(shieldFile)
		symb.File = &fname

		symb.Size = fmtFloat(r.Properties.GetFloat("shield-size"))
		symb.Fill = fmtColor(r.Properties.GetColor("shield-fill"))
		symb.Name = fmtField(r.Properties.GetFieldList("shield-name"))
		symb.TextOpacity = fmtFloat(r.Properties.GetFloat("shield-text-opacity"))
		symb.Opacity = fmtFloat(r.Properties.GetFloat("shield-opacity"))
		symb.Transform = fmtString(r.Properties.GetString("shield-transform"))
		symb.Simplify = fmtFloat(r.Properties.GetFloat("shield-simplify"))
		symb.SimplifyAlgorithm = fmtString(r.Properties.GetString("shield-simplify-algorithm"))
		symb.Smooth = fmtFloat(r.Properties.GetFloat("shield-smooth"))
		symb.CompOp = fmtString(r.Properties.GetString("shield-comp-op"))

		symb.Placement = fmtString(r.Properties.GetString("shield-placement"))
		symb.PlacementType = fmtString(r.Properties.GetString("shield-placement-type"))
		symb.Placements = fmtString(r.Properties.GetString("shield-placements"))
		symb.UnlockImage = fmtBool(r.Properties.GetBool("shield-unlock-image"))
		symb.HorizontalAlign = fmtString(r.Properties.GetString("shield-horizontal-alignment"))
		symb.VerticalAlign = fmtString(r.Properties.GetString("shield-vertical-alignment"))
		symb.JustifyAlign = fmtString(r.Properties.GetString("shield-justify-alignment"))

		symb.Clip = fmtBool(r.Properties.GetBool("shield-clip"))
		symb.AllowOverlap = fmtBool(r.Properties.GetBool("shield-allow-overlap"))
		symb.AvoidEdges = fmtBool(r.Properties.GetBool("shield-avoid-edges"))

		symb.HaloFill = fmtColor(r.Properties.GetColor("shield-halo-fill"))
		symb.HaloRadius = fmtFloat(r.Properties.GetFloat("shield-halo-radius"))
		symb.HaloRasterizer = fmtString(r.Properties.GetString("shield-halo-rasterizer"))
		symb.HaloTransform = fmtString(r.Properties.GetString("shield-halo-transform"))
		symb.HaloCompOp = fmtString(r.Properties.GetString("shield-halo-comp-op"))
		symb.HaloOpacity = fmtFloat(r.Properties.GetFloat("shield-halo-opacity"))

		symb.CharacterSpacing = fmtFloat(r.Properties.GetFloat("shield-character-spacing"))
		symb.WrapCharacter = fmtString(r.Properties.GetString("shield-wrap-character"))
		symb.WrapBefore = fmtBool(r.Properties.GetBool("shield-wrap-before"))
		symb.WrapWidth = fmtFloat(r.Properties.GetFloat("shield-wrap-width"))
		symb.LineSpacing = fmtFloat(r.Properties.GetFloat("shield-line-spacing"))
		symb.Dx = fmtFloat(r.Properties.GetFloat("shield-dx"))
		symb.Dy = fmtFloat(r.Properties.GetFloat("shield-dx"))
		symb.TextDx = fmtFloat(r.Properties.GetFloat("shield-text-dx"))
		symb.TextDy = fmtFloat(r.Properties.GetFloat("shield-text-dy"))
		symb.LabelPositionTolerance = fmtFloat(r.Properties.GetFloat("shield-label-position-tolerance"))
		symb.TextTransform = fmtString(r.Properties.GetString("shield-text-transform"))

		symb.Spacing = fmtFloat(r.Properties.GetFloat("shield-spacing"))
		symb.MinimumDistance = fmtFloat(r.Properties.GetFloat("shield-min-distance"))
		symb.MinimumPadding = fmtFloat(r.Properties.GetFloat("shield-min-padding"))
		symb.Margin = fmtFloat(r.Properties.GetFloat("shield-margin"))
		symb.RepeatDistance = fmtFloat(r.Properties.GetFloat("shield-repeat-distance"))

		if faceNames, ok := r.Properties.GetStringList("shield-face-name"); ok {
			symb.FontsetName = m.fontSetName(faceNames)
		}

		result.Symbolizers = append(result.Symbolizers, &symb)
	}
}

func (m *Map) addMarkerSymbolizer(result *Rule, r mss.Rule) {
	symb := MarkersSymbolizer{}
	symb.Width = fmtFloat(r.Properties.GetFloat("marker-width"))
	symb.Height = fmtFloat(r.Properties.GetFloat("marker-height"))
	symb.Fill = fmtColor(r.Properties.GetColor("marker-fill"))
	symb.FillOpacity = fmtFloat(r.Properties.GetFloat("marker-fill-opacity"))
	symb.Opacity = fmtFloat(r.Properties.GetFloat("marker-opacity"))
	symb.Placement = fmtString(r.Properties.GetString("marker-placement"))
	symb.Transform = fmtString(r.Properties.GetString("marker-transform"))
	symb.GeometryTransform = fmtString(r.Properties.GetString("marker-geometry-transform"))
	symb.Spacing = fmtFloat(r.Properties.GetFloat("marker-spacing"))
	symb.Stroke = fmtColor(r.Properties.GetColor("marker-line-color"))
	symb.StrokeOpacity = fmtFloat(r.Properties.GetFloat("marker-line-opacity"))
	symb.StrokeWidth = fmtFloat(r.Properties.GetFloat("marker-line-width"))
	symb.AllowOverlap = fmtBool(r.Properties.GetBool("marker-allow-overlap"))
	symb.MultiPolicy = fmtString(r.Properties.GetString("marker-multi-policy"))
	symb.AvoidEdges = fmtBool(r.Properties.GetBool("marker-avoid-edges"))
	symb.IgnorePlacement = fmtBool(r.Properties.GetBool("marker-ignore-placement"))
	symb.MaxError = fmtFloat(r.Properties.GetFloat("marker-max-error"))
	symb.Clip = fmtBool(r.Properties.GetBool("marker-clip"))
	symb.Simplify = fmtFloat(r.Properties.GetFloat("marker-simplify"))
	symb.SimplifyAlgorithm = fmtString(r.Properties.GetString("marker-simplify-algorithm"))
	symb.Smooth = fmtFloat(r.Properties.GetFloat("marker-smooth"))
	symb.Offset = fmtFloat(r.Properties.GetFloat("marker-offset"))
	symb.CompOp = fmtString(r.Properties.GetString("marker-comp-op"))
	symb.Direction = fmtString(r.Properties.GetString("marker-direction"))

	if markerFile, ok := r.Properties.GetString("marker-file"); ok {
		fname := m.locator.Image(markerFile)
		symb.File = &fname

	} else {
		// carto uses 'ellipse' as default for "marker-type"
		markerType, ok := r.Properties.GetString("marker-type")
		if !ok {
			markerType = "ellipse"
			// default marker type requires at least fill, stroke or strokewidth
			if symb.Fill == nil && symb.Stroke == nil && symb.StrokeWidth == nil {
				return
			}
		}
		symb.MarkerType = &markerType
	}
	result.Symbolizers = append(result.Symbolizers, &symb)
}

func (m *Map) addPointSymbolizer(result *Rule, r mss.Rule) {
	if pointFile, ok := r.Properties.GetString("point-file"); ok {
		symb := PointSymbolizer{}
		fname := m.locator.Image(pointFile)
		symb.File = &fname
		symb.AllowOverlap = fmtBool(r.Properties.GetBool("point-allow-overlap"))
		symb.Opacity = fmtFloat(r.Properties.GetFloat("point-opacity"))
		symb.Transform = fmtString(r.Properties.GetString("point-transform"))
		symb.IgnorePlacement = fmtBool(r.Properties.GetBool("point-ignore-placement"))
		symb.Placement = fmtString(r.Properties.GetString("point-placement"))
		symb.CompOp = fmtString(r.Properties.GetString("point-comp-op"))
		result.Symbolizers = append(result.Symbolizers, &symb)
	}
}

func (m *Map) addPolygonPatternSymbolizer(result *Rule, r mss.Rule) {
	if patFile, ok := r.Properties.GetString("polygon-pattern-file"); ok {
		symb := PolygonPatternSymbolizer{}
		fname := m.locator.Image(patFile)
		symb.File = &fname
		symb.Alignment = fmtString(r.Properties.GetString("polygon-pattern-alignment"))
		symb.Gamma = fmtFloat(r.Properties.GetFloat("polygon-pattern-gamma"))
		symb.Opacity = fmtFloat(r.Properties.GetFloat("polygon-pattern-opacity"))
		symb.Clip = fmtBool(r.Properties.GetBool("polygon-pattern-clip"))
		symb.Simplify = fmtFloat(r.Properties.GetFloat("polygon-pattern-simplify"))
		symb.SimplifyAlgorithm = fmtString(r.Properties.GetString("polygon-pattern-simplify-algorithm"))
		symb.Smooth = fmtFloat(r.Properties.GetFloat("polygon-pattern-smooth"))
		symb.GeometryTransform = fmtString(r.Properties.GetString("polygon-pattern-geometry-transform"))
		symb.CompOp = fmtString(r.Properties.GetString("polygon-pattern-comp-op"))
		result.Symbolizers = append(result.Symbolizers, &symb)
	}
}

func (m *Map) addBuildingSymbolizer(result *Rule, r mss.Rule) {
	if fill, ok := r.Properties.GetColor("building-fill"); ok {
		symb := BuildingSymbolizer{}
		symb.Fill = fmtColor(fill, true)
		symb.FillOpacity = fmtFloat(r.Properties.GetFloat("building-fill-opacity"))
		symb.Height = fmtFloat(r.Properties.GetFloat("building-height"))
		result.Symbolizers = append(result.Symbolizers, &symb)
	}
}

func (m *Map) addRasterSymbolizer(result *Rule, r mss.Rule) {
	opacity, ok := r.Properties.GetFloat("raster-opacity")
	if ok && opacity == 0.0 {
		return
	}
	symb := RasterSymbolizer{}
	if stops, ok := r.Properties.GetStopList("raster-colorizer-stops"); ok {
		for _, stop := range stops {
			symb.Stops = append(symb.Stops,
				Stop{
					Value: strconv.FormatInt(int64(stop.Value), 10),
					Color: *fmtColor(stop.Color, true),
				},
			)
		}
	}
	symb.Opacity = fmtFloat(r.Properties.GetFloat("raster-opacity"))
	symb.Epsilon = fmtFloat(r.Properties.GetFloat("raster-colorizer-epsilon"))
	symb.MeshSize = fmtFloat(r.Properties.GetFloat("raster-mesh-size"))
	symb.FilterFactor = fmtFloat(r.Properties.GetFloat("raster-filter-factor"))
	symb.CompOp = fmtString(r.Properties.GetString("raster-comp-op"))
	symb.Scaling = fmtString(r.Properties.GetString("raster-scaling"))
	symb.DefaultMode = fmtString(r.Properties.GetString("raster-colorizer-default-mode"))
	symb.DefaultColor = fmtColor(r.Properties.GetColor("raster-colorizer-default-color"))
	result.Symbolizers = append(result.Symbolizers, &symb)
}

func (m *Map) fontSetName(fontFaces []string) *string {
	str := fmt.Sprint(fontFaces)

	if fontSetName, ok := m.fontSets[str]; ok {
		return &fontSetName
	}
	fontSet := FontSet{}
	fontSet.Name = fmt.Sprintf("fontset-%d", len(m.fontSets)+1)
	m.fontSets[str] = fontSet.Name // cache fontsetname for this font combination
	for _, f := range fontFaces {
		fontSet.Fonts = append(fontSet.Fonts, Font{FaceName: f})
	}
	m.XML.FontSets = append(m.XML.FontSets, fontSet)
	return &fontSet.Name
}

func fmtField(vals []interface{}, ok bool) *string {
	if !ok {
		return nil
	}
	parts := []string{}
	for _, v := range vals {
		switch v.(type) {
		case mss.Field:
			parts = append(parts, string(v.(mss.Field)))
		case string:
			parts = append(parts, "'"+v.(string)+"'")
		}
	}
	r := strings.Join(parts, " + ")
	return &r
}

func fmtPattern(v []float64, ok bool) *string {
	if !ok {
		return nil
	}
	parts := make([]string, 0, len(v))
	for i := range v {
		parts = append(parts, strconv.FormatFloat(v[i], 'f', -1, 64))

	}
	r := strings.Join(parts, ", ")
	return &r
}

func fmtFloat(v float64, ok bool) *string {
	if !ok {
		return nil
	}
	r := strconv.FormatFloat(v, 'f', -1, 64)
	return &r
}

func fmtString(v string, ok bool) *string {
	if !ok {
		return nil
	}
	return &v
}

func fmtBool(v bool, ok bool) *string {
	if !ok {
		return nil
	}
	var r string
	if v {
		r = "true"
	} else {
		r = "false"
	}
	return &r
}

func fmtColor(v color.Color, ok bool) *string {
	if !ok {
		return nil
	}
	r := v.String()
	return &r
}

func fmtFilters(filters []mss.Filter) string {
	parts := []string{}
	for _, f := range filters {
		var value string
		switch v := f.Value.(type) {
		case nil:
			value = "null"
		case string:
			// TODO quote " in string?!
			value = `'` + v + `'`
		case float64:
			value = string(*fmtFloat(v, true))
		default:
			log.Printf("unknown type of filter value: %s", v)
			value = ""
		}
		field := f.Field
		if len(field) > 2 && field[0] == '"' && field[len(field)-1] == '"' {
			// strip quotes from field name
			field = field[1 : len(field)-1]
		}
		parts = append(parts, "(["+field+"] "+f.CompOp.String()+" "+value+")")
	}

	s := strings.Join(parts, " and ")
	if len(filters) > 1 {
		s = "(" + s + ")"
	}
	return s
}

var zoomRanges = []int64{
	1000000000,
	500000000,
	200000000,
	100000000,
	50000000,
	25000000,
	12500000,
	6500000,
	3000000,
	1500000,
	750000,
	400000,
	200000,
	100000,
	50000,
	25000,
	12500,
	5000,
	2500,
	1500,
	750,
	500,
	250,
	100,
}
