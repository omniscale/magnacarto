package tests

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/omniscale/magnacarto/builder"
	"github.com/omniscale/magnacarto/builder/mapnik"
	"github.com/omniscale/magnacarto/builder/mapserver"
	"github.com/omniscale/magnacarto/config"
	"github.com/stretchr/testify/assert"
)

func TestCompareExpected(t *testing.T) {
	mssFiles, err := filepath.Glob("./*.mss")
	if err != nil {
		t.Fatal(err)
	}

	for _, f := range mssFiles {
		name := strings.TrimSuffix(f, filepath.Ext(f))
		t.Run(name, func(t *testing.T) {
			buildAndCompare(t, "mapserver", name)
			buildAndCompare(t, "mapnik", name)
		})
	}
}

func buildAndCompare(t *testing.T, builderType string, name string) {
	suffix := "xml"
	if builderType == "mapserver" {
		suffix = "map"
	}

	actualMap := bytes.Buffer{}
	build(t, builderType, name+".mss", &actualMap)
	expectedFname := name + ".expected." + suffix
	expectedMap, err := ioutil.ReadFile(expectedFname)
	if os.IsNotExist(err) {
		actualFname := name + ".actual." + suffix
		ioutil.WriteFile(actualFname, actualMap.Bytes(), 0644)
		t.Fatalf("missing %s, saved actual file to %s", expectedFname, actualFname)
	} else if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, string(expectedMap), actualMap.String())
}

func build(t *testing.T, builderType string, mssFilename string, out io.Writer) {
	var m builder.MapWriter
	locator := &config.LookupLocator{}
	locator.UseRelPaths(true)

	switch builderType {
	case "mapserver":
		mm := mapserver.New(locator)
		mm.SetNoMapBlock(true)
		m = mm
	case "mapnik":
		m = mapnik.New(locator)
	default:
		t.Fatalf("unknown builderType %s", builderType)
	}

	b := builder.New(m)
	b.AddMSS(mssFilename)

	if err := b.Build(); err != nil {
		t.Fatal("error building style: ", err)
	}

	if err := m.Write(out); err != nil {
		t.Fatal("error writing style: ", err)
	}
}
