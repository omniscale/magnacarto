package mapnik

import (
	"encoding/xml"
)

type XMLMap struct {
	XMLName    xml.Name    `xml:"Map"`
	SRS        string      `xml:"srs,attr"`
	BgColor    *string     `xml:"background-color,attr"`
	Parameters []Parameter `xml:"Parameters>Parameter"`
	FontSets   []FontSet   `xml:"FontSet"`
	Styles     []Style     `xml:"Style"`
	Layers     []Layer     `xml:"Layer"`
}

type Parameter struct {
	Name  string `xml:"name,attr"`
	Value string `xml:",chardata"`
}

type FontSet struct {
	Name  string `xml:"name,attr"`
	Fonts []Font `xml:"Font"`
}

type Font struct {
	FaceName string `xml:"face-name,attr"`
}

type Style struct {
	Name       string   `xml:"name,attr"`
	FilterMode string   `xml:"filter-mode,attr"`
	CompOp     *string  `xml:"comp-op,attr"`
	Opacity    *float64 `xml:"opacity,attr"`
	Rules      []Rule   `xml:"Rule"`
}

type Rule struct {
	Zoom          string `xml:",comment"`
	MaxScaleDenom int64  `xml:"MaxScaleDenominator,omitempty"`
	MinScaleDenom int64  `xml:"MinScaleDenominator,omitempty"`
	Filter        string `xml:"Filter,omitempty"`
	Symbolizers   []interface{}
}

type Symbolizer struct {
	LineSymbolizer           *LineSymbolizer
	PolygonSymbolizer        *PolygonSymbolizer
	PolygonPatternSymbolizer *PolygonPatternSymbolizer
	PointSymbolizer          *PointSymbolizer
	TextSymbolizer           *TextSymbolizer
	MarkersSymbolizer        *MarkersSymbolizer
	ShieldSymbolizer         *ShieldSymbolizer
	RasterSymbolizer         *RasterSymbolizer
}

type Layer struct {
	Name          string       `xml:"name,attr"`
	SRS           *string      `xml:"srs,attr"`
	Status        string       `xml:"status,attr,omitempty"`
	MaxZoom       int64        `xml:"maxzoom,attr,omitempty"`
	MinZoom       int64        `xml:"minzoom,attr,omitempty"`
	MaxScaleDenom int64        `xml:"maximum-scale-denominator,attr,omitempty"`
	MinScaleDenom int64        `xml:"minimum-scale-denominator,attr,omitempty"`
	GroupBy       string       `xml:"group-by,attr,omitempty"`
	StyleNames    []string     `xml:"StyleName"`
	Datasource    *[]Parameter `xml:"Datasource>Parameter"` // as pointer to prevent empty Datasource tag for layers without datasource
}

type PolygonSymbolizer struct {
	XMLName     xml.Name `xml:"PolygonSymbolizer"`
	Clip        *string  `xml:"clip,attr"`
	Color       *string  `xml:"fill,attr"`
	Gamma       *string  `xml:"gamma,attr"`
	GammaMethod *string  `xml:"gamma-method,attr"`
	Opacity     *string  `xml:"fill-opacity,attr"`
	Rasterizer  *string  `xml:"rasterizer,attr"`
	Simplify    *string  `xml:"simplify,attr"`
	Smooth      *string  `xml:"smooth,attr"`
}

type PolygonPatternSymbolizer struct {
	XMLName   xml.Name `xml:"PolygonPatternSymbolizer"`
	File      *string  `xml:"file,attr"`
	Alignment *string  `xml:"alignment,attr"`
}

type BuildingSymbolizer struct {
	XMLName     xml.Name `xml:"BuildingSymbolizer"`
	Fill        *string  `xml:"fill,attr"`
	Height      *string  `xml:"height,attr"`
	FillOpacity *string  `xml:"fill-opacity,attr"`
}

type LineSymbolizer struct {
	XMLName     xml.Name `xml:"LineSymbolizer"`
	Clip        *string  `xml:"clip,attr"`
	Color       *string  `xml:"stroke,attr"`
	Dasharray   *string  `xml:"stroke-dasharray,attr"`
	Gamma       *string  `xml:"stroke-gamma,attr"`
	GammaMethod *string  `xml:"stroke-gamma-method,attr"`
	Linecap     *string  `xml:"stroke-linecap,attr"`
	Linejoin    *string  `xml:"stroke-linejoin,attr"`
	Offset      *string  `xml:"offset,attr"`
	Opacity     *string  `xml:"stroke-opacity,attr"`
	Rasterizer  *string  `xml:"stroke-rasterizer,attr"`
	Simplify    *string  `xml:"stroke-simplify,attr"`
	Smooth      *string  `xml:"stroke-smooth,attr"`
	Width       *string  `xml:"stroke-width,attr"`
}

type PointSymbolizer struct {
	XMLName         xml.Name `xml:"PointSymbolizer"`
	AllowOverlap    *string  `xml:"allow-overlap,attr"`
	File            *string  `xml:"file,attr"`
	Opacity         *string  `xml:"opacity,attr"`
	Transform       *string  `xml:"transform,attr"`
	IgnorePlacement *string  `xml:"ignore-placement,attr"`
}

type TextSymbolizer struct {
	XMLName          xml.Name `xml:"TextSymbolizer"`
	AllowOverlap     *string  `xml:"allow-overlap,attr"`
	AvoidEdges       *string  `xml:"avoid-edges,attr"`
	CharacterSpacing *string  `xml:"character-spacing,attr"`
	Clip             *string  `xml:"clip,attr"`
	Dx               *string  `xml:"dx,attr"`
	Dy               *string  `xml:"dy,attr"`
	FaceName         *string  `xml:"face-name,attr"`
	Fill             *string  `xml:"fill,attr"`
	FontsetName      *string  `xml:"fontset-name,attr"`
	HaloFill         *string  `xml:"halo-fill,attr"`
	HaloRadius       *string  `xml:"halo-radius,attr"`
	LineSpacing      *string  `xml:"line-spacing,attr"`
	MinimumDistance  *string  `xml:"minimum-distance,attr"`
	MinimumPadding   *string  `xml:"minimum-padding,attr"`
	Name             *string  `xml:",chardata"`
	Opacity          *string  `xml:"opacity,attr"`
	Placement        *string  `xml:"placement,attr"`
	Size             *string  `xml:"size,attr"`
	Spacing          *string  `xml:"spacing,attr"`
	TextTransform    *string  `xml:"text-transform,attr"`
	WrapBefore       *string  `xml:"wrap-before,attr"`
	WrapCharacter    *string  `xml:"wrap-character,attr"`
	WrapWidth        *string  `xml:"wrap-width,attr"`
}

type MarkersSymbolizer struct {
	XMLName      xml.Name `xml:"MarkersSymbolizer"`
	AllowOverlap *string  `xml:"allow-overlap,attr"`
	File         *string  `xml:"file,attr"`
	Fill         *string  `xml:"fill,attr"`
	Height       *string  `xml:"height,attr"`
	MarkerType   *string  `xml:"marker-type,attr"`
	Opacity      *string  `xml:"opacity,attr"`
	Placement    *string  `xml:"placement,attr"`
	Spacing      *string  `xml:"spacing,attr"`
	Stroke       *string  `xml:"stroke,attr"`
	StrokeWidth  *string  `xml:"stroke-width,attr"`
	Transform    *string  `xml:"transform,attr"`
	Width        *string  `xml:"width,attr"`
}

type ShieldSymbolizer struct {
	XMLName          xml.Name `xml:"ShieldSymbolizer"`
	AllowOverlap     *string  `xml:"allow-overlap,attr"`
	AvoidEdges       *string  `xml:"avoid-edges,attr"`
	CharacterSpacing *string  `xml:"character-spacing,attr"`
	Clip             *string  `xml:"clip,attr"`
	Dx               *string  `xml:"dx,attr"`
	Dy               *string  `xml:"dy,attr"`
	FaceName         *string  `xml:"face-name,attr"`
	File             *string  `xml:"file,attr"`
	Fill             *string  `xml:"fill,attr"`
	FontsetName      *string  `xml:"fontset-name,attr"`
	HaloFill         *string  `xml:"halo-fill,attr"`
	HaloRadius       *string  `xml:"halo-radius,attr"`
	LineSpacing      *string  `xml:"line-spacing,attr"`
	MinimumDistance  *string  `xml:"minimum-distance,attr"`
	MinimumPadding   *string  `xml:"minimum-padding,attr"`
	Name             *string  `xml:",chardata"`
	Opacity          *string  `xml:"opacity,attr"`
	Placement        *string  `xml:"placement,attr"`
	Size             *string  `xml:"size,attr"`
	Spacing          *string  `xml:"spacing,attr"`
	TextDx           *string  `xml:"text-dx,attr"`
	TextDy           *string  `xml:"text-dy,attr"`
	TextOpacity      *string  `xml:"text-opacity,attr"`
	TextTransform    *string  `xml:"text-transform,attr"`
	WrapBefore       *string  `xml:"wrap-before,attr"`
	WrapCharacter    *string  `xml:"wrap-character,attr"`
	WrapWidth        *string  `xml:"wrap-width,attr"`
}

type RasterSymbolizer struct {
	XMLName      xml.Name `xml:"RasterSymbolizer"`
	CompOp       *string  `xml:"comp-op,attr"`
	DefaultColor *string  `xml:"default-color,attr"`
	DefaultMode  *string  `xml:"default-mode,attr"`
	Epsilon      *string  `xml:"epsilon,attr"`
	FilterFactor *string  `xml:"filter-factor,attr"`
	MeshSize     *string  `xml:"mesh-size,attr"`
	Opacity      *string  `xml:"opacity,attr"`
	Scaling      *string  `xml:"scaling,attr"`
	Stops        []Stop
}

type Stop struct {
	XMLName xml.Name `xml:"stop"`
	Value   string   `xml:"value,attr"`
	Color   string   `xml:"color,attr"`
}
