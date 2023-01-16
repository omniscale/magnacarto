// Package mml parses mml-YAML/JSON files.
package mml

import (
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"gopkg.in/yaml.v2"
)

type MML struct {
	Name        string
	Layers      []Layer
	Stylesheets []string
	Map         Map
}

type auxMML struct {
	Name        string
	Stylesheets []string   `yaml:"Stylesheet"`
	Layers      []auxLayer `yaml:"Layer"`
	Map         Map        `yaml:"Map"`
}

type auxLayer struct {
	Datasource map[string]interface{} `yaml:"Datasource"`
	Geometry   string
	ID         string
	Name       string
	Class      string
	SRS        string
	Status     string
	Properties map[string]interface{}
}

func newLayer(l auxLayer) (*Layer, error) {
	ds, err := newDatasource(l.Datasource)
	if err != nil {
		return nil, err
	}

	isActive := true
	if l.Status == "off" {
		isActive = false
	}

	classes := strings.Split(l.Class, " ")
	groupBy, _ := l.Properties["group-by"].(string)
	clearLabelCache, _ := l.Properties["clear-label-cache"].(string)
	cacheFeatures, _ := l.Properties["cache-features"].(string)

	return &Layer{
		ID:              l.ID,
		Classes:         classes,
		Datasource:      ds,
		SRS:             l.SRS,
		Type:            parseGeometryType(l.Geometry),
		Active:          isActive,
		GroupBy:         groupBy,
		ClearLabelCache: clearLabelCache == "on",
		CacheFeatures:   cacheFeatures == "on",
		Properties:      l.Properties,
	}, nil
}

func parseGeometryType(t string) GeometryType {
	switch t {
	case "polygon":
		return Polygon
	case "linestring":
		return LineString
	case "point":
		return Point
	case "raster":
		return Raster
	default:
		return Unknown
	}
}

func newDatasource(params map[string]interface{}) (Datasource, error) {
	d := make(map[string]string, len(params))
	// convert all datasource params to strings (to support {srid: 1234} and {srid: "1234"}, etc.)
	for k, v := range params {
		if s, ok := v.(string); ok {
			d[k] = s
		} else {
			d[k] = fmt.Sprintf("%v", v)
		}
	}

	if d["type"] == "postgis" {
		return PostGIS{
			Username:      d["user"],
			Password:      d["password"],
			Query:         d["table"],
			Host:          d["host"],
			Port:          d["port"],
			Database:      d["dbname"],
			GeometryField: d["geometry_field"],
			Extent:        d["extent"],
			SRID:          d["srid"],
		}, nil
	} else if d["file"] != "" && (d["type"] == "shape" || d["type"] == "") {
		return Shapefile{
			Filename: d["file"],
			SRID:     d["srid"],
		}, nil
	} else if d["type"] == "sqlite" {
		return SQLite{
			Filename:      d["file"],
			SRID:          d["srid"],
			Query:         d["table"],
			GeometryField: d["geometry_field"],
			Extent:        d["extent"],
		}, nil
	} else if d["type"] == "ogr" {
		return OGR{
			Filename: d["file"],
			SRID:     d["srid"],
			Layer:    d["layer"],
			Query:    d["layer_by_sql"],
			Extent:   d["extent"],
		}, nil
	} else if d["type"] == "gdal" {
		processing := asStrings(params["processing"])
		return GDAL{
			Filename:   d["file"],
			SRID:       d["srid"],
			Extent:     d["extent"],
			Band:       d["band"],
			Processing: processing,
		}, nil
	} else if d["type"] == "geojson" {
		return GeoJson{
			Filename: d["file"],
		}, nil
	} else if d["type"] == "" {
		return nil, nil
	} else {
		return nil, fmt.Errorf("unsupported datasource type %s in %v", d["type"], d)
	}
}

func asStrings(v interface{}) []string {
	slice, ok := v.([]interface{})
	if !ok {
		return nil
	}
	var result []string
	for i := range slice {
		s, ok := slice[i].(string)
		if !ok {
			return nil
		}
		result = append(result, s)
	}
	return result
}

func Parse(r io.Reader) (*MML, error) {
	aux := auxMML{}
	input, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal([]byte(input), &aux)
	if err != nil {
		return nil, err
	}

	layers := []Layer{}
	for _, l := range aux.Layers {
		layer, err := newLayer(l)
		if err != nil {
			return nil, err
		}
		layers = append(layers, *layer)
	}

	m := MML{
		Name:        aux.Name,
		Layers:      layers,
		Stylesheets: aux.Stylesheets,
		Map:         aux.Map,
	}

	return &m, nil
}
