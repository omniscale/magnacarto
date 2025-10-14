// Package config implements configuration file parser for magnacarto.
package config

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/omniscale/magnacarto/mml"
)

type Magnacarto struct {
	Mapnik      Mapnik
	StylesDir   string `toml:"styles_dir"`
	OutDir      string `toml:"out_dir"`
	Datasources Datasource
	PostGIS     PostGIS
	BaseDir     string
}

type Mapnik struct {
	PluginDirs       []string      `toml:"plugin_dirs"`
	FontDirs         []string      `toml:"font_dirs"`
	CacheWaitTimeout time.Duration `toml:"cache_wait_timeout"`
}

type Datasource struct {
	ShapefileDirs []string `toml:"shapefile_dirs"`
	SQLiteDirs    []string `toml:"sqlite_dirs"`
	ImageDirs     []string `toml:"image_dirs"`
	DataDirs      []string `toml:"data_dirs"`
	FontDirs      []string `toml:"font_dirs"`
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
	Data(string) string
	PostGIS(mml.PostGIS) mml.PostGIS
	SetBaseDir(string)
	SetOutDir(string)
	UseRelPaths(bool)
	MissingFiles() []string
}

func Load(fileName string) (*Magnacarto, error) {
	config := Magnacarto{}
	config.BaseDir = filepath.Dir(fileName)
	_, err := toml.DecodeFile(fileName, &config)
	if err != nil {
		return &config, err
	}

	// make dirs relative to BaseDir
	// datasource dirs are converted in Locator
	if !filepath.IsAbs(config.StylesDir) {
		config.StylesDir = filepath.Join(config.BaseDir, config.StylesDir)
	}
	if !filepath.IsAbs(config.OutDir) {
		config.OutDir = filepath.Join(config.BaseDir, config.OutDir)
	}

	if config.Mapnik.CacheWaitTimeout == 0 {
		config.Mapnik.CacheWaitTimeout = 5 * time.Second
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
	locator := &LookupLocator{baseDir: m.BaseDir}
	for _, dir := range m.Datasources.SQLiteDirs {
		if !filepath.IsAbs(dir) {
			dir = filepath.Join(m.BaseDir, dir)
		}
		locator.AddSQLiteDir(dir)
	}
	for _, dir := range m.Datasources.ImageDirs {
		if !filepath.IsAbs(dir) {
			dir = filepath.Join(m.BaseDir, dir)
		}
		locator.AddImageDir(dir)
	}
	for _, dir := range m.Datasources.ShapefileDirs {
		if !filepath.IsAbs(dir) {
			dir = filepath.Join(m.BaseDir, dir)
		}
		locator.AddShapeDir(dir)
	}
	for _, dir := range m.Datasources.DataDirs {
		if !filepath.IsAbs(dir) {
			dir = filepath.Join(m.BaseDir, dir)
		}
		locator.AddDataDir(dir)
	}

	fontDirs := append([]string{}, m.Mapnik.FontDirs...)
	fontDirs = append(fontDirs, m.Datasources.FontDirs...)
	for _, dir := range fontDirs {
		if !filepath.IsAbs(dir) {
			dir = filepath.Join(m.BaseDir, dir)
		}
		locator.AddFontDir(dir)
	}
	locator.SetPGConfig(m.PostGIS)
	return locator
}

type LookupLocator struct {
	fontDirs   []string
	sqliteDirs []string
	shapeDirs  []string
	imageDirs  []string
	dataDirs   []string
	pgConfig   *PostGIS
	baseDir    string
	outDir     string
	relative   bool
	missing    map[string]struct{}
}

func (l *LookupLocator) SetBaseDir(dir string) {
	l.baseDir = dir
}

func (l *LookupLocator) SetOutDir(dir string) {
	l.outDir = dir
}

func (l *LookupLocator) UseRelPaths(rel bool) {
	l.relative = rel
}

func (l *LookupLocator) find(basename string, dirs []string) (string, bool) {
	// helper func: check if basename exists in dir
	var check func(string) string

	bracketsRe := regexp.MustCompile(`\[[^\[\]]*\]`)
	if bracketsRe.MatchString(basename) {
		// use globbing for data expressions in filenames, eg. `shield-[size].svg`
		pattern := bracketsRe.ReplaceAllString(basename, "*")

		check = func(dir string) string {
			m, err := filepath.Glob(filepath.Join(dir, pattern))
			if err == nil && len(m) > 0 {
				return filepath.Join(dir, basename)
			}
			return ""
		}
	} else {
		check = func(dir string) string {
			fname := filepath.Join(dir, basename)
			if _, err := os.Stat(fname); err == nil {
				return fname
			}
			return ""
		}
	}

	// check for file in different dirs, uses closure so that
	// we can return if we found the file
	fname, ok := func() (string, bool) {
		// check without any dir if it's an absolute path
		if filepath.IsAbs(basename) {
			if fname := check(""); fname != "" {
				return fname, true
			}
		}

		// check passed dirs
		for _, d := range dirs {
			if fname := check(d); fname != "" {
				return fname, true
			}
		}
		// check data dirs
		for _, d := range l.dataDirs {
			if fname := check(d); fname != "" {
				return fname, true
			}
		}

		// at last check with basedir
		if fname := check(l.baseDir); fname != "" {
			return fname, true
		}

		return "", false
	}()

	if !ok {
		// register as missing file
		if l.missing == nil {
			l.missing = make(map[string]struct{})
		}
		l.missing[basename] = struct{}{}
		fname = basename
	} else {
		absfname, err := filepath.Abs(fname)
		if err == nil {
			fname = absfname
		}
	}

	if l.relative {
		relfname, err := filepath.Rel(l.outDir, fname)
		if err == nil {
			fname = relfname
		}
	} else {
		if !filepath.IsAbs(fname) { // for missing files
			fname = filepath.Join(l.outDir, fname)
		}
	}
	return fname, ok

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
func (l *LookupLocator) AddDataDir(dir string) {
	l.dataDirs = append(l.dataDirs, dir)
}
func (l *LookupLocator) SetPGConfig(pgConfig PostGIS) {
	l.pgConfig = &pgConfig
}

func (l *LookupLocator) Font(basename string) string {
	for _, variation := range fontVariations(basename, ".ttf") {
		if file, ok := l.find(variation, l.fontDirs); ok {
			return file
		} else {
			// only record basename, if all variations fail
			delete(l.missing, variation)
		}
	}
	l.missing[basename] = struct{}{}
	return ""
}

func (l *LookupLocator) SQLite(basename string) string {
	fname, _ := l.find(basename, l.sqliteDirs)
	return fname
}
func (l *LookupLocator) Shape(basename string) string {
	fname, _ := l.find(basename, l.shapeDirs)
	return fname
}
func (l *LookupLocator) Image(basename string) string {
	fname, _ := l.find(basename, l.imageDirs)
	return fname
}
func (l *LookupLocator) Data(basename string) string {
	fname, _ := l.find(basename, nil) // dataDir is already searched by l.find
	return fname
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

func (l *LookupLocator) MissingFiles() []string {
	if len(l.missing) == 0 {
		return nil
	}
	files := make([]string, 0, len(l.missing))
	for f := range l.missing {
		files = append(files, f)
	}
	sort.Strings(files)
	return files
}

var _ Locator = &LookupLocator{}

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
