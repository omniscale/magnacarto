package config

import (
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

func TestLookupLocator(t *testing.T) {
	here, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	l := LookupLocator{
		baseDir:  here,               // ./magnacarto/config
		outDir:   filepath.Dir(here), // ./magnacarto
		relative: true,
	}

	if fname := l.Image("config.go"); fname != "config/config.go" {
		t.Error("unexpected location: ", fname)
	}

	if fname := l.Image("color.go"); fname != "color.go" {
		t.Error("unexpected location for missing file: ", fname)
	}
	if _, ok := l.missing["color.go"]; !ok {
		t.Error("missing file not recorded", l.missing)
	}

	l.AddImageDir(filepath.Join(here, "..", "color"))
	if fname := l.Image("color.go"); fname != "color/color.go" {
		t.Error("unexpected location: ", fname)
	}

	// missing abs path is made relative
	if fname := l.Image("/abs/foo.png"); !strings.HasSuffix(fname, "../../../abs/foo.png") {
		t.Error("unexpected location: ", fname)
	}
	if _, ok := l.missing["/abs/foo.png"]; !ok {
		t.Error("missing file not recorded", l.missing)
	}
}

func TestAbsoluteLookupLocator(t *testing.T) {
	here, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	l := LookupLocator{
		baseDir:  here, // ./magnacarto/config
		outDir:   "/tmp",
		relative: false,
	}

	// returns abs path
	if fname := l.Image("config.go"); fname != filepath.Join(here, "config.go") {
		t.Error("unexpected location: ", fname)
	}

	// missing files return abs path based on outDir
	if fname := l.Image("color.go"); fname != "/tmp/color.go" {
		t.Error("unexpected location for missing file: ", fname)
	}
	if _, ok := l.missing["color.go"]; !ok {
		t.Error("missing file not recorded", l.missing)
	}

	l.AddImageDir(filepath.Join(here, "..", "color"))
	if fname := l.Image("color.go"); fname != filepath.Join(here, "..", "color", "color.go") {
		t.Error("unexpected location: ", fname)
	}

	// missing abs path is returned as-is
	if fname := l.Image("/abs/foo.png"); fname != "/abs/foo.png" {
		t.Error("unexpected location: ", fname)
	}
	if _, ok := l.missing["/abs/foo.png"]; !ok {
		t.Error("missing file not recorded", l.missing)
	}
}
