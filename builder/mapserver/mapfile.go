// Package mapserver builds MapServer .map files.
package mapserver

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/omniscale/magnacarto/builder"
	"github.com/omniscale/magnacarto/builder/sql"
	"github.com/omniscale/magnacarto/color"
	"github.com/omniscale/magnacarto/config"
	"github.com/omniscale/magnacarto/mml"
	"github.com/omniscale/magnacarto/mss"
)

type maker struct{}

func (m maker) Type() string       { return "mapserver" }
func (m maker) FileSuffix() string { return ".map" }
func (m maker) New(locator config.Locator) builder.MapWriter {
	return New(locator)
}

type symbolOptions struct {
	fileName  string
	hasAnchor bool
	anchorX   float64
	anchorY   float64
}

var Maker = maker{}

type Map struct {
	Map            Block
	Layers         Block
	bgColor        *color.Color
	fonts          map[string]string
	svgSymbols     map[string]symbolOptions
	pointSymbols   map[string]Block
	locator        config.Locator
	autoTypeFilter bool
	noMapBlock     bool
	scaleFactor    float64
}

func New(locator config.Locator) *Map {
	mapBlock := NewBlock("MAP")
	mapBlock.Add("Name", quote("map"))
	mapBlock.Add("Imagetype", "png")
	mapBlock.Add("Size", "1600 800")
	mapBlock.Add("Units", "meters")
	mapBlock.Add("Defresolution", "72")
	mapBlock.Add("Extent", "-20037508.34 -20037508.34 20037508.34 20037508.34")
	mapBlock.Add("Config", `"MS_ERRORFILE" "stderr"`)
	mapBlock.Add("", NewBlock("Outputformat",
		Item{"Name", quote("png")},
		Item{"Driver", "AGG/PNG"},
		Item{"Mimetype", quote("image/png")},
		Item{"Imagemode", "RGBA"},
		Item{"Extension", quote("png")},
		Item{"Formatoption", quote("GAMMA=0.75")},
		Item{"Formatoption", quote("COMPRESSION=0")},
	))
	web := NewBlock("Web")
	web.Add("", NewBlock("Metadata",
		Item{"ows_enable_request", quote("*")},
		Item{"wms_srs", quote("EPSG:900913 EPSG:4326 EPSG:3857 EPSG:25833")},
		Item{"wms_extent", quote("-20037508.34 -20037508.34 20037508.34 20037508.34")},
		Item{"wms_onlineresource", quote("http://localhost/")},
		Item{"labelcache_map_edge_buffer", quote("-10")},
		Item{"wms_title", quote("osm")},
	))
	mapBlock.Add("", web)
	mapBlock.Add("", NewBlock("projection", Item{"", "'init=epsg:3857'"}))

	return &Map{
		Map:         mapBlock,
		locator:     locator,
		scaleFactor: 1.0,
	}
}

func (m *Map) SetBackgroundColor(c color.Color) {
	m.bgColor = &c
}

func (m *Map) SetAutoTypeFilter(enable bool) {
	m.autoTypeFilter = enable
}

func (m *Map) SetNoMapBlock(enable bool) {
	m.noMapBlock = enable
}

func (m *Map) String() string {
	if m.noMapBlock {
		// erase default MAP block
		m.Map = NewBlock("")
	}
	if m.bgColor != nil {
		m.Map.AddNonNil("ImageColor", fmtColor(*m.bgColor, true))
	}
	m.Map.Add("", m.Layers)
	m.addSymbols()
	return m.Map.String()
}

func (m *Map) addSymbols() {
	for shortName, options := range m.svgSymbols {
		s := NewBlock("SYMBOL")
		s.Add("name", shortName)
		s.Add("image", options.fileName)
		if strings.HasSuffix(options.fileName, "svg") {
			s.Add("type", "svg")
		} else {
			s.Add("type", "pixmap")
		}
		if options.hasAnchor {
			anchorString := fmt.Sprintf("%v %v", options.anchorX, options.anchorY)
			s.Add("anchorpoint", anchorString)
		}
		m.Map.Add("", s)
	}

	for _, s := range m.pointSymbols {
		m.Map.Add("", s)
	}
}

func (m *Map) Write(w io.Writer) error {
	_, err := w.Write([]byte(m.String()))
	return err
}

func (m *Map) writeFontsList(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	for font, shortName := range m.fonts {
		file := m.locator.Font(font)
		fmt.Fprintln(f, shortName, file)
	}
	m.Map.Add("Fontset", "'"+filepath.Base(filename)+"'")
	return nil
}

func (m *Map) WriteFiles(basename string) error {
	// call writeFontsList first to get FONTSET filename added to Map
	if len(m.fonts) > 0 {
		if err := m.writeFontsList(basename + "-fonts.lst"); err != nil {
			return err
		}
	}
	f, err := os.Create(basename)
	if err != nil {
		return err
	}
	defer f.Close()
	err = m.Write(f)
	if err != nil {
		return err
	}

	return nil
}

type classGroup struct {
	name    string
	classes []Block
	opacity float64
}

func (m *Map) AddLayer(layer mml.Layer, rules []mss.Rule) {
	if len(rules) == 0 {
		return
	}

	if layer.ScaleFactor != 0.0 {
		prevScaleFactor := m.scaleFactor
		defer func() { m.scaleFactor = prevScaleFactor }()
		m.scaleFactor = layer.ScaleFactor
	}

	styles := []classGroup{}
	style := classGroup{}

	var t string
	if layer.Type == mml.LineString {
		t = "LINE"
	} else if layer.Type == mml.Polygon {
		t = "POLYGON"
	} else if layer.Type == mml.Point {
		t = "POINT"
	} else {
		log.Println("unknown geometry type for layer", layer.ID)
		return
	}

	for _, r := range rules {
		styleName := r.Layer
		if r.Attachment != "" {
			styleName += "-" + r.Attachment
		}
		if style.name != styleName {
			if len(style.classes) > 0 {
				styles = append(styles, style)
			}
			style = classGroup{name: styleName}
		}
		if v, ok := r.Properties.GetFloat("opacity"); ok {
			style.opacity = v
		}
		c, ok := m.newClass(r, t)
		if ok {
			style.classes = append(style.classes, *c)
		}
	}

	if len(style.classes) > 0 {
		styles = append(styles, style)
	}

	for _, style := range styles {
		l := NewBlock("LAYER")
		l.Add("name", style.name)

		z := mss.RulesZoom(rules)
		if z := z.First(); z > 0 {
			l.Add("MaxScaleDenom", zoomRanges[z])
		}
		if z := z.Last(); z < 22 {
			l.Add("MinScaleDenom", zoomRanges[z+1])
		}

		if style.opacity != 0 {
			l.AddNonNil("Opacity", fmtFloat(style.opacity*m.scaleFactor, true))
		}

		if layer.Active {
			l.Add("status", "ON")
		} else {
			l.Add("status", "OFF")
		}
		l.Add("type", t)

		if layer.PostLabelCache {
			l.Add("postlabelcache", "true")
		}

		m.addDatasource(&l, layer.Datasource, rules)
		for _, c := range style.classes {
			l.Add("", c)
		}
		m.Layers.Add("", l)
	}
}

/*
xxFactors and RESOLUTION
The same line widths, font sizes and some other properties will result in different
renderings in Mapnik and Mapserver. This is basically the result of different internal DPI
definitions. Mapnik uses 90.7 DPI and Mapserver 72 DPI.
Mapserver can be configured to use 90.7 DPI with DEFRESOLUTION and this solves
+ font sizes
but
- line width are to wide
- SVG symbols are too small
Line widths can be scaled easily, but SVG symbols not.
So we keep using 72 DPI and scale the fonts accordingly.
*/

// LineWidthFactor is used to adjust differences of line width and pattern sizes between Mapnik and Mapserver.
const LineWidthFactor = 1 // 1 for DEFRESOLUTION 72, 1.25 (90.7/72) for DEFRESOLUTION 90.7

// HaoWidthFactor is used to adjust differences of halo outline/radius sizes between Mapnik and Mapserver.
const HaloWidthFactor = 2 * 72 / 90.7 // twice the radius

// FontFactor is used to adjust differences of font sized between Mapnik and Mapserver.
const FontFactor = 72 /*dpi*/ / 90.7 /*dpi*/

func (m *Map) newClass(r mss.Rule, layerType string) (b *Block, styled bool) {
	b = &Block{Name: "CLASS"}

	if r.Zoom != mss.AllZoom {
		b.Add("", "# "+r.Zoom.String())
	}
	if l := r.Zoom.First(); l > 0 {
		b.Add("MaxScaleDenom", zoomRanges[l])
	}
	if l := r.Zoom.Last(); l < 22 {
		b.Add("MinScaleDenom", zoomRanges[l+1])
	}
	filter := fmtFilters(r.Filters)
	if filter != "" {
		b.Add("Expression", filter)
	}

	prefixes := mss.SortedPrefixes(r.Properties, []string{"line-", "polygon-", "polygon-pattern-", "text-", "shield-", "marker-", "point-", "building-"})

	for _, p := range prefixes {
		prefixStyled := false
		hidden := false
		r.Properties.SetDefaultInstance(p.Instance)
		switch p.Name {
		case "line-":
			if layerType == "POLYGON" {
				prefixStyled, hidden = m.addPolygonOutlineSymbolizer(b, r)
			} else if layerType == "LINE" {
				prefixStyled, hidden = m.addLineSymbolizer(b, r)
			}
		case "polygon-":
			if layerType == "POLYGON" {
				prefixStyled = m.addPolygonSymbolizer(b, r)
			}
		case "polygon-pattern-":
			prefixStyled = m.addPolygonPatternSymbolizer(b, r)
		case "text-":
			prefixStyled = m.addTextSymbolizer(b, r, layerType == "LINE")
		case "shield-":
			prefixStyled = m.addShieldSymbolizer(b, r)
		case "marker-":
			prefixStyled = m.addMarkerSymbolizer(b, r, layerType == "LINE")
		case "point-":
			prefixStyled = m.addPointSymbolizer(b, r)
		case "building-":
			prefixStyled = m.addBuildingSymbolizer(b, r)
		default:
			log.Println("invalid prefix", p)
		}
		if prefixStyled {
			styled = true
		}
		if hidden {
			b.Add("", NewBlock("STYLE"))
		}

		r.Properties.SetDefaultInstance("")
	}
	return
}

func (m *Map) addLineSymbolizer(b *Block, r mss.Rule) (styled, hidden bool) {
	if width, ok := r.Properties.GetFloat("line-width"); ok {
		if width == 0.0 {
			return true, true
		}
		style := NewBlock("STYLE")
		style.AddNonNil("Width", fmtFloat(width*LineWidthFactor*m.scaleFactor, true))
		if width*LineWidthFactor > 32 {
			// MaxWidth defaults to 32, override for wider lines
			style.AddNonNil("MaxWidth", fmtFloat(width*LineWidthFactor*m.scaleFactor, true))
		}
		if v, ok := r.Properties.GetFloatList("line-dasharray"); ok {
			style.Add("", fmtPattern(v, m.scaleFactor, true))
		}

		style.AddDefault("Color", fmtColor(r.Properties.GetColor("line-color")), "0 0 0")
		if opacity, ok := r.Properties.GetFloat("line-opacity"); ok {
			style.AddNonNil("Opacity", fmtFloat(opacity*100, true))
		}
		style.AddDefault("Linecap", fmtKeyword(r.Properties.GetString("line-cap")), "BUTT")
		style.AddDefault("Linejoin", fmtKeyword(r.Properties.GetString("line-join")), "MITER")
		b.Add("", style)
		return true, false
	}
	return false, false
}

func (m *Map) addPolygonOutlineSymbolizer(b *Block, r mss.Rule) (styled, hidden bool) {
	if width, ok := r.Properties.GetFloat("line-width"); ok {
		if width == 0.0 {
			return true, true
		}
		style := NewBlock("STYLE")
		style.AddNonNil("Width", fmtFloat(width*LineWidthFactor*m.scaleFactor, true))
		if c, ok := r.Properties.GetColor("line-color"); ok {
			if opacity, ok := r.Properties.GetFloat("line-opacity"); ok {
				c = color.FadeOut(c, 1-opacity)
			}
			style.AddNonNil("OutlineColor", fmtColor(c, true))
		}
		if v, ok := r.Properties.GetFloatList("line-dasharray"); ok {
			style.Add("", fmtPattern(v, m.scaleFactor, true))
		}
		style.AddDefault("Linecap", fmtKeyword(r.Properties.GetString("line-cap")), "BUTT")
		style.AddDefault("Linejoin", fmtKeyword(r.Properties.GetString("line-join")), "MITER")
		b.Add("", style)
		return true, false
	}
	return false, false
}

func (m *Map) addPolygonSymbolizer(b *Block, r mss.Rule) (styled bool) {
	if fill, ok := r.Properties.GetColor("polygon-fill"); ok {
		style := NewBlock("STYLE")
		style.AddNonNil("Color", fmtColor(fill, true))
		if opacity, ok := r.Properties.GetFloat("polygon-opacity"); ok {
			style.AddNonNil("Opacity", fmtFloat(opacity*100, true))
		}
		b.Add("", style)
		return true
	}
	return false
}

func (m *Map) addPolygonPatternSymbolizer(b *Block, r mss.Rule) (styled bool) {
	if file, ok := r.Properties.GetString("polygon-pattern-file"); ok {
		style := NewBlock("STYLE")
		style.Add("SYMBOL", *m.symbolName(file, symbolOptions{}))
		b.Add("", style)
		return true
	}
	return false
}

func (m *Map) addTextSymbolizer(b *Block, r mss.Rule, isLine bool) (styled bool) {
	if textSize, ok := r.Properties.GetFloat("text-size"); ok {
		style := NewBlock("LABEL")
		style.AddNonNil("Size", fmtFloat(textSize*FontFactor*m.scaleFactor-0.5, true))
		style.AddNonNil("Color", fmtColor(r.Properties.GetColor("text-fill")))
		style.AddNonNil("Text", fmtFieldString(r.Properties.GetFieldList("text-name")))

		style.AddNonNil("Force", fmtBool(r.Properties.GetBool("text-allow-overlap")))

		if avoidEdges, ok := r.Properties.GetBool("text-avoid-edges"); ok {
			style.AddNonNil("Partials", fmtBool(!avoidEdges, true))
		}

		if v, ok := r.Properties.GetFloat("text-orientation"); ok {
			style.AddNonNil("Angle", fmtFloat(v, true))
		} else if v, ok := r.Properties.GetFieldList("text-orientation"); ok {
			style.AddNonNil("Angle", fmtField(v, true))
		}

		// TODO http://mapserver.org/development/rfc/ms-rfc-57.html
		style.AddNonNil("MinDistance", fmtFloatProp(r.Properties, "text-spacing", m.scaleFactor))
		style.AddNonNil("RepeatDistance", fmtFloatProp(r.Properties, "text-spacing", m.scaleFactor))
		// text-min-padding -> padding to map edge

		// TODO min-distance to other label, does not work in _mapnik_ with placement-line!
		// if dist, ok := r.Properties.GetFloat("text-min-distance"); ok {
		// 	style.AddNonNil("Buffer", fmtFloat(dist/2, true))
		// }

		if fill, ok := r.Properties.GetColor("text-halo-fill"); ok {
			style.AddNonNil("OutlineColor", fmtColor(fill, true))
			if radius, ok := r.Properties.GetFloat("text-halo-radius"); ok {
				style.AddNonNil("OutlineWidth", fmtFloat(radius*HaloWidthFactor*m.scaleFactor, true))
			}
		}

		addOffsetPosition(&style, r.Properties)

		if faceNames, ok := r.Properties.GetStringList("text-face-name"); ok {
			fontNames := m.fontNames(faceNames)
			style.AddNonNil("Font", fontNames)
		}

		style.Add("Type", "truetype")
		if isLine {
			style.Add("Angle", "FOLLOW")
			style.Add("MinFeatureSize", "AUTO")
		}
		if wrapWidth, ok := r.Properties.GetFloat("text-wrap-width"); ok {
			maxLength := wrapWidth / textSize
			style.AddNonNil("MaxLength", fmtFloat(maxLength*m.scaleFactor, true))
			style.AddDefault("Wrap", fmtString(r.Properties.GetString("text-wrap-character")), quote(" "))
			style.Add("Align", "CENTER")
		}
		b.Add("", style)
		return true
	}
	return false
}

func (m *Map) addShieldSymbolizer(b *Block, r mss.Rule) (styled bool) {
	if shieldFile, ok := r.Properties.GetString("shield-file"); ok {
		style := NewBlock("LABEL")

		if shieldSize, ok := r.Properties.GetFloat("shield-size"); ok {
			style.AddNonNil("Size", fmtFloat(shieldSize*FontFactor-0.5*m.scaleFactor, true))
			style.AddNonNil("Color", fmtColor(r.Properties.GetColor("shield-fill")))
			style.AddNonNil("Text", fmtFieldString(r.Properties.GetFieldList("shield-name")))

			style.AddNonNil("Force", fmtBool(r.Properties.GetBool("shield-allow-overlap")))

			style.AddNonNil("MinDistance", fmtFloatProp(r.Properties, "shield-min-distance", m.scaleFactor))
			style.AddNonNil("RepeatDistance", fmtFloatProp(r.Properties, "shield-spacing", m.scaleFactor))
			style.AddNonNil("Buffer", fmtFloatProp(r.Properties, "shield-min-padding", m.scaleFactor))

			if avoidEdges, ok := r.Properties.GetBool("shield-avoid-edges"); ok {
				style.AddNonNil("Partials", fmtBool(!avoidEdges, true))
			}

			if color, ok := r.Properties.GetColor("shield-halo-fill"); ok {
				style.AddNonNil("OutlineColor", fmtColor(color, true))
				style.AddNonNil("OutlineWidth", fmtFloatProp(r.Properties, "shield-halo-radius", m.scaleFactor))
			}

			if faceNames, ok := r.Properties.GetStringList("shield-face-name"); ok {
				fontNames := m.fontNames(faceNames)
				style.AddNonNil("Font", fontNames)
			}

			style.Add("Type", "truetype")

		}

		shield := NewBlock("STYLE")
		shield.Add("SYMBOL", *m.symbolName(shieldFile, symbolOptions{}))

		style.Add("", shield)

		b.Add("", style)

		return true
	}
	return false
}

func (m *Map) addPointSymbolizer(b *Block, r mss.Rule) (styled bool) {
	if pointFile, ok := r.Properties.GetString("point-file"); ok {
		style := NewBlock("STYLE")

		style.Add("SYMBOL", *m.symbolName(pointFile, symbolOptions{}))
		style.AddNonNil("Opacity", fmtFloat(r.Properties.GetFloat("point-opacity")))
		// style.AddNonNil("Force", fmtBool(r.Properties.GetBool("point-allow-overlap")))

		b.Add("", style)
		return true
	}
	return false
}

func (m *Map) addBuildingSymbolizer(b *Block, r mss.Rule) (styled bool) {
	if fill, ok := r.Properties.GetColor("building-fill"); ok {
		outline := color.Darken(fill, 0.15)
		// fake buildings by rendering walls as two separate lines with offset
		b.Add("", NewBlock("STYLE",
			Item{"Width", "2"},
			Item{"Offset", "-0.5 -1.5"},
			Item{"Outlinecolor", *fmtColor(outline, true)},
			Item{"Linecap", "SQUARE"},
			Item{"Linejoin", "MITER"},
		))
		b.Add("", NewBlock("STYLE",
			Item{"Width", "2"},
			Item{"Outlinecolor", *fmtColor(outline, true)},
			Item{"Linecap", "SQUARE"},
			Item{"Linejoin", "MITER"},
		))
		// render roof
		b.Add("", NewBlock("STYLE",
			Item{"Width", "1"},
			Item{"Offset", "-0.5 -3"},
			Item{"Color", *fmtColor(fill, true)},
			Item{"Outlinecolor", *fmtColor(outline, true)},
			// buffer needed for higher "walls"
			// Item{"Geomtransform", "(buffer([shape], 1))"},
			Item{"Linecap", "SQUARE"},
			Item{"Linejoin", "MITER"},
		))
		return true
	}
	return false
}

func (m *Map) addMarkerSymbolizer(b *Block, r mss.Rule, isLine bool) (styled bool) {
	if markerFile, ok := r.Properties.GetString("marker-file"); ok {
		style := NewBlock("STYLE")

		size, sizeOk := r.Properties.GetFloat("marker-height")
		style.AddNonNil("Opacity", fmtFloat(r.Properties.GetFloat("marker-opacity")))
		if avoidEdges, ok := r.Properties.GetBool("marker-avoid-edges"); ok {
			style.AddNonNil("Partials", fmtBool(!avoidEdges, true))
		}

		symOpts := symbolOptions{}

		if transform, ok := r.Properties.GetString("marker-transform"); ok {
			width, wOk := r.Properties.GetFloat("marker-width")
			height, hOk := r.Properties.GetFloat("marker-height")
			if wOk && hOk {
				width *= m.scaleFactor
				height *= m.scaleFactor
				tr, err := parseTransform(transform)
				if err != nil {
					log.Println(err)
				}
				if tr.rotate != 0.0 {
					style.AddNonNil("Angle", fmtFloat(tr.rotate, true))
				}
				if sizeOk && tr.scale != 0.0 {
					size *= tr.scale
				}
				if tr.hasRotateAnchor {
					symOpts.hasAnchor = true
					symOpts.anchorX = (tr.rotateAnchor[0]*m.scaleFactor + width/2) / width
					symOpts.anchorY = (tr.rotateAnchor[1]*m.scaleFactor + height/2) / height
				}
			}
		}
		style.AddNonNil("Size", fmtFloat(size*m.scaleFactor, sizeOk))
		style.Add("SYMBOL", *m.symbolName(markerFile, symOpts))

		// style.AddNonNil("Force", fmtBool(r.Properties.GetBool("marker-allow-overlap")))

		b.Add("", style)
		return true
	}

	if markerType, ok := r.Properties.GetString("marker-type"); ok {
		style := NewBlock("STYLE")

		var size float64
		if markerType == "arrow" {
			style.Add("SYMBOL", *m.arrowSymbol())
			size = 12.0 // matches arrow of mapnik default size
		} else if markerType == "ellipse" {
			style.Add("SYMBOL", *m.ellipseSymbol())
			size = 10.0 // matches arrow of mapnik default size
		} else {
			log.Printf("marker-type %s not supported", markerType)
			return false
		}
		// emulate marker-opacity by fading marker-fill
		if fill, ok := r.Properties.GetColor("marker-fill"); ok {
			if opacity, ok := r.Properties.GetFloat("marker-opacity"); ok {
				style.AddNonNil("Color", fmtColor(color.FadeOut(fill, opacity), true))
			} else {
				style.AddNonNil("Color", fmtColor(fill, true))
			}
		}
		// emulate marker-opacity by fading marker-line-color
		if linecolor, ok := r.Properties.GetColor("marker-line-color"); ok {
			if opacity, ok := r.Properties.GetFloat("marker-opacity"); ok {
				style.AddNonNil("OutlineColor", fmtColor(color.FadeOut(linecolor, opacity), true))
			} else {
				style.AddNonNil("OutlineColor", fmtColor(linecolor, true))
			}
		}

		if lineWidth, ok := r.Properties.GetFloat("marker-line-width"); ok {
			style.AddNonNil("Width", fmtFloat(lineWidth, true))
		}

		if transform, ok := r.Properties.GetString("marker-transform"); ok {
			tr, err := parseTransform(transform)
			if err != nil {
				log.Println(err)
			}
			if tr.rotate != 0.0 {
				style.AddNonNil("Angle", fmtFloat(tr.rotate, true))
			}
			if tr.scale != 0.0 {
				size *= tr.scale
			}
		}

		if width, ok := r.Properties.GetFloat("marker-width"); ok {
			style.AddNonNil("Size", fmtFloat(width*m.scaleFactor, true))
		} else {
			style.AddNonNil("Size", fmtFloat(size*m.scaleFactor, true))
		}

		if isLine {
			if spacing, ok := r.Properties.GetFloat("marker-spacing"); ok {
				style.AddNonNil("Gap", fmtFloat(-spacing*m.scaleFactor, true))
				style.AddNonNil("InitialGap", fmtFloat(spacing*0.5*m.scaleFactor, true))

			} else {
				style.AddNonNil("Gap", fmtFloat(-100*m.scaleFactor, true)) // mapnik default
			}
		}

		b.Add("", style)
		return true
	}

	return false
}

var sanitizeFontName = regexp.MustCompile("[^-a-zA-Z0-9]")
var sanitizeSymbolName = regexp.MustCompile("[^-a-zA-Z0-9]")

func (m *Map) fontNames(fontFaces []string) *string {
	shortNames := []string{}
	for _, fullName := range fontFaces {
		shortName := sanitizeFontName.ReplaceAllString(fullName, "")

		if m.fonts == nil {
			m.fonts = make(map[string]string)
		}
		m.fonts[fullName] = shortName
		shortNames = append(shortNames, shortName)
	}
	result := `"` + strings.Join(shortNames, ",") + `"`
	return &result
}

func (m *Map) symbolName(symbol mss.Value, options symbolOptions) *string {
	str := symbol.(string)
	if str == "" {
		return nil
	}

	shortName := sanitizeSymbolName.ReplaceAllString(str, "-")
	if strings.HasPrefix(shortName, "-") {
		shortName = shortName[1:len(shortName)]
	}
	if options.hasAnchor {
		shortName = fmt.Sprintf("anchor-%v-%v-%v", options.anchorX, options.anchorY, shortName)
		shortName = sanitizeSymbolName.ReplaceAllString(shortName, "-")
	}

	file := m.locator.Image(str)
	if m.svgSymbols == nil {
		m.svgSymbols = make(map[string]symbolOptions)
	}
	options.fileName = file
	m.svgSymbols[shortName] = options
	result := quote(shortName)
	return &result
}

func (m *Map) arrowSymbol() *string {
	name := "arrow"

	if m.pointSymbols == nil {
		m.pointSymbols = make(map[string]Block)
	}

	if _, ok := m.pointSymbols[name]; !ok {
		s := NewBlock("SYMBOL")
		s.Add("type", "vector")
		s.Add("name", quote(name))
		s.Add("filled", "true")
		s.Add("", NewBlock("points", Item{"", `
			0 5
			20 5
			19 0
			28 6
			19 12
			20 7
			0 7
			`}))
		m.pointSymbols[name] = s
	}
	result := quote(name)
	return &result
}

func (m *Map) ellipseSymbol() *string {
	name := "ellipse"

	if m.pointSymbols == nil {
		m.pointSymbols = make(map[string]Block)
	}

	if _, ok := m.pointSymbols[name]; !ok {
		s := NewBlock("SYMBOL")
		s.Add("type", "ellipse")
		s.Add("name", quote(name))
		s.Add("filled", "true")
		s.Add("", NewBlock("points", Item{"", `
			10 10
			`}))
		m.pointSymbols[name] = s
	}

	result := quote(name)
	return &result
}

// addOffsetPosition add text-dx/dy offsets
// set POSITION so that the offsets are from the outer bounds of the
// label. e.g. dx=10 moves the left bound of the label 10 pixels to the right
func addOffsetPosition(style *Block, properties *mss.Properties) {
	dx, _ := properties.GetFloat("text-dx")
	dy, _ := properties.GetFloat("text-dy")

	dy = -dy

	var position string
	if dy < 0 {
		dy = -dy
		position += "l"
	} else if dy > 0 {
		position += "u"
	} else {
		position += "c"
	}
	if dx < 0 {
		dx = -dx
		position += "l"
	} else if dx > 0 {
		position += "r"
	} else {
		position += "c"
	}

	if dx != 0 || dy != 0 {
		style.Add("OFFSET", fmt.Sprintf("%.0f %.0f", dx, dy))
	}
	style.Add("Position", position)
}

var sqlComments = regexp.MustCompile("-- .*?\\n")

func pqSelectString(query, srid string, rules []mss.Rule, autoTypeFilter bool) string {
	/*
	   (select * from osm_landusages where type in ('forest', 'woods')) as landusages
	   ->
	   geometry from (select *, NULL as nullid from (select * from osm_landusages where type in ('forest', 'woods')) as landusages) as nullidq using unique nullid using srid=900913
	*/
	query = sqlComments.ReplaceAllString(query, " ")
	query = strings.Replace(query, "\n", " ", -1)
	query = strings.Replace(query, `"`, `\"`, -1)
	query = strings.Replace(query, `!bbox!`, `!BOX!`, -1)

	if autoTypeFilter {
		filter := sql.FilterString(rules)
		query = sql.WrapWhere(query, filter)
	}

	splitedQuery := strings.Split(strings.TrimRight(query, " "), " ")
	if len(splitedQuery) > 2 && strings.ToLower(splitedQuery[len(splitedQuery)-2]) == "as" {
		return "geometry from (select *, NULL as nullid from " + query + ") as nullidq using unique nullid using srid=" + srid
	}
	return "geometry from " + query
}

func sqliteSelectString(query, srid string) string {
	return "select type, geometry from " + query
}

func pqConnectionString(pg mml.PostGIS) string {
	/*
	   (select * from osm_landusages where type in ('forest', 'woods')) as landusages
	   ->
	   landusages.geometry from (select * from osm_landusages where type in ('forest', 'woods')) as landusages using unique landusages.osm_id using srid=900913
	*/
	parts := []string{}
	if pg.Host != "" {
		parts = append(parts, "host="+pg.Host)
	}
	if pg.Database != "" {
		parts = append(parts, "dbname="+pg.Database)
	}
	if pg.Username != "" {
		parts = append(parts, "user="+pg.Username)
	}
	if pg.Password != "" {
		parts = append(parts, "password="+pg.Password)
	}
	if pg.Port != "" {
		parts = append(parts, "port="+pg.Port)
	}

	return strings.Join(parts, " ")
}

func (m *Map) addDatasource(block *Block, ds mml.Datasource, rules []mss.Rule) {
	switch ds := ds.(type) {
	case mml.PostGIS:
		ds = m.locator.PostGIS(ds)
		block.Add("data", quote(pqSelectString(ds.Query, ds.SRID, rules, m.autoTypeFilter)))
		block.Add("connection", quote(pqConnectionString(ds)))
		block.Add("connectiontype", "postgis")
		block.Add("processing", quote("CLOSE_CONNECTION=DEFER"))
		block.Add("extent", ds.Extent)
		block.Add("", NewBlock("projection", Item{"", quote("init=epsg:" + ds.SRID)}))
	// 	return []Parameter{
	// 		{Name: "host", Value: ds.Host},
	// 		{Name: "port", Value: ds.Port},
	// 		{Name: "geometry_field", Value: ds.GeometryField},
	// 		// {Name: "dbname", Value: ds.Database},
	// 		// {Name: "user", Value: ds.Username},
	// 		// {Name: "password", Value: ds.Password},
	// 		{Name: "extent", Value: ds.Extent},
	// 		{Name: "table", Value: ds.Query},
	// 		{Name: "srid", Value: ds.SRID},
	// 		{Name: "type", Value: "postgis"},
	// 	}
	case mml.Shapefile:
		fname := m.locator.Shape(ds.Filename)
		idx := strings.LastIndex(fname, ".") // without suffix
		block.Add("data", quote(fname[:idx]))
		block.Add("", NewBlock("projection", Item{"", quote("init=epsg:" + ds.SRID)}))
	case mml.SQLite:
		fname := m.locator.SQLite(ds.Filename)
		block.Add("connection", quote(fname))
		block.Add("data", quote(sqliteSelectString(ds.Query, ds.SRID)))
		block.Add("connectiontype", "ogr")
		block.Add("", NewBlock("projection", Item{"", quote("init=epsg:" + ds.SRID)}))
	case mml.OGR:
		block.Add("connection", quote(ds.Filename))
		if ds.Query != "" {
			block.Add("data", quote(ds.Query))
		} else if ds.Layer != "" {
			block.Add("data", quote(ds.Layer))
		}
		block.Add("connectiontype", "ogr")
		block.Add("", NewBlock("projection", Item{"", quote("init=epsg:" + ds.SRID)}))

	// 	return []Parameter{
	// 		// {Name: "file", Value: "/Users/olt/dev/osm_data/sqlites/" + ds.Filename},
	// 		{Name: "file", Value: "/tmp/" + ds.Filename},
	// 		// {Name: "file", Value: ds.Filename},
	// 		{Name: "srid", Value: ds.SRID},
	// 		{Name: "extent", Value: ds.Extent},
	// 		{Name: "geometry_field", Value: ds.GeometryField},
	// 		{Name: "table", Value: ds.Query},
	// 		{Name: "type", Value: "sqlite"},
	// 	}
	case nil:
		// datasource might be nil for exports withour mml
	default:
		fmt.Fprintf(os.Stderr, "datasource not supported by Mapserver: %v\n", ds)
	}
}

type Layer struct {
}

type Item struct {
	Name  string
	Value interface{}
}

func (i Item) String() string {
	switch v := i.Value.(type) {
	case string:
		if i.Name == "" {
			return v
		}
		return strings.ToUpper(i.Name) + " " + v
	default:
		if i.Name == "" {
			return fmt.Sprintf("%v", i.Value)
		}
		return fmt.Sprintf("%s %v", strings.ToUpper(i.Name), i.Value)
	}
}

type Block struct {
	Name  string
	items []Item
}

func (b *Block) Add(name string, value interface{}) {
	b.items = append(b.items, Item{Name: name, Value: value})
}

func (b *Block) AddNonNil(name string, value *string) {
	if value != nil {
		b.items = append(b.items, Item{Name: name, Value: *value})
	}
}

func (b *Block) AddDefault(name string, value *string, def string) {
	if value != nil {
		b.items = append(b.items, Item{Name: name, Value: *value})
	} else {
		b.items = append(b.items, Item{Name: name, Value: def})
	}
}

func (b *Block) Len() int {
	return len(b.items)
}

func NewBlock(name string, items ...Item) Block {
	return Block{name, items}
}

func (b Block) String() string {
	lines := make([]string, 0, len(b.items)+2)

	if b.Name != "" {
		lines = append(lines, strings.ToUpper(b.Name))
		for _, item := range b.items {
			lines = append(lines, indent(item.String(), "  "))
		}
		lines = append(lines, "END")
	} else {
		for _, item := range b.items {
			lines = append(lines, item.String())
		}
	}
	return strings.Join(lines, "\n")
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
