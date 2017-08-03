package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestFontVariations(t *testing.T) {
	variations := fontVariations("Foo Sans Bold Oblique", ".ttf")
	if !reflect.DeepEqual(
		variations,
		[]string{
			"FooSansBoldOblique.ttf",
			"Foo-SansBoldOblique.ttf",
			"FooSans-BoldOblique.ttf",
			"FooSansBold-Oblique.ttf",
			"FooSansBold.ttf",
		}) {
		t.Fatal(variations)
	}
}

func TestLookupLocator_RelPaths(t *testing.T) {
	d := newtmpDirTree(t)
	d.addFile("test.png")
	d.addFile("img/test2.png")
	d.addFile("data/data.geojson")
	d.addFile("shp/file.shp")
	d.addFile("sqlite/file.sqlite")

	check := func(baseDir string) {
		relPrefix := filepath.Join("..", filepath.Base(d.dir))
		outDir, err := filepath.Abs(filepath.Join(baseDir, "..", "out"))
		if err != nil {
			t.Fatal("unable to build outdir", err)
		}
		l := LookupLocator{
			baseDir:  baseDir,
			outDir:   outDir,
			relative: true,
		}

		missing := func(fname, expected, recorded string) {
			if fname != expected {
				t.Errorf("unexpected location for missing file %s != %s", fname, expected)
			}
			if _, ok := l.missing[recorded]; !ok {
				t.Errorf("missing file not recorded as %s in %s", recorded, l.missing)
			}

		}

		existing := func(fname, expected string) {
			if fname != expected {
				t.Errorf("unexpected location %s != %s", fname, expected)
			}
		}

		l.AddDataDir(filepath.Join(d.dir, "data"))
		l.AddShapeDir(filepath.Join(d.dir, "shp"))
		l.AddSQLiteDir(filepath.Join(d.dir, "sqlite"))

		// returns abs path
		existing(l.Image("test.png"), filepath.Join(relPrefix, "test.png"))

		// l.Shape looks in ShapeDir and DataDir
		existing(l.Shape("file.shp"), filepath.Join(relPrefix, "shp", "file.shp"))
		existing(l.Shape("data.geojson"), filepath.Join(relPrefix, "data", "data.geojson"))
		missing(l.Shape("file.sqlite"), "file.sqlite", "file.sqlite")

		// l.SQLITE looks in SqliteDir and DataDir
		existing(l.SQLite("file.sqlite"), filepath.Join(relPrefix, "sqlite", "file.sqlite"))
		existing(l.SQLite("data.geojson"), filepath.Join(relPrefix, "data", "data.geojson"))
		missing(l.SQLite("file.shp"), "file.shp", "file.shp")

		// l.Data looks only in DataDir
		existing(l.Data("data.geojson"), filepath.Join(relPrefix, "data", "data.geojson"))
		missing(l.Data("file.shp"), "file.shp", "file.shp")

		missing(l.Image("test2.png"), "test2.png", "test2.png")

		l.AddImageDir(filepath.Join(d.dir, "img"))
		existing(l.Image("test2.png"), filepath.Join(relPrefix, "img", "test2.png"))

		// missing abs path is converted to relative path
		if fname := l.Image("/abs/foo.png"); !strings.HasSuffix(fname, "../../abs/foo.png") {
			t.Errorf("unexpected location for missing file %s", fname)
		}
	}

	// check with absolute basedir
	check(d.dir)

	// and with relative  basedir
	here, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	relPath, err := filepath.Rel(here, d.dir)
	if err != nil {
		t.Fatal("unable to build relpath", err)
	}
	check(relPath)
}

func TestLookupLocator_AbsPaths(t *testing.T) {
	d := newtmpDirTree(t)
	d.addFile("test.png")
	d.addFile("img/test2.png")
	d.addFile("data/data.geojson")
	d.addFile("shp/file.shp")
	d.addFile("sqlite/file.sqlite")

	check := func(baseDir string) {
		l := LookupLocator{
			baseDir:  baseDir,
			outDir:   "/tmp",
			relative: false,
		}

		missing := func(fname, expected, recorded string) {
			if fname != expected {
				t.Errorf("unexpected location for missing file %s != %s", fname, expected)
			}
			if _, ok := l.missing[recorded]; !ok {
				t.Errorf("missing file not recorded as %s in %s", recorded, l.missing)
			}

		}

		existing := func(fname, expected string) {
			if fname != expected {
				t.Errorf("unexpected location %s != %s", fname, expected)
			}
		}

		l.AddDataDir(filepath.Join(d.dir, "data"))
		l.AddShapeDir(filepath.Join(d.dir, "shp"))
		l.AddSQLiteDir(filepath.Join(d.dir, "sqlite"))

		// returns abs path
		existing(l.Image("test.png"), filepath.Join(d.dir, "test.png"))

		// l.Shape looks in ShapeDir and DataDir
		existing(l.Shape("file.shp"), filepath.Join(d.dir, "shp", "file.shp"))
		existing(l.Shape("data.geojson"), filepath.Join(d.dir, "data", "data.geojson"))
		missing(l.Shape("file.sqlite"), filepath.Join("/tmp", "file.sqlite"), "file.sqlite")

		// l.SQLITE looks in SqliteDir and DataDir
		existing(l.SQLite("file.sqlite"), filepath.Join(d.dir, "sqlite", "file.sqlite"))
		existing(l.SQLite("data.geojson"), filepath.Join(d.dir, "data", "data.geojson"))
		missing(l.SQLite("file.shp"), filepath.Join("/tmp", "file.shp"), "file.shp")

		// l.Data looks only in DataDir
		existing(l.Data("data.geojson"), filepath.Join(d.dir, "data", "data.geojson"))
		missing(l.Data("file.shp"), filepath.Join("/tmp", "file.shp"), "file.shp")

		// missing files return abs path based on outDir
		missing(l.Image("test2.png"), "/tmp/test2.png", "test2.png")

		l.AddImageDir(filepath.Join(d.dir, "img"))
		existing(l.Image("test2.png"), filepath.Join(d.dir, "img", "test2.png"))

		// missing abs path is returned as-is
		missing(l.Image("/abs/foo.png"), "/abs/foo.png", "/abs/foo.png")
	}

	// check with absolute basedir
	check(d.dir)

	// and with relative  basedir
	here, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	relPath, err := filepath.Rel(here, d.dir)
	if err != nil {
		t.Fatal("unable to build relpath", err)
	}
	check(relPath)
}

// tmpDirTree creates a temporary directory with (empty) test files
type tmpDirTree struct {
	dir string
	t   *testing.T
}

func newtmpDirTree(t *testing.T) *tmpDirTree {
	dir, err := ioutil.TempDir("", "magnacarto_test")
	if err != nil {
		t.Fatal("unable to create tmp dir", err)
	}
	return &tmpDirTree{dir: dir, t: t}
}

func (d *tmpDirTree) addFile(fname string) {
	dir := filepath.Join(d.dir, filepath.Dir(fname))
	if err := os.MkdirAll(dir, 0755); err != nil {
		d.t.Fatal("unable to create dir:", err)
	}
	f, err := os.Create(filepath.Join(d.dir, fname))
	if err != nil {
		d.t.Fatal("unable to create test file:", err)
	}
	f.Close()
}

func (d *tmpDirTree) remove() {
	if err := os.RemoveAll(d.dir); err != nil {
		d.t.Fatal("unable to remove tmp dir", err)
	}
}
