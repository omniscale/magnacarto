package mml

type PostGIS struct {
	Id            string
	Host          string
	Port          string
	Database      string
	Username      string
	Password      string
	Query         string
	SRID          string
	GeometryField string
	Extent        string
}

type Shapefile struct {
	Id       string
	Filename string
}

type SQLite struct {
	Id            string
	Filename      string
	SRID          string
	Query         string
	GeometryField string
	Extent        string
}

type OGR struct {
	Id       string
	Filename string
	SRID     string
	Layer    string
	Extent   string
}

type Datasource interface{}
