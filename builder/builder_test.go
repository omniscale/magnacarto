package builder

import (
	"path/filepath"
	"testing"

	"github.com/omniscale/magnacarto/mml"
	"github.com/omniscale/magnacarto/mss"
)

type mockLayer struct {
	layer mml.Layer
	rules []mss.Rule
}

type mockMap struct {
	layers []mockLayer
}

func (m *mockMap) AddLayer(layer mml.Layer, rules []mss.Rule) {
	m.layers = append(m.layers, mockLayer{layer, rules})
}

func (m *mockMap) UnsupportedFeatures() []string {
	return nil
}

func TestBuildEmpty(t *testing.T) {
	m := mockMap{}
	b := New(&m)
	if err := b.Build(); err != nil {
		t.Error(err)
	}
}

func TestBuildMissingMSSFile(t *testing.T) {
	m := mockMap{}
	b := New(&m)
	b.AddMSS("invalid file")
	if err := b.Build(); err == nil {
		t.Error("no error returned for missing file")
	}
}

func TestBuildSimpleMSS(t *testing.T) {
	m := mockMap{}
	b := New(&m)
	b.AddMSS(filepath.Join("tests", "001-single-layer.mss"))
	if err := b.Build(); err != nil {
		t.Fatal(err)
	}
	if len(m.layers) != 1 {
		t.Fatal(m.layers)
	}
	if len(m.layers[0].rules) != 1 || m.layers[0].rules[0].Layer != "foo" {
		t.Error(m.layers[0].rules)
	}
}

func TestBuilEmptyMSS(t *testing.T) {
	m := mockMap{}
	b := New(&m)
	b.SetMML(filepath.Join("tests", "002-empty-mss.mml"))
	if err := b.Build(); err != nil {
		t.Fatal(err)
	}
	if len(m.layers) != 0 {
		t.Fatal(m.layers)
	}
}

func TestBuilActiveInactiveMSS(t *testing.T) {
	m := mockMap{}
	b := New(&m)
	b.SetMML(filepath.Join("tests", "003-two-layers.mml"))
	if err := b.Build(); err != nil {
		t.Fatal(err)
	}
	if len(m.layers) != 2 {
		t.Fatal(m.layers)
	}

	m = mockMap{}
	b = New(&m)
	b.SetIncludeInactive(false)
	b.SetMML(filepath.Join("tests", "003-two-layers.mml"))
	if err := b.Build(); err != nil {
		t.Fatal(err)
	}
	if len(m.layers) != 1 {
		t.Fatal(m.layers)
	}
	if len(m.layers[0].rules) != 1 || m.layers[0].rules[0].Layer != "foo" {
		t.Error(m.layers[0].rules)
	}
}
