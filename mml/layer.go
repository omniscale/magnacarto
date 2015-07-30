package mml

type GeometryType string

const (
	Unknown    GeometryType = "Unknown"
	LineString GeometryType = "LineString"
	Polygon    GeometryType = "Polygon"
	Point      GeometryType = "Point"
)

type Layer struct {
	ID         string
	Classes    []string
	SRS        string
	Datasource Datasource
	Type       GeometryType
	Active     bool
	GroupBy    string
}
