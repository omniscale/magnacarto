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
			buildAndCompare(t, "mapnik", name)
			buildAndCompare(t, "mapserver", name)
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
	writeActual := false
	if os.IsNotExist(err) {
		writeActual = true
		t.Logf("missing expected file %s", expectedFname)
	} else if err != nil {
		t.Fatal(err)
	}

	if strings.TrimSpace(string(expectedMap)) != actualMap.String() {
		writeActual = true
	}
	if writeActual {
		actualFname := name + ".actual." + suffix
		ioutil.WriteFile(actualFname, actualMap.Bytes(), 0644)
		t.Logf("saved actual file to %s", actualFname)
	}
	assert.Equal(t, strings.TrimSpace(string(expectedMap)), actualMap.String())
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

	if m.UnsupportedFeatures() != nil {
		return
	}

	if err := m.Write(out); err != nil {
		t.Fatal("error writing style: ", err)
	}
}
