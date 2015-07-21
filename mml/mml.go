// Package mml parses mml-JSON files.
package mml

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type MML struct {
	Layers      []Layer
	Stylesheets []string
}

type auxMML struct {
	Stylesheets []string   `json:"Stylesheet"`
	Layers      []auxLayer `json:"Layer"`
}

type auxLayer struct {
	Datasource map[string]string
	Geometry   string
	Id         string
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
	return &Layer{
		Name:       l.Name,
		Classes:    classes,
		Datasource: ds,
		SRS:        l.SRS,
		Type:       parseGeometryType(l.Geometry),
		Active:     isActive,
		GroupBy:    groupBy,
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
	default:
		return Unknown
	}
}

func newDatasource(d map[string]string) (Datasource, error) {
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
			Extent:   d["extent"],
		}, nil
	} else if d["type"] == "gdal" {
		return GDAL{
			Filename: d["file"],
			SRID:     d["srid"],
			Extent:   d["extent"],
			Band:     d["band"],
		}, nil
	} else if d["type"] == "" {
		return nil, nil
	} else {
		return nil, fmt.Errorf("unsupported datasource type %s in %v", d["type"], d)
	}
}

func Parse(r io.Reader) (*MML, error) {
	aux := auxMML{}
	d := json.NewDecoder(r)
	err := d.Decode(&aux)
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
		Layers:      layers,
		Stylesheets: aux.Stylesheets,
	}

	return &m, nil
}
