package mss

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"github.com/omniscale/magnacarto/color"

	"github.com/stretchr/testify/assert"
)

func TestDecodeFiles(t *testing.T) {
	t.Parallel()
	decodeFiles(t)
}

func BenchmarkTestDecode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		decodeFiles(b)
	}
}

func decodeFiles(t testing.TB) {
	testFiles, err := filepath.Glob("./tests/*.mss")
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range testFiles {
		d := New()
		if err := d.ParseFile(f); err != nil {
			t.Fatal(err)
		}
		if err := d.Evaluate(); err != nil {
			t.Fatal(err)
		}
		for _, layer := range d.MSS().Layers() {
			d.MSS().LayerRules(layer)
		}
		for _, w := range d.warnings {
			t.Error(w.String())
		}
	}
}

func decodeString(content string) (*Decoder, error) {
	d := New()
	err := d.ParseString(content)
	if err := d.Evaluate(); err != nil {
		return nil, err
	}
	return d, err
}

func TestParseColor(t *testing.T) {
	var err error
	var d *Decoder
	_, err = decodeString(`@foo: #zzz;`)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid hex color")

	_, err = decodeString(`@foo: #abcde;`)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid hex color")

	_, err = decodeString(`@foo: #zzzzzz;`)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid hex color")

	d, err = decodeString(`@foo: #abcdef;`)
	assert.NoError(t, err)
	assert.Equal(t, color.MustParse("#abcdef"), d.vars.getKey(key{name: "foo"}))

	d, err = decodeString(`@foo: #fff;`)
	assert.NoError(t, err)
	assert.Equal(t, color.Color{0.0, 0.0, 1.0, 1.0, false}, d.vars.getKey(key{name: "foo"}))

	d, err = decodeString(`@foo: #000;`)
	assert.NoError(t, err)
	assert.Equal(t, color.Color{0.0, 0.0, 0.0, 1.0, false}, d.vars.getKey(key{name: "foo"}))

	d, err = decodeString(`@foo: #00ff00;`)
	assert.NoError(t, err)
	assert.Equal(t, color.Color{120.0, 1.0, 0.5, 1.0, false}, d.vars.getKey(key{name: "foo"}))

	d, err = decodeString(`@foo: rgb(255, 102, 0);`)
	assert.NoError(t, err)
	assert.Equal(t, color.Color{24.0, 1.0, 0.5, 1.0, false}, d.vars.getKey(key{name: "foo"}))

	d, err = decodeString(`@foo: rgb(255, 102, 0, 50);`)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rgb takes exactly three arguments")

	d, err = decodeString(`@foo: rgba(255, 102, 0, 102);`)
	assert.NoError(t, err)
	assert.Equal(t, color.Color{24.0, 1.0, 0.5, 0.4, false}, d.vars.getKey(key{name: "foo"}))

	d, err = decodeString(`@foo: rgb(100%, 40%, 0);`)
	assert.NoError(t, err)
	assert.Equal(t, color.Color{24.0, 1.0, 0.5, 1.0, false}, d.vars.getKey(key{name: "foo"}))

	d, err = decodeString(`@foo: rgba(100%, 102, 0, 20%);`)
	assert.NoError(t, err)
	assert.Equal(t, color.Color{24.0, 1.0, 0.5, 0.2, false}, d.vars.getKey(key{name: "foo"}))

	d, err = decodeString(`@foo: -mc-set-hue(#996644, red); `)
	assert.NoError(t, err)
	assert.Equal(t, "#c64545", d.vars.getKey(key{name: "foo"}).(color.Color).String())

}

func TestParseExpression(t *testing.T) {
	var err error
	var d *Decoder
	_, err = decodeString(`@foo: foo(1);`)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown function foo")

	d, err = decodeString(`@foo: __echo__(1);`)
	assert.NoError(t, err)
	assert.Equal(t, 1, d.vars.getKey(key{name: "foo"}))

	d, err = decodeString(`@foo: fadeout(rgba(255, 255, 255, 255), 50%);`)
	assert.NoError(t, err)
	assert.Equal(t, color.Color{0.0, 0.0, 1.0, 0.5, false}, d.vars.getKey(key{name: "foo"}))

	d, err = decodeString(`@foo: fadein(rgba(255, 255, 255, 0), 50%);`)
	assert.NoError(t, err)
	assert.Equal(t, color.Color{0.0, 0.0, 1.0, 0.5, false}, d.vars.getKey(key{name: "foo"}))

	d, err = decodeString(`@foo: 2 + 2 * 3;`)
	assert.NoError(t, err)
	assert.Equal(t, float64(8), d.vars.getKey(key{name: "foo"}))

	d, err = decodeString(`@foo: (2 + 2) * 3;`)
	assert.NoError(t, err)
	assert.Equal(t, float64(12), d.vars.getKey(key{name: "foo"}))

	d, err = decodeString(`@foo: 2 + 2 * 3 - 2 * 2;`)
	assert.NoError(t, err)
	assert.Equal(t, float64(4), d.vars.getKey(key{name: "foo"}))

	d, err = decodeString(`@foo: ((2 + 2) * 3 - 2) * 2;`)
	assert.NoError(t, err)
	assert.Equal(t, float64(20), d.vars.getKey(key{name: "foo"}))

	d, err = decodeString(`@foo: -(-(2 + 2) * 3 - 2) * 2;`)
	assert.NoError(t, err)
	assert.Equal(t, float64(28), d.vars.getKey(key{name: "foo"}))
}

func decodeLayerProperties(t *testing.T, mss string) *Properties {
	d, err := decodeString(mss)
	assert.NoError(t, err)
	err = d.Evaluate()
	assert.NoError(t, err)

	r := d.MSS().LayerRules("foo")
	fmt.Println(r)
	return r[0].Properties
}

func TestParseInstanceProperties(t *testing.T) {
	var p *Properties

	p = decodeLayerProperties(t, `#foo { a/foo: 2; foo: 1 }`)
	assert.Equal(t, 2, p.getKey(key{name: "foo", instance: "a"}))
	assert.Equal(t, 1, p.getKey(key{name: "foo"}))
	assert.Equal(t, nil, p.getKey(key{name: "foo", instance: "unknown"}))

	// with default instance
	v, _ := p.get("foo")
	assert.Equal(t, 1, v)
	p.SetDefaultInstance("a")
	v, _ = p.get("foo")
	assert.Equal(t, 2, v)
}

func TestDeferEval(t *testing.T) {
	d := New()
	err := d.ParseString(`@foo: red; @bar1: @foo; @foo: blue; @bar2: @foo;`)
	assert.NoError(t, err)
	err = d.Evaluate()
	assert.NoError(t, err)
	// with deferred evaluation bar1 and bar2 reference latest @foo value
	assert.Equal(t, color.MustParse("blue"), d.vars.getKey(key{name: "bar1"}))
	assert.Equal(t, color.MustParse("blue"), d.vars.getKey(key{name: "bar2"}))

	d = New()
	err = d.ParseString(`
		@foo: red;
		#foo {
			line-color: @foo;
			line-width: 1;
		}
		@foo: blue;`)
	assert.NoError(t, err)
	err = d.Evaluate()
	assert.NoError(t, err)
	// with deferred evaluation line-color references latest @foo value
	r := d.MSS().LayerRules("foo")
	c, _ := r[0].Properties.get("line-color")
	assert.Equal(t, color.MustParse("blue"), c)
}

func TestRecursiveDeferEval(t *testing.T) {
	d := New()
	err := d.ParseString(`
		@foo: red;
		@bar: @foo;
		#foo {
			line-color: @baz;
			line-width: 1;
		}
		@baz: @bar;
		@foo: blue;`)
	assert.NoError(t, err)
	err = d.Evaluate()
	assert.NoError(t, err)
	// with deferred evaluation line-color references latest @foo value
	r := d.MSS().LayerRules("foo")
	c, _ := r[0].Properties.get("line-color")
	assert.Equal(t, color.MustParse("blue"), c)
}

func TestParseMissingVar(t *testing.T) {
	var err error
	_, err = decodeString(`@foo: @bar + 1;`)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing var bar")
}

func TestParseBoolVar(t *testing.T) {
	d, err := decodeString(`@foo: false; @bar: @foo;`)
	assert.NoError(t, err)
	assert.Equal(t, false, d.vars.getKey(key{name: "bar"}))
}

func TestParseList(t *testing.T) {
	d, err := decodeString(`@foo: 1, 2, 3;`)
	assert.NoError(t, err)
	assert.Equal(t, []Value{1, 2, 3}, d.vars.getKey(key{name: "foo"}))
}

func TestParseStopList(t *testing.T) {
	d, err := decodeString(`@foo: stop(50, #fff) stop(100, #000);`)
	assert.NoError(t, err)
	assert.Equal(t, []Value{Stop{50, color.Color{0.0, 0.0, 1.0, 1.0, false}}, Stop{100, color.Color{0.0, 0.0, 0.0, 1.0, false}}}, d.vars.getKey(key{name: "foo"}))
}

func TestParseNull(t *testing.T) {
	d, err := decodeString(`@foo: null;
    #foo[type!=null]{line-width: 1}
	`)
	assert.NoError(t, err)
	if val := d.vars.getKey(key{name: "foo"}); val != nil {
		t.Error("foo not nil", val)
	}
	assert.Equal(t, Filter{"type", NEQ, nil}, d.MSS().root.blocks[0].selectors[0].Filters[0])
}

func TestParseMapBlock(t *testing.T) {
	d, err := decodeString(`
	#foo {line-width: 1}
	Map {background-color: red}
	#bar {line-width: 1}
	`)
	assert.NoError(t, err)
	rules := allRules(d.MSS())
	assertRulesEq(t, rules, []Rule{
		Rule{Layer: "foo", Attachment: "", Properties: newProperties("line-width", float64(1)), Zoom: AllZoom, order: 0},
		Rule{Layer: "bar", Attachment: "", Properties: newProperties("line-width", float64(1)), Zoom: AllZoom, order: 1},
	})
	assert.Equal(t, color.MustParse("red"), d.MSS().Map().getKey(key{name: "background-color"}))
}

func allRules(mss *MSS) []Rule {
	rules := []Rule{}
	for _, l := range mss.Layers() {
		rules = append(rules, mss.LayerRules(l)...)
	}
	return rules
}

func loadRules(t *testing.T, f string, layer string, classes ...string) []Rule {
	r, err := os.Open(f)
	if err != nil {
		t.Fatal(err)
	}

	content, err := ioutil.ReadAll(r)
	r.Close()
	if err != nil {
		t.Fatal(err)
	}

	d := New()
	if err = d.ParseString(string(content)); err != nil {
		t.Fatal(err)
	}
	if err = d.Evaluate(); err != nil {
		t.Fatal(err)

	}
	if layer == "ALL" {
		return allRules(d.MSS())
	}

	return d.MSS().LayerRules(layer, classes...)
}

func callLoc(stackDepth int) string {
	_, file, line, _ := runtime.Caller(stackDepth + 1)
	return fmt.Sprintf("in %s:%d", filepath.Base(file), line)
}

func assertRuleEq(t *testing.T, a, b Rule) {
	_assertRulesEq(t, []Rule{a}, []Rule{b})
}

func assertRulesEq(t *testing.T, a, b []Rule) {
	_assertRulesEq(t, a, b)
}

func _assertRulesEq(t *testing.T, a, b []Rule) {
	if len(a) != len(b) {
		t.Fatalf("length do not match %d != %d (%s)", len(a), len(b), callLoc(2))
	}
	for i := range a {
		if !a[i].same(b[i]) {
			t.Fatalf("rule #%d selector do not match (%s)\n\t%v\n !=\t%v", i+1, callLoc(2), a[i], b[i])
		}

		errs := []string{}
		for _, ak := range a[i].Properties.keys() {
			av := a[i].Properties.getKey(ak)
			bv := b[i].Properties.getKey(ak)
			if bv == nil {
				errs = append(errs, fmt.Sprintf("\t\t%s missing in b", ak))
			} else if !reflect.DeepEqual(av, bv) {
				errs = append(errs, fmt.Sprintf("\t\t%s %v != %v", ak, av, bv))
			}
		}
		for _, bk := range b[i].Properties.keys() {
			if a[i].Properties.getKey(bk) == nil {
				errs = append(errs, fmt.Sprintf("\t\t%s missing in a", bk))
			}
		}
		if len(errs) > 0 {
			t.Errorf("rule #%d properties do not match (%s)\n\t%v\n !=\t%v", i+1, callLoc(2), a[i], b[i])
			for _, err := range errs {
				t.Error(err)
			}
		}
	}
}

func TestDecoderRules(t *testing.T) {
	var rules []Rule
	rules = loadRules(t, "tests/001-empty.mss", "ALL")
	assertRulesEq(t, rules, []Rule{})

	rules = loadRules(t, "tests/002-declaration.mss", "ALL")
	assertRulesEq(t, rules, []Rule{
		Rule{Layer: "num", Zoom: AllZoom, Properties: newProperties("line-width", float64(12)), order: 1},
		Rule{Layer: "hash", Zoom: AllZoom, Properties: newProperties("line-width", float64(1), "line-color", color.Color{199.99999999999997, 1.0, 0.7, 1.0, false}), order: 1},
		Rule{Layer: "hash2", Zoom: AllZoom, Properties: newProperties("line-width", float64(1), "line-color", color.Color{240.0, 0.5000000000000001, 0.6000000000000001, 1.0, false}), order: 1},
		Rule{Layer: "rgb", Zoom: AllZoom, Properties: newProperties("line-width", float64(1), "line-color", color.Color{264.0, 1.0, 0.5, 1.0, false}), order: 1},
		Rule{Layer: "rgbpercent", Zoom: AllZoom, Properties: newProperties("line-width", float64(1), "line-color", color.Color{264.0, 1.0, 0.5, 1.0, false}), order: 1},
		Rule{Layer: "rgba", Zoom: AllZoom, Properties: newProperties("line-width", float64(1), "line-color", color.Color{144.0, 1.0, 0.5, 0.4, false}), order: 1},
		Rule{Layer: "rgbacompat", Zoom: AllZoom, Properties: newProperties("line-width", float64(1), "line-color", color.Color{144.0, 1.0, 0.5, 0.4, false}), order: 1},
		Rule{Layer: "rgbapercent", Zoom: AllZoom, Properties: newProperties("line-width", float64(1), "line-color", color.Color{144.0, 1.0, 0.5, 0.4, false}), order: 1},
		Rule{Layer: "list", Zoom: AllZoom, Properties: newProperties("text-name", "foo", "text-size", float64(12), "text-face-name", []Value{"Foo", "Bar", "Baz"}), order: 1},
		Rule{Layer: "listnum", Zoom: AllZoom, Properties: newProperties("line-width", float64(1), "line-dasharray", []Value{float64(2), float64(3), float64(4)}), order: 1},
	})

	rules = loadRules(t, "tests/040-nested.mss", "roads")
	assertRulesEq(t, rules, []Rule{
		Rule{Layer: "roads", Attachment: "", Filters: []Filter{{"service", EQ, "yard"}, {"type", EQ, "rail"}}, Zoom: newZoomRange(EQ, 17), Properties: newProperties("line-width", float64(5), "line-color", color.Color{0.0, 1.0, 0.5, 1.0, false}), order: 0},
		Rule{Layer: "roads", Attachment: "", Filters: []Filter{{"service", EQ, "yard"}, {"type", EQ, "rail"}}, Zoom: AllZoom, Properties: newProperties("line-width", float64(1), "line-color", color.Color{0.0, 1.0, 0.5, 1.0, false}), order: 1},
		Rule{Layer: "roads", Attachment: "", Filters: []Filter{{"type", EQ, "rail"}}, Zoom: newZoomRange(EQ, 17), Properties: newProperties("line-width", float64(5), "line-color", color.Color{60.0, 1.0, 0.5, 1.0, false}), order: 3},
		Rule{Layer: "roads", Attachment: "", Filters: []Filter{}, Zoom: newZoomRange(EQ, 17), Properties: newProperties("line-width", float64(2), "line-color", color.Color{60.0, 1.0, 0.5, 1.0, false}), order: 2},
		Rule{Layer: "roads", Attachment: "", Filters: []Filter{}, Zoom: AllZoom, Properties: newProperties("line-width", float64(1)), order: 2},
	})

	rules = loadRules(t, "tests/021-zoom-specific.mss", "roads")
	assertRulesEq(t, rules, []Rule{
		Rule{Layer: "roads", Attachment: "", Filters: []Filter{{"type", EQ, "primary"}}, Zoom: newZoomRange(EQ, 15), Properties: newProperties("line-width", float64(5), "line-color", color.Color{0.0, 1.0, 0.5, 1.0, false}, "line-cap", "round", "line-join", "bevel")},
		Rule{Layer: "roads", Attachment: "", Filters: []Filter{{"type", EQ, "primary"}}, Zoom: newZoomRange(GTE, 14), Properties: newProperties("line-color", color.Color{0.0, 1.0, 0.5, 1.0, false}, "line-cap", "round", "line-join", "bevel")},
		Rule{Layer: "roads", Attachment: "", Filters: []Filter{}, Zoom: newZoomRange(EQ, 15), Properties: newProperties("line-width", float64(5), "line-color", color.Color{0.0, 0.0, 1.0, 1.0, false}, "line-cap", "round", "line-join", "bevel")},
		Rule{Layer: "roads", Attachment: "", Filters: []Filter{}, Zoom: newZoomRange(GTE, 14), Properties: newProperties("line-color", color.Color{0.0, 0.0, 1.0, 1.0, false}, "line-cap", "round", "line-join", "bevel")},
		Rule{Layer: "roads", Attachment: "", Filters: []Filter{}, Zoom: AllZoom, Properties: newProperties("line-join", "bevel")},
	})

}

func TestDecoderClasses(t *testing.T) {
	var rules []Rule
	rules = loadRules(t, "tests/014-classes.mss", "lakes", "land")
	assertRulesEq(t, rules, []Rule{
		Rule{Layer: "lakes", Class: "land", Attachment: "", Filters: []Filter{}, Zoom: AllZoom, Properties: newProperties("line-width", float64(0.5), "line-color", color.Color{0.0, 1.0, 0.5, 1.0, false}, "polygon-fill", color.Color{240.0, 1.0, 0.5, 1.0, false})},
	})

	// basin class is inside water, no match
	rules = loadRules(t, "tests/014-classes.mss", "", "basin")
	assertRulesEq(t, rules, []Rule{})

	rules = loadRules(t, "tests/014-classes.mss", "", "water")
	assertRulesEq(t, rules, []Rule{
		Rule{Layer: "", Class: "water", Attachment: "", Filters: []Filter{}, Zoom: AllZoom, Properties: newProperties("polygon-fill", color.Color{120.0, 1.0, 0.5, 1.0, false}, "line-width", float64(1))},
	})

	// return .water.basin property regardless of requested class order
	rules = loadRules(t, "tests/014-classes.mss", "", "basin", "water")
	assertRulesEq(t, rules, []Rule{
		Rule{Layer: "", Class: "basin", Attachment: "", Filters: []Filter{}, Zoom: AllZoom, Properties: newProperties("polygon-fill", color.Color{0.0, 0.0, 1.0, 1.0, false}, "line-width", float64(1), "polygon-opacity", float64(0.5))},
	})
	rules = loadRules(t, "tests/014-classes.mss", "", "water", "basin")
	assertRulesEq(t, rules, []Rule{
		Rule{Layer: "", Class: "water", Attachment: "", Filters: []Filter{}, Zoom: AllZoom, Properties: newProperties("polygon-fill", color.Color{0.0, 0.0, 1.0, 1.0, false}, "line-width", float64(1), "polygon-opacity", float64(0.5))},
	})

}
