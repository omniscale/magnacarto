// Package mapserver builds Mapnik .xml files.
package mapnik

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
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
	scaleFactor    float64
	autoTypeFilter bool
	mapnik2        bool
	zoomScales     []int
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
		fontSets:    make(map[string]string),
		XML:         &XMLMap{SRS: "+init=epsg:3857"},
		locator:     locator,
		scaleFactor: 1.0,
		zoomScales:  webmercZoomScales,
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

func (m *Map) SetZoomScales(zoomScales []int) {
	m.zoomScales = zoomScales
}

func (m *Map) AddLayer(l mml.Layer, rules []mss.Rule) {
	if l.ScaleFactor != 0.0 {
		prevScaleFactor := m.scaleFactor
		defer func() { m.scaleFactor = prevScaleFactor }()
		m.scaleFactor = l.ScaleFactor
	}
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

	if l.ClearLabelCache {
		layer.ClearLabelCache = "on"
	}
	if l.CacheFeatures {
		layer.CacheFeatures = "true"
	}

	z := mss.RulesZoom(rules)
	if z != mss.AllZoom {
		if l := z.First(); l > 0 {
			if l > len(m.zoomScales) {
				l = len(m.zoomScales)
			}
			if m.mapnik2 {
				layer.MaxZoom = m.zoomScales[l-1]
			} else {
				layer.MaxScaleDenom = m.zoomScales[l-1]
			}
		}
		if l := z.Last(); l < len(m.zoomScales) {
			if m.mapnik2 {
				layer.MinZoom = m.zoomScales[l]
			} else {
				layer.MinScaleDenom = m.zoomScales[l]
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

// whether a string is a connection (PG:xxx) or filename
var isOgrConnection = regexp.MustCompile(`^[a-zA-Z]{2,}:`)

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
		file := ds.Filename
		if !isOgrConnection.MatchString(ds.Filename) {
			file = m.locator.Data(ds.Filename)
		}
		params = []Parameter{
			{Name: "file", Value: file},
			{Name: "srid", Value: ds.SRID},
			{Name: "extent", Value: ds.Extent},
			{Name: "layer", Value: ds.Layer},
			{Name: "layer_by_sql", Value: ds.Query},
			{Name: "type", Value: "ogr"},
		}
	case mml.GDAL:
		params = []Parameter{
			{Name: "file", Value: m.locator.Data(ds.Filename)},
			{Name: "srid", Value: ds.SRID},
			{Name: "extent", Value: ds.Extent},
			{Name: "band", Value: ds.Band},
			{Name: "type", Value: "gdal"},
		}
	case mml.GeoJson:
		fname := m.locator.Shape(ds.Filename)
		params = []Parameter{
			{Name: "file", Value: fname},
			{Name: "type", Value: "geojson"},
		}
	case nil:
		// datasource might be nil for exports without mml
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
		if l > len(m.zoomScales) {
			l = len(m.zoomScales)
		}
		result.MaxScaleDenom = m.zoomScales[l-1]
	}
	if l := r.Zoom.Last(); l < len(m.zoomScales) {
		result.MinScaleDenom = m.zoomScales[l]
	}

	result.Filter = fmtFilters(r.Filters)
	prefixes := mss.SortedPrefixes(r.Properties, []string{"line-", "polygon-", "polygon-pattern-", "text-", "shield-", "marker-", "point-", "building-", "raster-"})

	for _, p := range prefixes {
		r.Properties.SetDefaultInstance(p.Instance)
		switch p.Name {
		case "line-":
			m.addLineSymbolizer(result, r)
		case "line-pattern-":
			m.addLinePatternSymbolizer(result, r)
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
		case "dot-":
			m.addDotSymbolizer(result, r)
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
	if width, ok := r.Properties.GetFloat("line-width"); ok && width != 0.0 {
		symb := LineSymbolizer{}
		symb.Width = fmtFloat(width*m.scaleFactor, true)
		symb.Clip = fmtBool(r.Properties.GetBool("line-clip"))
		symb.Color = fmtColor(r.Properties.GetColor("line-color"))
		if v, ok := r.Properties.GetFloatList("line-dasharray"); ok {
			symb.Dasharray = fmtPattern(v, m.scaleFactor, true)
		}
		if v, ok := r.Properties.GetFloatList("line-dash-offset"); ok {
			symb.DashOffset = fmtPattern(v, m.scaleFactor, true)
		}
		symb.Gamma = fmtFloat(r.Properties.GetFloat("line-gamma"))
		symb.GammaMethod = fmtString(r.Properties.GetString("line-gamma-method"))
		symb.Linecap = fmtString(r.Properties.GetString("line-cap"))
		symb.Miterlimit = fmtFloatProp(r.Properties, "line-miterlimit", m.scaleFactor)
		symb.Linejoin = fmtString(r.Properties.GetString("line-join"))
		symb.Offset = fmtFloatProp(r.Properties, "line-offset", m.scaleFactor)
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
		symb.Offset = fmtFloatProp(r.Properties, "line-pattern-offset", m.scaleFactor)
		symb.Clip = fmtBool(r.Properties.GetBool("line-pattern-clip"))
		symb.Simplify = fmtFloat(r.Properties.GetFloat("line-pattern-simplify"))
		symb.SimplifyAlgorithm = fmtString(r.Properties.GetString("line-pattern-simplify-algorithm"))
		symb.Smooth = fmtFloat(r.Properties.GetFloat("line-pattern-smooth"))
		symb.GeometryTransform = fmtString(r.Properties.GetString("line-pattern-geometry-transform"))
		symb.CompOp = fmtString(r.Properties.GetString("line-pattern-comp-op"))

		if !m.mapnik2 {
			symb.Opacity = fmtFloat(r.Properties.GetFloat("line-pattern-opacity"))
		}

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
		symb.Size = fmtFloat(size*m.scaleFactor, true)
		symb.Fill = fmtColor(r.Properties.GetColor("text-fill"))
		symb.Name = fmtField(r.Properties.GetFieldList("text-name"))
		symb.AvoidEdges = fmtBool(r.Properties.GetBool("text-avoid-edges"))
		symb.HaloFill = fmtColor(r.Properties.GetColor("text-halo-fill"))
		symb.HaloRadius = fmtFloatProp(r.Properties, "text-halo-radius", m.scaleFactor)
		symb.HaloRasterizer = fmtString(r.Properties.GetString("text-halo-rasterizer"))
		symb.Opacity = fmtFloat(r.Properties.GetFloat("text-opacity"))
		symb.WrapCharacter = fmtString(r.Properties.GetString("text-wrap-character"))
		symb.WrapBefore = fmtString(r.Properties.GetString("text-wrap-before"))
		symb.WrapWidth = fmtFloatProp(r.Properties, "text-wrap-width", m.scaleFactor)
		symb.Ratio = fmtFloat(r.Properties.GetFloat("text-ratio"))
		symb.MaxCharAngleDelta = fmtFloat(r.Properties.GetFloat("text-max-char-angle-delta"))

		symb.Placement = fmtString(r.Properties.GetString("text-placement"))
		symb.PlacementType = fmtString(r.Properties.GetString("text-placement-type"))
		symb.Placements = fmtString(r.Properties.GetString("text-placements"))
		symb.LabelPositionTolerance = fmtFloatProp(r.Properties, "text-label-position-tolerance", m.scaleFactor)
		symb.VerticalAlign = fmtString(r.Properties.GetString("text-vertical-alignment"))
		symb.HorizontalAlign = fmtString(r.Properties.GetString("text-horizontal-alignment"))
		symb.JustifyAlign = fmtString(r.Properties.GetString("text-justify-alignment"))
		symb.CompOp = fmtString(r.Properties.GetString("text-comp-op"))

		symb.Dx = fmtFloatProp(r.Properties, "text-dx", m.scaleFactor)
		symb.Dy = fmtFloatProp(r.Properties, "text-dy", m.scaleFactor)

		if v, ok := r.Properties.GetFloat("text-orientation"); ok {
			symb.Orientation = fmtFloat(v, true)
		} else if v, ok := r.Properties.GetFieldList("text-orientation"); ok {
			symb.Orientation = fmtField(v, true)
		}

		symb.CharacterSpacing = fmtFloatProp(r.Properties, "text-character-spacing", m.scaleFactor)
		symb.LineSpacing = fmtFloatProp(r.Properties, "text-line-spacing", m.scaleFactor)

		symb.AllowOverlap = fmtBool(r.Properties.GetBool("text-allow-overlap"))

		// TODO see for issue/upcoming fixes with 3.0 https://github.com/mapnik/mapnik/issues/2362

		// spacing between repeated labels
		symb.Spacing = fmtFloatProp(r.Properties, "text-spacing", m.scaleFactor)
		// min-distance to other label, does not work with placement-line
		symb.MinimumDistance = fmtFloatProp(r.Properties, "text-min-distance", m.scaleFactor)
		// min-padding to map edge
		symb.MinimumPadding = fmtFloatProp(r.Properties, "text-min-padding", m.scaleFactor)
		symb.MinPathLength = fmtFloatProp(r.Properties, "text-min-path-length", m.scaleFactor)

		symb.Clip = fmtBool(r.Properties.GetBool("text-clip"))
		symb.TextTransform = fmtString(r.Properties.GetString("text-transform"))

		if faceNames, ok := r.Properties.GetStringList("text-face-name"); ok {
			symb.FontsetName = m.fontSetName(faceNames)
		}

		if !m.mapnik2 {
			symb.HaloOpacity = fmtFloat(r.Properties.GetFloat("text-halo-opacity"))
			symb.HaloTransform = fmtString(r.Properties.GetString("text-halo-transform"))
			symb.HaloCompOp = fmtString(r.Properties.GetString("text-halo-comp-op"))
			symb.RepeatWrapCharacter = fmtBool(r.Properties.GetBool("text-repeat-wrap-characater"))
			symb.Margin = fmtFloatProp(r.Properties, "text-margin", m.scaleFactor)
			symb.Simplify = fmtFloat(r.Properties.GetFloat("text-simplify"))
			symb.SimplifyAlgorithm = fmtString(r.Properties.GetString("text-simplify-algorithm"))
			symb.Smooth = fmtFloat(r.Properties.GetFloat("text-smooth"))
			symb.RotateDisplacement = fmtBool(r.Properties.GetBool("text-rotate-displacement"))
			symb.Upright = fmtString(r.Properties.GetString("text-upgright"))
			symb.FontFeatureSettings = fmtString(r.Properties.GetString("font-feature-settings"))
			symb.LargestBboxOnly = fmtBool(r.Properties.GetBool("text-largest-bbox-only"))
			symb.RepeatDistance = fmtFloatProp(r.Properties, "text-repeat-distance", m.scaleFactor)
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

		symb.Size = fmtFloatProp(r.Properties, "shield-size", m.scaleFactor)
		symb.Fill = fmtColor(r.Properties.GetColor("shield-fill"))
		symb.Name = fmtField(r.Properties.GetFieldList("shield-name"))
		symb.TextOpacity = fmtFloat(r.Properties.GetFloat("shield-text-opacity"))
		symb.Opacity = fmtFloat(r.Properties.GetFloat("shield-opacity"))
		symb.Transform = fmtString(r.Properties.GetString("shield-transform"))
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
		symb.HaloRadius = fmtFloatProp(r.Properties, "shield-halo-radius", m.scaleFactor)
		symb.HaloRasterizer = fmtString(r.Properties.GetString("shield-halo-rasterizer"))

		symb.CharacterSpacing = fmtFloatProp(r.Properties, "shield-character-spacing", m.scaleFactor)
		symb.WrapCharacter = fmtString(r.Properties.GetString("shield-wrap-character"))
		symb.WrapBefore = fmtBool(r.Properties.GetBool("shield-wrap-before"))
		symb.WrapWidth = fmtFloatProp(r.Properties, "shield-wrap-width", m.scaleFactor)
		symb.LineSpacing = fmtFloatProp(r.Properties, "shield-line-spacing", m.scaleFactor)
		symb.Dx = fmtFloatProp(r.Properties, "shield-dx", m.scaleFactor)
		symb.Dy = fmtFloatProp(r.Properties, "shield-dx", m.scaleFactor)
		symb.TextDx = fmtFloatProp(r.Properties, "shield-text-dx", m.scaleFactor)
		symb.TextDy = fmtFloatProp(r.Properties, "shield-text-dy", m.scaleFactor)
		symb.TextTransform = fmtString(r.Properties.GetString("shield-text-transform"))

		symb.Spacing = fmtFloatProp(r.Properties, "shield-spacing", m.scaleFactor)
		symb.MinimumDistance = fmtFloatProp(r.Properties, "shield-min-distance", m.scaleFactor)
		symb.MinimumPadding = fmtFloatProp(r.Properties, "shield-min-padding", m.scaleFactor)

		if faceNames, ok := r.Properties.GetStringList("shield-face-name"); ok {
			symb.FontsetName = m.fontSetName(faceNames)
		}

		if !m.mapnik2 {
			symb.HaloTransform = fmtString(r.Properties.GetString("shield-halo-transform"))
			symb.HaloCompOp = fmtString(r.Properties.GetString("shield-halo-comp-op"))
			symb.HaloOpacity = fmtFloat(r.Properties.GetFloat("shield-halo-opacity"))
			symb.LabelPositionTolerance = fmtFloatProp(r.Properties, "shield-label-position-tolerance", m.scaleFactor)
			symb.Margin = fmtFloatProp(r.Properties, "shield-margin", m.scaleFactor)
			symb.RepeatDistance = fmtFloatProp(r.Properties, "shield-repeat-distance", m.scaleFactor)
			symb.Simplify = fmtFloat(r.Properties.GetFloat("shield-simplify"))
			symb.SimplifyAlgorithm = fmtString(r.Properties.GetString("shield-simplify-algorithm"))
			symb.Smooth = fmtFloat(r.Properties.GetFloat("shield-smooth"))
		}

		result.Symbolizers = append(result.Symbolizers, &symb)
	}
}

func (m *Map) addMarkerSymbolizer(result *Rule, r mss.Rule) {
	symb := MarkersSymbolizer{}
	symb.Width = fmtFloatProp(r.Properties, "marker-width", m.scaleFactor)
	symb.Height = fmtFloatProp(r.Properties, "marker-height", m.scaleFactor)
	symb.Fill = fmtColor(r.Properties.GetColor("marker-fill"))
	symb.FillOpacity = fmtFloat(r.Properties.GetFloat("marker-fill-opacity"))
	symb.Opacity = fmtFloat(r.Properties.GetFloat("marker-opacity"))
	symb.Placement = fmtString(r.Properties.GetString("marker-placement"))
	symb.Transform = fmtString(r.Properties.GetString("marker-transform"))
	symb.GeometryTransform = fmtString(r.Properties.GetString("marker-geometry-transform"))
	symb.Spacing = fmtFloatProp(r.Properties, "marker-spacing", m.scaleFactor)
	symb.Stroke = fmtColor(r.Properties.GetColor("marker-line-color"))
	symb.StrokeOpacity = fmtFloat(r.Properties.GetFloat("marker-line-opacity"))
	symb.StrokeWidth = fmtFloatProp(r.Properties, "marker-line-width", m.scaleFactor)
	symb.AllowOverlap = fmtBool(r.Properties.GetBool("marker-allow-overlap"))
	symb.MultiPolicy = fmtString(r.Properties.GetString("marker-multi-policy"))
	symb.IgnorePlacement = fmtBool(r.Properties.GetBool("marker-ignore-placement"))
	symb.MaxError = fmtFloatProp(r.Properties, "marker-max-error", m.scaleFactor)
	symb.Clip = fmtBool(r.Properties.GetBool("marker-clip"))
	symb.Smooth = fmtFloat(r.Properties.GetFloat("marker-smooth"))
	symb.CompOp = fmtString(r.Properties.GetString("marker-comp-op"))

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

	if !m.mapnik2 {
		symb.AvoidEdges = fmtBool(r.Properties.GetBool("marker-avoid-edges"))
		symb.Simplify = fmtFloat(r.Properties.GetFloat("marker-simplify"))
		symb.SimplifyAlgorithm = fmtString(r.Properties.GetString("marker-simplify-algorithm"))
		symb.Offset = fmtFloatProp(r.Properties, "marker-offset", m.scaleFactor)
		symb.Direction = fmtString(r.Properties.GetString("marker-direction"))
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
		symb.Height = fmtFloatProp(r.Properties, "building-height", m.scaleFactor)
		result.Symbolizers = append(result.Symbolizers, &symb)
	}
}

func (m *Map) addDotSymbolizer(result *Rule, r mss.Rule) {
	if !m.mapnik2 {
		if fill, ok := r.Properties.GetColor("dot-fill"); ok {
			symb := DotSymbolizer{}
			symb.Fill = fmtColor(fill, true)
			symb.Opacity = fmtFloat(r.Properties.GetFloat("dot-opacity"))
			symb.Width = fmtFloatProp(r.Properties, "dot-width", m.scaleFactor)
			symb.Height = fmtFloatProp(r.Properties, "dot-height", m.scaleFactor)
			symb.CompOp = fmtString(r.Properties.GetString("dot-comp-op"))
			result.Symbolizers = append(result.Symbolizers, &symb)
		}
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

func fmtPattern(v []float64, scale float64, ok bool) *string {
	if !ok {
		return nil
	}
	parts := make([]string, 0, len(v))
	for i := range v {
		parts = append(parts, strconv.FormatFloat(v[i]*scale, 'f', -1, 64))

	}
	r := strings.Join(parts, ", ")
	return &r
}

func fmtFloatProp(p *mss.Properties, name string, scale float64) *string {
	v, ok := p.GetFloat(name)
	if !ok {
		return nil
	}
	r := strconv.FormatFloat(v*scale, 'f', -1, 64)
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
		case mss.ModuloComparsion:
			value = fmt.Sprintf("%d %s %d", v.Div, v.CompOp, v.Value)
		default:
			log.Printf("unknown type of filter value: %s", v)
			value = ""
		}
		field := f.Field
		if len(field) > 2 && field[0] == '"' && field[len(field)-1] == '"' {
			// strip quotes from field name
			field = field[1 : len(field)-1]
		}
		if f.CompOp == mss.REGEX {
			parts = append(parts, "(["+field+"].match("+value+"))")
		} else {
			parts = append(parts, "(["+field+"] "+f.CompOp.String()+" "+value+")")
		}
	}

	s := strings.Join(parts, " and ")
	if len(filters) > 1 {
		s = "(" + s + ")"
	}
	return s
}

var webmercZoomScales = []int{
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
