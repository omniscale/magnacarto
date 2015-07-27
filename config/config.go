// Package config implements configuration file parser for magnacarto.
package config

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/omniscale/magnacarto/mml"
)

type Magnacarto struct {
	Mapnik      Mapnik
	MapServer   MapServer
	StylesDir   string `toml:"styles_dir"`
	OutDir      string `toml:"out_dir"`
	Datasources Datasource
	PostGIS     PostGIS
	BaseDir     string
}

type Mapnik struct {
	PluginDirs []string `toml:"plugin_dirs"`
	FontDirs   []string `toml:"font_dirs"`
}

type MapServer struct {
	Bin    string
	DevBin string
}

type Datasource struct {
	NoCheckFiles  bool     `toml:"no_check_files"`
	ShapefileDirs []string `toml:"shapefile_dirs"`
	SQLiteDirs    []string `toml:"sqlite_dirs"`
	ImageDirs     []string `toml:"image_dirs"`
}

type PostGIS struct {
	Host     string
	Port     string
	Database string
	Username string
	Password string
	SRID     string
}

type Locator interface {
	Font(string) string
	SQLite(string) string
	Shape(string) string
	Image(string) string
	PostGIS(mml.PostGIS) mml.PostGIS
}

func Load(fileName string) (*Magnacarto, error) {
	config := Magnacarto{}
	config.BaseDir = filepath.Dir(fileName)
	_, err := toml.DecodeFile(fileName, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (m *Magnacarto) Load(fileName string) error {
	_, err := toml.DecodeFile(fileName, &m)
	if err != nil {
		return err
	}
	return nil
}

func (m *Magnacarto) Locator() Locator {
	if m.Datasources.NoCheckFiles {
		locator := &StaticLocator{baseDir: m.BaseDir}
		if len(m.Mapnik.FontDirs) > 0 {
			locator.fontDir = m.Mapnik.FontDirs[0]
		}
		if len(m.Datasources.SQLiteDirs) > 0 {
			locator.sqliteDir = m.Datasources.SQLiteDirs[0]
		}
		if len(m.Datasources.ShapefileDirs) > 0 {
			locator.shapeDir = m.Datasources.ShapefileDirs[0]
		}
		if len(m.Datasources.ImageDirs) > 0 {
			locator.imageDir = m.Datasources.ImageDirs[0]
		}
		return locator
	}
	locator := &LookupLocator{baseDir: m.BaseDir}
	for _, dir := range m.Datasources.SQLiteDirs {
		locator.AddSQLiteDir(dir)
	}
	for _, dir := range m.Datasources.ImageDirs {
		locator.AddImageDir(dir)
	}
	for _, dir := range m.Datasources.ShapefileDirs {
		locator.AddShapeDir(dir)
	}
	for _, dir := range m.Mapnik.FontDirs {
		locator.AddFontDir(dir)
	}
	locator.AddPGConfig(m.PostGIS)
	return locator
}

type StaticLocator struct {
	fontDir   string
	sqliteDir string
	shapeDir  string
	imageDir  string
	pgConfig  *PostGIS
	baseDir   string
}

func (l *StaticLocator) path(basename, dir string) string {
	if dir != "" {
		return filepath.Join(dir, basename)
	}
	return filepath.Join(l.baseDir, basename)
}

func (l *StaticLocator) Font(basename string) string {
	return filepath.Join(basename, l.fontDir)
}
func (l *StaticLocator) SQLite(basename string) string {
	return l.path(basename, l.sqliteDir)
}
func (l *StaticLocator) Shape(basename string) string {
	return l.path(basename, l.shapeDir)
}
func (l *StaticLocator) Image(basename string) string {
	return l.path(basename, l.imageDir)
}
func (l *StaticLocator) PostGIS(ds mml.PostGIS) mml.PostGIS {
	if l.pgConfig == nil {
		return ds
	}
	c := l.pgConfig
	if c.Host != "" {
		ds.Host = c.Host
	}
	if c.Port != "" {
		ds.Port = c.Port
	}
	if c.Database != "" {
		ds.Database = c.Database
	}
	if c.Username != "" {
		ds.Username = c.Username
	}
	if c.Password != "" {
		ds.Password = c.Password
	}
	if c.SRID != "" {
		ds.SRID = c.SRID
	}

	return ds
}

type LookupLocator struct {
	fontDirs   []string
	sqliteDirs []string
	shapeDirs  []string
	imageDirs  []string
	pgConfig   *PostGIS
	baseDir    string
}

func (l *LookupLocator) find(basename string, dirs []string) string {
	if len(dirs) == 0 {
		if _, err := os.Stat(basename); err == nil {
			return basename
		}
	}
	if filepath.IsAbs(basename) {
		if _, err := os.Stat(basename); err == nil {
			return basename
		}
		basename = filepath.Base(basename)
	}
	if len(dirs) == 0 {
		dirs = []string{l.baseDir}
	}
	for _, d := range dirs {
		fname, err := filepath.Abs(filepath.Join(d, basename))
		if err != nil {
			log.Printf("unable to build abs path for %s and %s", d, basename)
			continue
		}
		if _, err := os.Stat(fname); err == nil {
			return fname
		}
	}
	return ""
}

func (l *LookupLocator) AddFontDir(dir string) {
	l.fontDirs = append(l.fontDirs, dir)
}
func (l *LookupLocator) AddSQLiteDir(dir string) {
	l.sqliteDirs = append(l.sqliteDirs, dir)
}
func (l *LookupLocator) AddShapeDir(dir string) {
	l.shapeDirs = append(l.shapeDirs, dir)
}
func (l *LookupLocator) AddImageDir(dir string) {
	l.imageDirs = append(l.imageDirs, dir)
}
func (l *LookupLocator) AddPGConfig(pgConfig PostGIS) {
	l.pgConfig = &pgConfig
}

func (l *LookupLocator) Font(basename string) string {
	for _, variation := range fontVariations(basename, ".ttf") {
		if file := l.find(variation, l.fontDirs); file != "" {
			return file
		}
	}
	return ""
}

func fontVariations(font, suffix string) []string {
	parts := strings.Split(font, " ")
	var result []string

	result = append(result, strings.Join(parts, "")+suffix)

	for i := 1; i < len(parts); i++ {
		result = append(result,
			strings.Join(parts[:i], "")+"-"+strings.Join(parts[i:], "")+suffix,
		)
	}

	if len(parts) > 1 { // drop last part for "DejaVu Sans Book" -> DejaVuSans.ttf variation
		result = append(result, strings.Join(parts[:len(parts)-1], "")+suffix)
	}

	return result
}

func (l *LookupLocator) SQLite(basename string) string {
	return l.find(basename, l.sqliteDirs)
}
func (l *LookupLocator) Shape(basename string) string {
	return l.find(basename, l.shapeDirs)
}
func (l *LookupLocator) Image(basename string) string {
	return l.find(basename, l.imageDirs)
}
func (l *LookupLocator) PostGIS(ds mml.PostGIS) mml.PostGIS {
	if l.pgConfig == nil {
		return ds
	}
	c := l.pgConfig
	if c.Host != "" {
		ds.Host = c.Host
	}
	if c.Port != "" {
		ds.Port = c.Port
	}
	if c.Database != "" {
		ds.Database = c.Database
	}
	if c.Username != "" {
		ds.Username = c.Username
	}
	if c.Password != "" {
		ds.Password = c.Password
	}
	if c.SRID != "" {
		ds.SRID = c.SRID
	}

	return ds
}

var _ Locator = &LookupLocator{}
