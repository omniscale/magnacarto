package mml

type GeometryType string

const (
	Unknown    GeometryType = "Unknown"
	LineString GeometryType = "LineString"
	Polygon    GeometryType = "Polygon"
	Point      GeometryType = "Point"
	Raster     GeometryType = "Raster"
)

type Layer struct {
	ID              string
	Classes         []string
	SRS             string
	Datasource      Datasource
	Type            GeometryType
	Active          bool
	GroupBy         string
	ClearLabelCache bool
	PostLabelCache  bool
	CacheFeatures   bool
	ScaleFactor     float64
}
