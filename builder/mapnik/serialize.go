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
	zoomScales     []int
	proj4          bool
}

type maker struct {
	proj4 bool
}

func (m maker) Type() string       { return "mapnik" }
func (m maker) FileSuffix() string { return ".xml" }
func (m maker) New(locator config.Locator) builder.MapWriter {
	mm := New(locator)
	mm.SetProj4(m.proj4)
	return mm
}

var Maker3 = maker{}
var Maker3Proj4 = maker{proj4: true}

func New(locator config.Locator) *Map {
	return &Map{
		fontSets:    make(map[string]string),
		XML:         &XMLMap{SRS: "epsg:3857"},
		locator:     locator,
		scaleFactor: 1.0,
		zoomScales:  webmercZoomScales,
	}
}

func (m *Map) SetAutoTypeFilter(enable bool) {
	m.autoTypeFilter = enable
}

func (m *Map) SetBackgroundColor(c color.Color) {
	v := c.String()
	m.XML.BgColor = &v
}

func (m *Map) SetZoomScales(zoomScales []int) {
	m.zoomScales = zoomScales
}

// SetProj4 sets the base SRS to Proj4 compatible +init-style if enabled.
func (m *Map) SetProj4(enable bool) {
	if enable {
		m.XML.SRS = "+init=epsg:3857"
		m.proj4 = true
	} else {
		m.XML.SRS = "epsg:3857"
	}
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
	if m.proj4 && strings.HasPrefix(strings.ToLower(*layer.SRS), "epsg:") {
		srs := "+init=" + *layer.SRS
		layer.SRS = &srs
	}
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
			layer.MaxScaleDenom = m.zoomScales[l-1]
		}
		if l := z.Last(); l < len(m.zoomScales) {
			layer.MinScaleDenom = m.zoomScales[l]
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

func (m *Map) UnsupportedFeatures() []string {
	return nil
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
		if ds.SimplifyGeometries == "true" {
			params = append(params, Parameter{Name: "simplify_geometries", Value: "true"})
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
					style.ImageFilters = fmtFunctions(r.Properties.GetFunctions("image-filters"))
					style.DirectImageFilters = fmtFunctions(r.Properties.GetFunctions("direct-image-filters"))
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
	width, fok := r.Properties.GetFloat("line-width")
	_, sok := r.Properties.GetString("line-width")
	if (fok && width != 0.0) || sok {
		symb := LineSymbolizer{}
		symb.Width = fmtFloatScaled(r.Properties, "line-width", m.scaleFactor)
		symb.Clip = fmtBool(r.Properties, "line-clip")
		symb.Color = fmtColor(r.Properties, "line-color")
		if v, ok := r.Properties.GetFloatList("line-dasharray"); ok {
			symb.Dasharray = fmtPattern(v, m.scaleFactor, true)
		}
		if v, ok := r.Properties.GetFloatList("line-dash-offset"); ok {
			symb.DashOffset = fmtPattern(v, m.scaleFactor, true)
		}
		symb.Gamma = fmtFloat(r.Properties, "line-gamma")
		symb.GammaMethod = fmtString(r.Properties, "line-gamma-method")
		symb.Linecap = fmtString(r.Properties, "line-cap")
		symb.Miterlimit = fmtFloatScaled(r.Properties, "line-miterlimit", m.scaleFactor)
		symb.Linejoin = fmtString(r.Properties, "line-join")
		symb.Offset = fmtFloatScaled(r.Properties, "line-offset", m.scaleFactor)
		symb.Opacity = fmtFloat(r.Properties, "line-opacity")
		symb.Rasterizer = fmtString(r.Properties, "line-rasterizer")
		symb.Simplify = fmtFloat(r.Properties, "line-simplify")
		symb.SimplifyAlgorithm = fmtString(r.Properties, "line-simplify-algorithm")
		symb.Smooth = fmtFloat(r.Properties, "line-smooth")
		symb.CompOp = fmtString(r.Properties, "line-comp-op")
		symb.GeometryTransform = fmtString(r.Properties, "line-geometry-transform")
		result.Symbolizers = append(result.Symbolizers, &symb)
	}
}

func (m *Map) addLinePatternSymbolizer(result *Rule, r mss.Rule) {
	if patFile, ok := r.Properties.GetString("line-pattern-file"); ok {
		symb := LinePatternSymbolizer{}
		fname := m.locator.Image(patFile)
		symb.File = &fname
		symb.Offset = fmtFloatScaled(r.Properties, "line-pattern-offset", m.scaleFactor)
		symb.Clip = fmtBool(r.Properties, "line-pattern-clip")
		symb.Simplify = fmtFloat(r.Properties, "line-pattern-simplify")
		symb.SimplifyAlgorithm = fmtString(r.Properties, "line-pattern-simplify-algorithm")
		symb.Smooth = fmtFloat(r.Properties, "line-pattern-smooth")
		symb.GeometryTransform = fmtString(r.Properties, "line-pattern-geometry-transform")
		symb.CompOp = fmtString(r.Properties, "line-pattern-comp-op")
		symb.Opacity = fmtFloat(r.Properties, "line-pattern-opacity")

		result.Symbolizers = append(result.Symbolizers, &symb)
	}
}

func (m *Map) addPolygonSymbolizer(result *Rule, r mss.Rule) {
	_, cok := r.Properties.GetColor("polygon-fill")
	_, sok := r.Properties.GetString("polygon-fill")
	if cok || sok {
		symb := PolygonSymbolizer{}
		symb.Color = fmtColor(r.Properties, "polygon-fill")
		symb.Opacity = fmtFloat(r.Properties, "polygon-opacity")
		symb.Gamma = fmtFloat(r.Properties, "polygon-gamma")
		symb.GammaMethod = fmtString(r.Properties, "polygon-gamma-method")
		symb.Clip = fmtBool(r.Properties, "polygon-clip")
		symb.Simplify = fmtFloat(r.Properties, "polygon-simplify")
		symb.SimplifyAlgorithm = fmtString(r.Properties, "polygon-simplify-algorithm")
		symb.Smooth = fmtFloat(r.Properties, "polygon-smooth")
		symb.GeometryTransform = fmtString(r.Properties, "polygon-geometry-transform")
		symb.CompOp = fmtString(r.Properties, "polygon-comp-op")

		result.Symbolizers = append(result.Symbolizers, &symb)
	}
}

func (m *Map) addTextSymbolizer(result *Rule, r mss.Rule) {
	_, fok := r.Properties.GetFloat("text-size")
	_, sok := r.Properties.GetString("text-size")
	if !fok && !sok {
		return
	}
	symb := TextSymbolizer{TextParameters: m.convertTextParameters(r.Properties)}
	if pl, ok := r.Properties.GetPropertiesList("text-placement-list"); ok {
		placementType := "list"
		symb.PlacementType = &placementType
		for _, p := range pl {
			symb.PlacementList = append(symb.PlacementList, Placement{TextParameters: m.convertTextParameters(p)})
		}
	}
	if symb.RawName != nil && *symb.RawName != "" {
		result.Symbolizers = append(result.Symbolizers, &symb)
	}
}

func (m *Map) convertTextParameters(p *mss.Properties) TextParameters {
	symb := TextParameters{}
	symb.Size = fmtFloatScaled(p, "text-size", m.scaleFactor)
	symb.Fill = fmtColor(p, "text-fill")
	symb.RawName = fmtFormatField(p.GetFieldList("text-name"))
	symb.AvoidEdges = fmtBool(p, "text-avoid-edges")
	symb.HaloFill = fmtColor(p, "text-halo-fill")
	symb.HaloRadius = fmtFloatScaled(p, "text-halo-radius", m.scaleFactor)
	symb.HaloRasterizer = fmtString(p, "text-halo-rasterizer")
	symb.Opacity = fmtFloat(p, "text-opacity")
	symb.WrapCharacter = fmtString(p, "text-wrap-character")
	symb.WrapBefore = fmtString(p, "text-wrap-before")
	symb.WrapWidth = fmtFloatScaled(p, "text-wrap-width", m.scaleFactor)
	symb.Ratio = fmtFloat(p, "text-ratio")
	symb.MaxCharAngleDelta = fmtFloat(p, "text-max-char-angle-delta")

	symb.Placement = fmtString(p, "text-placement")
	symb.PlacementType = fmtString(p, "text-placement-type")
	symb.Placements = fmtString(p, "text-placements")
	symb.LabelPositionTolerance = fmtFloatScaled(p, "text-label-position-tolerance", m.scaleFactor)
	symb.Lang = fmtString(p, "text-lang")
	symb.VerticalAlign = fmtString(p, "text-vertical-alignment")
	symb.HorizontalAlign = fmtString(p, "text-horizontal-alignment")
	symb.JustifyAlign = fmtString(p, "text-justify-alignment")
	symb.CompOp = fmtString(p, "text-comp-op")

	symb.Dx = fmtFloatScaled(p, "text-dx", m.scaleFactor)
	symb.Dy = fmtFloatScaled(p, "text-dy", m.scaleFactor)

	if _, ok := p.GetFloat("text-orientation"); ok {
		symb.Orientation = fmtFloat(p, "text-orientation")
	} else if v, ok := p.GetFieldList("text-orientation"); ok {
		symb.Orientation = fmtField(v, true)
	}

	symb.CharacterSpacing = fmtFloatScaled(p, "text-character-spacing", m.scaleFactor)
	symb.LineSpacing = fmtFloatScaled(p, "text-line-spacing", m.scaleFactor)

	symb.AllowOverlap = fmtBool(p, "text-allow-overlap")

	// TODO see for issue/upcoming fixes with 3.0 https://github.com/mapnik/mapnik/issues/2362

	// spacing between repeated labels
	symb.Spacing = fmtFloatScaled(p, "text-spacing", m.scaleFactor)
	// min-distance to other label, does not work with placement-line
	symb.MinimumDistance = fmtFloatScaled(p, "text-min-distance", m.scaleFactor)
	// min-padding to map edge
	symb.MinimumPadding = fmtFloatScaled(p, "text-min-padding", m.scaleFactor)
	symb.MinPathLength = fmtFloatScaled(p, "text-min-path-length", m.scaleFactor)

	symb.Clip = fmtBool(p, "text-clip")
	symb.TextTransform = fmtString(p, "text-transform")

	if faceNames, ok := p.GetStringList("text-face-name"); ok {
		symb.FontsetName = m.fontSetName(faceNames)
	}

	symb.HaloOpacity = fmtFloat(p, "text-halo-opacity")
	symb.HaloTransform = fmtString(p, "text-halo-transform")
	symb.HaloCompOp = fmtString(p, "text-halo-comp-op")
	symb.RepeatWrapCharacter = fmtBool(p, "text-repeat-wrap-characater")
	symb.Margin = fmtFloatScaled(p, "text-margin", m.scaleFactor)
	symb.Simplify = fmtFloat(p, "text-simplify")
	symb.SimplifyAlgorithm = fmtString(p, "text-simplify-algorithm")
	symb.Smooth = fmtFloat(p, "text-smooth")
	symb.RotateDisplacement = fmtBool(p, "text-rotate-displacement")
	symb.Upright = fmtString(p, "text-upgright")
	symb.FontFeatureSettings = fmtString(p, "font-feature-settings")
	symb.LargestBboxOnly = fmtBool(p, "text-largest-bbox-only")
	symb.RepeatDistance = fmtFloatScaled(p, "text-repeat-distance", m.scaleFactor)
	return symb
}

func (m *Map) addShieldSymbolizer(result *Rule, r mss.Rule) {
	if shieldFile, ok := r.Properties.GetString("shield-file"); ok {
		symb := ShieldSymbolizer{}

		fname := m.locator.Image(shieldFile)
		symb.File = &fname

		symb.Size = fmtFloatScaled(r.Properties, "shield-size", m.scaleFactor)
		symb.Fill = fmtColor(r.Properties, "shield-fill")
		symb.Name = fmtField(r.Properties.GetFieldList("shield-name"))
		symb.TextOpacity = fmtFloat(r.Properties, "shield-text-opacity")
		symb.Opacity = fmtFloat(r.Properties, "shield-opacity")
		symb.Transform = fmtString(r.Properties, "shield-transform")
		symb.CompOp = fmtString(r.Properties, "shield-comp-op")

		symb.Placement = fmtString(r.Properties, "shield-placement")
		symb.PlacementType = fmtString(r.Properties, "shield-placement-type")
		symb.Placements = fmtString(r.Properties, "shield-placements")
		symb.UnlockImage = fmtBool(r.Properties, "shield-unlock-image")
		symb.HorizontalAlign = fmtString(r.Properties, "shield-horizontal-alignment")
		symb.VerticalAlign = fmtString(r.Properties, "shield-vertical-alignment")
		symb.JustifyAlign = fmtString(r.Properties, "shield-justify-alignment")

		symb.Clip = fmtBool(r.Properties, "shield-clip")
		symb.AllowOverlap = fmtBool(r.Properties, "shield-allow-overlap")
		symb.AvoidEdges = fmtBool(r.Properties, "shield-avoid-edges")

		symb.HaloFill = fmtColor(r.Properties, "shield-halo-fill")
		symb.HaloRadius = fmtFloatScaled(r.Properties, "shield-halo-radius", m.scaleFactor)
		symb.HaloRasterizer = fmtString(r.Properties, "shield-halo-rasterizer")

		symb.CharacterSpacing = fmtFloatScaled(r.Properties, "shield-character-spacing", m.scaleFactor)
		symb.WrapCharacter = fmtString(r.Properties, "shield-wrap-character")
		symb.WrapBefore = fmtBool(r.Properties, "shield-wrap-before")
		symb.WrapWidth = fmtFloatScaled(r.Properties, "shield-wrap-width", m.scaleFactor)
		symb.LineSpacing = fmtFloatScaled(r.Properties, "shield-line-spacing", m.scaleFactor)
		symb.Dx = fmtFloatScaled(r.Properties, "shield-dx", m.scaleFactor)
		symb.Dy = fmtFloatScaled(r.Properties, "shield-dx", m.scaleFactor)
		symb.TextDx = fmtFloatScaled(r.Properties, "shield-text-dx", m.scaleFactor)
		symb.TextDy = fmtFloatScaled(r.Properties, "shield-text-dy", m.scaleFactor)
		symb.TextTransform = fmtString(r.Properties, "shield-text-transform")

		symb.Spacing = fmtFloatScaled(r.Properties, "shield-spacing", m.scaleFactor)
		symb.MinimumDistance = fmtFloatScaled(r.Properties, "shield-min-distance", m.scaleFactor)
		symb.MinimumPadding = fmtFloatScaled(r.Properties, "shield-min-padding", m.scaleFactor)

		if faceNames, ok := r.Properties.GetStringList("shield-face-name"); ok {
			symb.FontsetName = m.fontSetName(faceNames)
		}

		symb.HaloTransform = fmtString(r.Properties, "shield-halo-transform")
		symb.HaloCompOp = fmtString(r.Properties, "shield-halo-comp-op")
		symb.HaloOpacity = fmtFloat(r.Properties, "shield-halo-opacity")
		symb.LabelPositionTolerance = fmtFloatScaled(r.Properties, "shield-label-position-tolerance", m.scaleFactor)
		symb.Margin = fmtFloatScaled(r.Properties, "shield-margin", m.scaleFactor)
		symb.RepeatDistance = fmtFloatScaled(r.Properties, "shield-repeat-distance", m.scaleFactor)
		symb.Simplify = fmtFloat(r.Properties, "shield-simplify")
		symb.SimplifyAlgorithm = fmtString(r.Properties, "shield-simplify-algorithm")
		symb.Smooth = fmtFloat(r.Properties, "shield-smooth")

		result.Symbolizers = append(result.Symbolizers, &symb)
	}
}

func (m *Map) addMarkerSymbolizer(result *Rule, r mss.Rule) {
	symb := MarkersSymbolizer{}
	symb.Width = fmtFloatScaled(r.Properties, "marker-width", m.scaleFactor)
	symb.Height = fmtFloatScaled(r.Properties, "marker-height", m.scaleFactor)
	symb.Fill = fmtColor(r.Properties, "marker-fill")
	symb.FillOpacity = fmtFloat(r.Properties, "marker-fill-opacity")
	symb.Opacity = fmtFloat(r.Properties, "marker-opacity")
	symb.Placement = fmtString(r.Properties, "marker-placement")
	symb.Transform = fmtString(r.Properties, "marker-transform")
	symb.GeometryTransform = fmtString(r.Properties, "marker-geometry-transform")
	symb.Spacing = fmtFloatScaled(r.Properties, "marker-spacing", m.scaleFactor)
	symb.Stroke = fmtColor(r.Properties, "marker-line-color")
	symb.StrokeOpacity = fmtFloat(r.Properties, "marker-line-opacity")
	symb.StrokeWidth = fmtFloatScaled(r.Properties, "marker-line-width", m.scaleFactor)
	symb.AllowOverlap = fmtBool(r.Properties, "marker-allow-overlap")
	symb.MultiPolicy = fmtString(r.Properties, "marker-multi-policy")
	symb.IgnorePlacement = fmtBool(r.Properties, "marker-ignore-placement")
	symb.MaxError = fmtFloatScaled(r.Properties, "marker-max-error", m.scaleFactor)
	symb.Clip = fmtBool(r.Properties, "marker-clip")
	symb.Smooth = fmtFloat(r.Properties, "marker-smooth")
	symb.CompOp = fmtString(r.Properties, "marker-comp-op")

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

	symb.AvoidEdges = fmtBool(r.Properties, "marker-avoid-edges")
	symb.Simplify = fmtFloat(r.Properties, "marker-simplify")
	symb.SimplifyAlgorithm = fmtString(r.Properties, "marker-simplify-algorithm")
	symb.Offset = fmtFloatScaled(r.Properties, "marker-offset", m.scaleFactor)
	symb.Direction = fmtString(r.Properties, "marker-direction")

	result.Symbolizers = append(result.Symbolizers, &symb)
}

func (m *Map) addPointSymbolizer(result *Rule, r mss.Rule) {
	if pointFile, ok := r.Properties.GetString("point-file"); ok {
		symb := PointSymbolizer{}
		fname := m.locator.Image(pointFile)
		symb.File = &fname
		symb.AllowOverlap = fmtBool(r.Properties, "point-allow-overlap")
		symb.Opacity = fmtFloat(r.Properties, "point-opacity")
		symb.Transform = fmtString(r.Properties, "point-transform")
		symb.IgnorePlacement = fmtBool(r.Properties, "point-ignore-placement")
		symb.Placement = fmtString(r.Properties, "point-placement")
		symb.CompOp = fmtString(r.Properties, "point-comp-op")
		result.Symbolizers = append(result.Symbolizers, &symb)
	}
}

func (m *Map) addPolygonPatternSymbolizer(result *Rule, r mss.Rule) {
	if patFile, ok := r.Properties.GetString("polygon-pattern-file"); ok {
		symb := PolygonPatternSymbolizer{}
		fname := m.locator.Image(patFile)
		symb.File = &fname
		symb.Alignment = fmtString(r.Properties, "polygon-pattern-alignment")
		symb.Gamma = fmtFloat(r.Properties, "polygon-pattern-gamma")
		symb.Opacity = fmtFloat(r.Properties, "polygon-pattern-opacity")
		symb.Clip = fmtBool(r.Properties, "polygon-pattern-clip")
		symb.Simplify = fmtFloat(r.Properties, "polygon-pattern-simplify")
		symb.SimplifyAlgorithm = fmtString(r.Properties, "polygon-pattern-simplify-algorithm")
		symb.Smooth = fmtFloat(r.Properties, "polygon-pattern-smooth")
		symb.GeometryTransform = fmtString(r.Properties, "polygon-pattern-geometry-transform")
		symb.CompOp = fmtString(r.Properties, "polygon-pattern-comp-op")
		result.Symbolizers = append(result.Symbolizers, &symb)
	}
}

func (m *Map) addBuildingSymbolizer(result *Rule, r mss.Rule) {
	_, cok := r.Properties.GetColor("building-fill")
	_, sok := r.Properties.GetString("building-fill")
	if cok || sok {
		symb := BuildingSymbolizer{}
		symb.Fill = fmtColor(r.Properties, "building-fill")
		symb.FillOpacity = fmtFloat(r.Properties, "building-fill-opacity")
		symb.Height = fmtFloatScaled(r.Properties, "building-height", m.scaleFactor)
		result.Symbolizers = append(result.Symbolizers, &symb)
	}
}

func (m *Map) addDotSymbolizer(result *Rule, r mss.Rule) {
	_, cok := r.Properties.GetColor("dot-fill")
	_, sok := r.Properties.GetString("dot-fill")
	if cok || sok {
		symb := DotSymbolizer{}
		symb.Fill = fmtColor(r.Properties, "dot-fill")
		symb.Opacity = fmtFloat(r.Properties, "dot-opacity")
		symb.Width = fmtFloatScaled(r.Properties, "dot-width", m.scaleFactor)
		symb.Height = fmtFloatScaled(r.Properties, "dot-height", m.scaleFactor)
		symb.CompOp = fmtString(r.Properties, "dot-comp-op")
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
					Color: stop.Color.String(),
				},
			)
		}
	}
	symb.Opacity = fmtFloat(r.Properties, "raster-opacity")
	symb.Epsilon = fmtFloat(r.Properties, "raster-colorizer-epsilon")
	symb.MeshSize = fmtFloat(r.Properties, "raster-mesh-size")
	symb.FilterFactor = fmtFloat(r.Properties, "raster-filter-factor")
	symb.CompOp = fmtString(r.Properties, "raster-comp-op")
	symb.Scaling = fmtString(r.Properties, "raster-scaling")
	symb.DefaultMode = fmtString(r.Properties, "raster-colorizer-default-mode")
	symb.DefaultColor = fmtColor(r.Properties, "raster-colorizer-default-color")
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

// fmtFields concatenates fields and strings with ' + ', but XML with space only.
// Escapes all strings so that the result can be safely used as raw XML.
func fmtFormatField(vals []interface{}, ok bool) *string {
	if !ok {
		return nil
	}

	var b strings.Builder

	type kind int
	const (
		kindNone kind = iota
		kindExpr      // field or string
		kindXML
	)
	prev := kindNone

	addExpr := func(s string) {
		if b.Len() > 0 {
			if prev == kindExpr {
				b.WriteString(" + ")
			} else {
				b.WriteByte(' ')
			}
		}
		xml.EscapeText(&b, []byte(s))
		prev = kindExpr
	}

	addXML := func(s string) {
		if b.Len() > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(s)
		prev = kindXML
	}

	var addVals func(vals []interface{})
	addVals = func(vals []interface{}) {
		for _, v := range vals {
			switch t := v.(type) {
			case mss.Field:
				addExpr(string(t))
			case string:
				addExpr(`"` + t + `"`)
			case []mss.FormatParameter:
				// <Format k1="v1" k2="v2">
				addXML("<Format ")

				for i, p := range t {
					val, ok := p.Value.(string)
					if !ok {
						val = fmt.Sprint(p.Value)
					}
					fmt.Fprintf(&b, `%s=%q`, p.Key, val)
					if i+1 < len(t) {
						b.WriteByte(' ')
					}
				}
				b.WriteByte('>')
			case mss.FormatEnd:
				addXML("</Format>")
			case []mss.Value:
				for _, v := range t {
					switch t := v.(type) {
					case mss.Field:
						addExpr(string(t))
					case string:
						addExpr(`"` + t + `"`)
					}
				}
			default:
				panic(fmt.Sprintf("unexpected field value %#v %T", v, v))
			}
		}
	}

	addVals(vals)

	out := b.String()
	return &out
}

func fmtFunctions(vals []interface{}, ok bool) *string {
	if !ok {
		return nil
	}
	parts := []string{}
	for _, v := range vals {
		switch t := v.(type) {
		case mss.Function:
			function := t.Name + "("
			var params []string
			for _, arg := range t.Params {
				params = append(params, fmt.Sprintf("%v", arg))
			}
			if params != nil {
				function += strings.Join(params, ", ")
			}
			function += ")"
			parts = append(parts, function)
		}
	}
	r := strings.Join(parts, ", ")
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

func fmtFloatScaled(p *mss.Properties, name string, scale float64) *string {
	if v, ok := p.GetFloat(name); ok {
		r := strconv.FormatFloat(v*scale, 'f', -1, 64)
		return &r
	}
	if v, ok := p.GetString(name); ok {
		if scale != 1.0 {
			v = "(" + v + ")*" + strconv.FormatFloat(scale, 'f', -1, 64)
		}
		return &v
	}
	return nil
}

func fmtFloat(p *mss.Properties, name string) *string {
	if v, ok := p.GetFloat(name); ok {
		r := strconv.FormatFloat(v, 'f', -1, 64)
		return &r
	}
	if v, ok := p.GetString(name); ok {
		return &v
	}
	return nil
}

func fmtString(p *mss.Properties, name string) *string {
	if v, ok := p.GetString(name); ok {
		return &v
	}
	return nil
}

func fmtBool(p *mss.Properties, name string) *string {
	if v, ok := p.GetBool(name); ok {
		var r string
		if v {
			r = "true"
		} else {
			r = "false"
		}
		return &r
	}
	if v, ok := p.GetString(name); ok {
		return &v
	}
	return nil
}

func fmtColor(p *mss.Properties, name string) *string {
	if v, ok := p.GetColor(name); ok {
		r := v.String()
		return &r
	}
	if v, ok := p.GetString(name); ok {
		return &v
	}
	return nil
}

func fmtFilters(filters []mss.Filter) string {
	parts := []string{}
	for _, f := range filters {
		var value string
		switch v := f.Value.(type) {
		case nil:
			value = "null"
		case bool:
			if v {
				value = "true"
			} else {
				value = "false"
			}
		case string:
			// TODO quote " in string?!
			value = `'` + v + `'`
		case float64:
			value = strconv.FormatFloat(v, 'f', -1, 64)
		case mss.ModuloComparsion:
			value = fmt.Sprintf("%d %s %d", v.Div, v.CompOp, v.Value)
		default:
			log.Printf("unknown type of filter value: %v", v)
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
