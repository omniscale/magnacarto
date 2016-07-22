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

	d, err = decodeString(`@foo: hsl(125, 55%, 25%);`)
	assert.NoError(t, err)
	assert.Equal(t, color.Color{125.0, 0.55, 0.25, 1.0, false}, d.vars.getKey(key{name: "foo"}))

	d, err = decodeString(`@foo: hsla(125, 0.55, 0.25, 20%);`)
	assert.NoError(t, err)
	assert.Equal(t, color.Color{125.0, 0.55, 0.25, 0.2, false}, d.vars.getKey(key{name: "foo"}))

	d, err = decodeString(`@foo: husl(125, 55%, 25%);`)
	assert.NoError(t, err)
	assert.Equal(t, color.Color{125.0, 0.55, 0.25, 1.0, true}, d.vars.getKey(key{name: "foo"}))

	d, err = decodeString(`@foo: husla(125, 0.55, 0.25, 20%);`)
	assert.NoError(t, err)
	assert.Equal(t, color.Color{125.0, 0.55, 0.25, 0.2, true}, d.vars.getKey(key{name: "foo"}))

	d, err = decodeString(`@foo: -mc-set-hue(#996644, red); `)
	assert.NoError(t, err)
	assert.Equal(t, "#c64545", d.vars.getKey(key{name: "foo"}).(color.Color).String())

}

func TestParseExpression(t *testing.T) {
	tests := []struct {
		expr  string
		err   string
		value interface{}
	}{

		{`@foo: __echo__(1);`, "", 1.0},
		{`@foo: fadeout(rgba(255, 255, 255, 255), 50%);`, "", color.Color{0.0, 0.0, 1.0, 0.5, false}},
		{`@foo: fadein(rgba(255, 255, 255, 0), 50%);`, "", color.Color{0.0, 0.0, 1.0, 0.5, false}},
		{`@foo: 2 + 2 * 3;`, "", float64(8)},
		{`@foo: (2 + 2) * 3;`, "", float64(12)},
		{`@foo: 2 + 2 * 3 - 2 * 2;`, "", float64(4)},
		{`@foo: ((2 + 2) * 3 - 2) * 2;`, "", float64(20)},
		{`@foo: -(-(2 + 2) * 3 - 2) * 2;`, "", float64(28)},
		{`@one: "one";
		@two: "two";
		@foo: @one + " + " + @two;`,
			"", "one + two"},
		{`@one: "one";
		@two: [field];
		@foo: @one + " + " + @two;`,
			"", "one + [field]"},

		{`@foo: lighten(red, 10%);`, "", color.Color{0, 1, 0.6, 1, false}},
		{`@foo: lighten(red, 10%, 20%);`, "function lighten takes exactly two arguments", nil},
		{`@foo: lighten(122, 10%);`, "function lighten requires color as first argument", nil},
		{`@foo: lighten(red, red);`, "function lighten requires number/percent as second argument", nil},

		{`@foo: hue(red);`, "", 0.0},
		{`@foo: hue(red, red);`, "function hue takes exactly one argument", nil},
		{`@foo: hue(123);`, "function hue requires color as argument", nil},

		{`@foo: mix(red, blue, 0%);`, "", color.MustParse("blue")},
		{`@foo: mix(red, blue);`, "function mix takes exactly three arguments", nil},
		{`@foo: mix(red, blue);`, "function mix takes exactly three arguments", nil},
		{`@foo: mix(red, blue, red, blue);`, "function mix takes exactly three arguments", nil},

		{`@foo: mix(123, blue, 0%);`, "function mix requires color as first and second argument", nil},
		{`@foo: mix(red, 123, 0%);`, "function mix requires color as first and second argument", nil},
		{`@foo: mix(red, blue, red);`, "function mix requires number/percent as third argument", nil},

		{`@foo: rgb(0, 0, 0, 0);`, "rgb takes exactly three arguments", nil},
		{`@foo: rgba(0, 0, 0);`, "rgba takes exactly four arguments", nil},

		{`@foo: [field1] + " " + [field2] + "!";`, "", []Value{Field("[field1]"), " ", Field("[field2]"), "!"}},
		{`@foo: [field1] + [field2];`, "", []Value{Field("[field1]"), Field("[field2]")}},
		{`@foo: "hello " + [field2];`, "", []Value{"hello ", Field("[field2]")}},

		{`@foo: red * 0.5;`, "", color.Color{0, 1.0, 0.25, 1, false}},
		{`@foo: red * blue;`, "unsupported operation", nil},
	}

	for _, tt := range tests {
		d, err := decodeString(tt.expr)
		if tt.err == "" {
			assert.NoError(t, err, "expr %q returnd error %q", tt.expr, err)
		} else if err == nil {
			t.Errorf("expected error %q for %q", tt.err, tt.expr)
		} else {
			assert.Contains(t, err.Error(), tt.err)
		}
		if tt.value != nil {
			assert.Equal(t, tt.value, d.vars.getKey(key{name: "foo"}))
		}
	}
}

func TestParserErrors(t *testing.T) {
	tests := []struct {
		expr string
		msg  string
	}{
		// selector
		{`#foo, #bar[zoom=3] {}`, ""},
		{`#foo, #bar[zoom=3], {}`, ""}, // dangling coma is ok
		{`#foo, #bar[zoom=3], 123 {}`, "expected layer, attachment, class or filter, got NUMBER"},
		{`#bar[zoom=3]{}`, ""},
		{`#bar[zoom 3]{}`, "expected comparsion, got '3'"},
		{`#bar[zoom=3.1]{}`, "invalid zoom level NUMBER"},
		{`#bar[zoom=~3]{}`, "regular expressions are not allowed for zoom levels"},
		{`#bar[zoom="foo"]{}`, "zoom requires num, got STRING"},
		{`@bar: `, "unexpected value EOF"},
		{`@bar 123`, "expected COLON found NUMBER"},
		{`@bar:;`, "unexpected value SEMICOLON"},
		{`@bar: "Foo`, "unclosed quotation mark"},
		{`#bar {1}`, "unexpected token NUM"},

		// root
		{`Map {}`, ""},
		{`Foo {}`, "only 'Map' identifier expected at top level"},
		{`123`, "unexpected token at top level"},

		// fields
		{`@foo: [field];`, ""},
		{`@foo: [123];`, "expected identifier in field name, got NUMBER"},

		// filter
		{`[foo="bar"]{}`, ""},
		{`[foo=null]{}`, ""},
		{`[foo=bar]{}`, "unexpected value in filter 'bar'"},

		// instances
		{`#foo{ a/line-width: 1}`, ""},
		{`#foo{ a/: 1}`, "expected property name for instance, found COLON"},
		{`#foo{ a: 1}`, ""},

		// functions
		{`@foo: bar(red, 10%);`, "unknown function bar"},
		{`@foo: lighten(red, 10%);`, ""},
		{`@foo: lighten(red, 10%, );`, "unexpected value RPAREN"},
		{`@foo: lighten(red, 10%`, "expected end of function or comma, got EOF"},
		{`@foo: lighten(red, 10% 123)`, "expected end of function or comma, got NUM"},
	}

	for _, tt := range tests {
		_, err := decodeString(tt.expr)
		if tt.msg == "" {
			assert.NoError(t, err, "expr %q returnd error %q", tt.expr, err)
		} else if err == nil {
			t.Errorf("expected error %q for %q", tt.msg, tt.expr)
		} else {
			assert.Contains(t, err.Error(), tt.msg)
		}
	}
}

func TestParserWarnings(t *testing.T) {
	tests := []struct {
		expr string
		msg  string
	}{
		// selector
		{`#foo {line-width: "foo"}`, "invalid property value for line-width"},
		{`#foo {line-wi: "foo"}`, "invalid property line-wi"},
	}

	for _, tt := range tests {
		d, err := decodeString(tt.expr)
		assert.NoError(t, err)
		if len(d.warnings) == 1 {
			assert.Contains(t, d.warnings[0].String(), tt.msg)
		} else {
			t.Errorf("parsing %q did not return expected warnings: %q", tt.expr, d.warnings)
		}
	}
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
	assert.Equal(t, 2.0, p.getKey(key{name: "foo", instance: "a"}))
	assert.Equal(t, 1.0, p.getKey(key{name: "foo"}))
	assert.Equal(t, nil, p.getKey(key{name: "foo", instance: "unknown"}))

	// with default instance
	v, _ := p.get("foo")
	assert.Equal(t, 1.0, v)
	p.SetDefaultInstance("a")
	v, _ = p.get("foo")
	assert.Equal(t, 2.0, v)
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
	assert.Equal(t, []Value{1.0, 2.0, 3.0}, d.vars.getKey(key{name: "foo"}))
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
		Rule{Layer: "foo", Attachment: "", Properties: NewProperties("line-width", float64(1)), Zoom: AllZoom, order: 0},
		Rule{Layer: "bar", Attachment: "", Properties: NewProperties("line-width", float64(1)), Zoom: AllZoom, order: 1},
	})
	assert.Equal(t, color.MustParse("red"), d.MSS().Map().getKey(key{name: "background-color"}))
}

func TestLayerZoomRules(t *testing.T) {
	d := New()
	err := d.ParseString(`
		#foo[zoom=13] {
			line-color: red;
			line-width: 1;
		}`)
	assert.NoError(t, err)
	err = d.Evaluate()
	assert.NoError(t, err)
	r := d.MSS().LayerZoomRules("foo", NewZoomRange(EQ, 13))
	assert.Len(t, r, 1)
	c, _ := r[0].Properties.get("line-color")
	assert.Equal(t, color.MustParse("red"), c)

	assert.Empty(t, d.MSS().LayerZoomRules("foo", NewZoomRange(EQ, 12)))
	assert.Empty(t, d.MSS().LayerZoomRules("foo", NewZoomRange(GT, 13)))
	assert.Empty(t, d.MSS().LayerZoomRules("foo", NewZoomRange(LT, 13)))
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
			t.Errorf("rule #%d selector do not match (%s)\n\t%v\n !=\t%v", i+1, callLoc(2), a[i], b[i])
			return
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
		Rule{Layer: "num", Zoom: AllZoom, Properties: NewProperties("line-width", float64(12)), order: 1},
		Rule{Layer: "hash", Zoom: AllZoom, Properties: NewProperties("line-width", float64(1), "line-color", color.Color{199.99999999999997, 1.0, 0.7, 1.0, false}), order: 1},
		Rule{Layer: "hash2", Zoom: AllZoom, Properties: NewProperties("line-width", float64(1), "line-color", color.Color{240.0, 0.5000000000000001, 0.6000000000000001, 1.0, false}), order: 1},
		Rule{Layer: "rgb", Zoom: AllZoom, Properties: NewProperties("line-width", float64(1), "line-color", color.Color{264.0, 1.0, 0.5, 1.0, false}), order: 1},
		Rule{Layer: "rgbpercent", Zoom: AllZoom, Properties: NewProperties("line-width", float64(1), "line-color", color.Color{264.0, 1.0, 0.5, 1.0, false}), order: 1},
		Rule{Layer: "rgba", Zoom: AllZoom, Properties: NewProperties("line-width", float64(1), "line-color", color.Color{144.0, 1.0, 0.5, 0.4, false}), order: 1},
		Rule{Layer: "rgbacompat", Zoom: AllZoom, Properties: NewProperties("line-width", float64(1), "line-color", color.Color{144.0, 1.0, 0.5, 0.4, false}), order: 1},
		Rule{Layer: "rgbapercent", Zoom: AllZoom, Properties: NewProperties("line-width", float64(1), "line-color", color.Color{144.0, 1.0, 0.5, 0.4, false}), order: 1},
		Rule{Layer: "list", Zoom: AllZoom, Properties: NewProperties("text-name", "foo", "text-size", float64(12), "text-face-name", []Value{"Foo", "Bar", "Baz"}), order: 1},
		Rule{Layer: "listnum", Zoom: AllZoom, Properties: NewProperties("line-width", float64(1), "line-dasharray", []Value{float64(2), float64(3), float64(4)}), order: 1},
	})

	rules = loadRules(t, "tests/040-nested.mss", "roads")
	assertRulesEq(t, rules, []Rule{
		Rule{Layer: "roads", Attachment: "", Filters: []Filter{{"service", EQ, "yard"}, {"type", EQ, "rail"}}, Zoom: NewZoomRange(EQ, 17), Properties: NewProperties("line-width", float64(5), "line-color", color.Color{0.0, 1.0, 0.5, 1.0, false}), order: 0},
		Rule{Layer: "roads", Attachment: "", Filters: []Filter{{"service", EQ, "yard"}, {"type", EQ, "rail"}}, Zoom: AllZoom, Properties: NewProperties("line-width", float64(1), "line-color", color.Color{0.0, 1.0, 0.5, 1.0, false}), order: 1},
		Rule{Layer: "roads", Attachment: "", Filters: []Filter{{"type", EQ, "rail"}}, Zoom: NewZoomRange(EQ, 17), Properties: NewProperties("line-width", float64(5), "line-color", color.Color{60.0, 1.0, 0.5, 1.0, false}), order: 3},
		Rule{Layer: "roads", Attachment: "", Filters: []Filter{}, Zoom: NewZoomRange(EQ, 17), Properties: NewProperties("line-width", float64(2), "line-color", color.Color{60.0, 1.0, 0.5, 1.0, false}), order: 2},
		Rule{Layer: "roads", Attachment: "", Filters: []Filter{}, Zoom: AllZoom, Properties: NewProperties("line-width", float64(1)), order: 2},
	})

	rules = loadRules(t, "tests/021-zoom-specific.mss", "roads")
	assertRulesEq(t, rules, []Rule{
		Rule{Layer: "roads", Attachment: "", Filters: []Filter{{"type", EQ, "primary"}}, Zoom: NewZoomRange(EQ, 15), Properties: NewProperties("line-width", float64(5), "line-color", color.Color{0.0, 1.0, 0.5, 1.0, false}, "line-cap", "round", "line-join", "bevel")},
		Rule{Layer: "roads", Attachment: "", Filters: []Filter{{"type", EQ, "primary"}}, Zoom: NewZoomRange(GTE, 14), Properties: NewProperties("line-color", color.Color{0.0, 1.0, 0.5, 1.0, false}, "line-cap", "round", "line-join", "bevel")},
		Rule{Layer: "roads", Attachment: "", Filters: []Filter{}, Zoom: NewZoomRange(EQ, 15), Properties: NewProperties("line-width", float64(5), "line-color", color.Color{0.0, 0.0, 1.0, 1.0, false}, "line-cap", "round", "line-join", "bevel")},
		Rule{Layer: "roads", Attachment: "", Filters: []Filter{}, Zoom: NewZoomRange(GTE, 14), Properties: NewProperties("line-color", color.Color{0.0, 0.0, 1.0, 1.0, false}, "line-cap", "round", "line-join", "bevel")},
		Rule{Layer: "roads", Attachment: "", Filters: []Filter{}, Zoom: AllZoom, Properties: NewProperties("line-join", "bevel")},
	})

}

func TestDecoderClasses(t *testing.T) {
	var rules []Rule
	rules = loadRules(t, "tests/014-classes.mss", "lakes", "land")
	assertRulesEq(t, rules, []Rule{
		Rule{Layer: "lakes", Class: "land", Attachment: "", Filters: []Filter{}, Zoom: AllZoom, Properties: NewProperties("line-width", float64(0.5), "line-color", color.Color{0.0, 1.0, 0.5, 1.0, false}, "polygon-fill", color.Color{240.0, 1.0, 0.5, 1.0, false})},
	})

	// basin class is inside water, no match
	rules = loadRules(t, "tests/014-classes.mss", "", "basin")
	assertRulesEq(t, rules, []Rule{})

	rules = loadRules(t, "tests/014-classes.mss", "", "water")
	assertRulesEq(t, rules, []Rule{
		Rule{Layer: "", Class: "water", Attachment: "", Filters: []Filter{}, Zoom: AllZoom, Properties: NewProperties("polygon-fill", color.Color{120.0, 1.0, 0.5, 1.0, false}, "line-width", float64(1))},
	})

	// return .water.basin property regardless of requested class order
	rules = loadRules(t, "tests/014-classes.mss", "", "basin", "water")
	assertRulesEq(t, rules, []Rule{
		Rule{Layer: "", Class: "basin", Attachment: "", Filters: []Filter{}, Zoom: AllZoom, Properties: NewProperties("polygon-fill", color.Color{0.0, 0.0, 1.0, 1.0, false}, "line-width", float64(1), "polygon-opacity", float64(0.5))},
	})
	rules = loadRules(t, "tests/014-classes.mss", "", "water", "basin")
	assertRulesEq(t, rules, []Rule{
		Rule{Layer: "", Class: "water", Attachment: "", Filters: []Filter{}, Zoom: AllZoom, Properties: NewProperties("polygon-fill", color.Color{0.0, 0.0, 1.0, 1.0, false}, "line-width", float64(1), "polygon-opacity", float64(0.5))},
	})

}
