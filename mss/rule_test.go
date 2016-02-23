package mss

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLayerRules(t *testing.T) {
	files, err := filepath.Glob("tests/*.mss")
	if err != nil {
		t.Fatal(err)
	}

	for _, f := range files {
		r, err := os.Open(f)
		if err != nil {
			t.Fatal(err)
		}
		bytes, err := ioutil.ReadAll(r)
		if err != nil {
			t.Fatal(err)
		}
		d, err := decodeString(string(bytes))
		if err != nil {
			t.Fatal(err)
		}

		layers := map[string]struct{}{}
		for _, b := range d.mss.root.blocks {
			for _, s := range b.selectors {
				layers[s.Layer] = struct{}{}
			}
		}
		for layer, _ := range layers {
			d.MSS().LayerRules(layer)
		}
		r.Close()
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestRuleSame(t *testing.T) {
	assert.True(t, Rule{Layer: "Foo"}.same(Rule{Layer: "Foo"}))
	assert.False(t, Rule{Layer: "Foo", Attachment: "Bar"}.same(Rule{Layer: "Foo"}))
	assert.False(t, Rule{Layer: "Foo"}.same(Rule{Layer: "Foo", Attachment: "Bar"}))

	assert.True(t, Rule{Zoom: newZoomRange(EQ, 5)}.same(Rule{Zoom: newZoomRange(EQ, 5)}))
	assert.True(t, Rule{Zoom: newZoomRange(LT, 5)}.same(Rule{Zoom: newZoomRange(LTE, 4)}))
	assert.False(t, Rule{Zoom: newZoomRange(LT, 5)}.same(Rule{Zoom: newZoomRange(LTE, 5)}))

	assert.True(t, Rule{
		Layer: "Bar", Attachment: "Foo", Filters: []Filter{Filter{"foo", EQ, "foo"}},
	}.same(Rule{
		Layer: "Bar", Attachment: "Foo", Filters: []Filter{Filter{"foo", EQ, "foo"}}},
	))

	assert.True(t, Rule{
		Layer: "Bar", Attachment: "Foo", Filters: []Filter{Filter{"foo", EQ, "foo"}}, Zoom: newZoomRange(EQ, 4),
	}.childOf(Rule{
		Layer: "Bar", Attachment: "Foo", Filters: []Filter{Filter{"foo", EQ, "foo"}}, Zoom: newZoomRange(EQ, 4)},
	))
	assert.False(t, Rule{
		Layer: "Bar", Attachment: "Foo", Filters: []Filter{Filter{"foo", EQ, "foo"}}, Zoom: newZoomRange(EQ, 4),
	}.childOf(Rule{
		Layer: "Bar", Attachment: "Foo", Filters: []Filter{Filter{"foo", EQ, "foo"}}, Zoom: newZoomRange(EQ, 5)},
	))

}

func TestRuleChildOf(t *testing.T) {
	assert.True(t, Rule{Layer: "roads", Zoom: AllZoom}.childOf(Rule{Layer: "roads", Zoom: AllZoom}))
	assert.False(t, Rule{Layer: "roads", Zoom: AllZoom}.childOf(Rule{Layer: "landusage", Zoom: AllZoom}))
	assert.True(t, Rule{Layer: "roads"}.childOf(Rule{Layer: "roads"}))
	assert.False(t, Rule{Layer: "roads"}.childOf(Rule{Layer: "landusage"}))
	assert.True(t, Rule{Layer: "roads", Attachment: "inline"}.childOf(Rule{Layer: "roads"}))
	assert.False(t, Rule{Layer: "roads", Attachment: "inline"}.childOf(Rule{Layer: "Baz"}))
	assert.True(t, Rule{Layer: "roads", Attachment: "inline"}.childOf(Rule{Layer: "roads", Attachment: "inline"}))
	assert.False(t, Rule{Layer: "roads", Attachment: "inline"}.childOf(Rule{Layer: "roads", Attachment: "coutline"}))

	assert.True(t, Rule{
		Layer: "roads", Attachment: "inline", Filters: []Filter{Filter{"highway", EQ, "path"}},
	}.childOf(Rule{Layer: "roads", Attachment: "inline"}))

	assert.True(t, Rule{
		Layer: "roads", Attachment: "inline", Filters: []Filter{Filter{"highway", EQ, "path"}},
	}.childOf(Rule{
		Layer: "roads", Attachment: "inline", Filters: []Filter{Filter{"highway", EQ, "path"}}},
	))

	assert.True(t, Rule{
		Layer: "roads", Attachment: "inline", Filters: []Filter{Filter{"highway", EQ, "path"}},
	}.childOf(Rule{
		Layer: "roads", Attachment: "inline", Filters: []Filter{Filter{"highway", EQ, "path"}}},
	))

	assert.True(t, Rule{
		Layer: "roads", Attachment: "inline", Filters: []Filter{Filter{"highway", EQ, "path"}, Filter{"oneway", EQ, "yes"}},
	}.childOf(Rule{
		Layer: "roads", Attachment: "inline", Filters: []Filter{Filter{"highway", EQ, "path"}}},
	))

	assert.True(t, Rule{Zoom: newZoomRange(EQ, 4)}.childOf(Rule{Zoom: newZoomRange(LTE, 5)}))
	assert.True(t, Rule{Zoom: newZoomRange(EQ, 5)}.childOf(Rule{Zoom: newZoomRange(LTE, 5)}))
	assert.False(t, Rule{Zoom: newZoomRange(EQ, 6)}.childOf(Rule{Zoom: newZoomRange(LTE, 5)}))
}

func TestRuleAffectedBy(t *testing.T) {
	assert.True(t, Rule{Layer: "roads", Zoom: AllZoom}.overlaps(Rule{Layer: "roads", Zoom: AllZoom}))
	assert.False(t, Rule{Layer: "roads", Zoom: AllZoom}.overlaps(Rule{Layer: "landusage", Zoom: AllZoom}))
	assert.True(t, Rule{Layer: "roads", Attachment: "inline", Zoom: AllZoom}.overlaps(Rule{Layer: "roads", Zoom: AllZoom}))
	assert.False(t, Rule{Layer: "roads"}.overlaps(Rule{Layer: "landusage"}))

	assert.False(t, Rule{Layer: "roads", Zoom: AllZoom}.overlaps(Rule{Layer: "landusage", Zoom: AllZoom}))

	assert.True(t, Rule{Layer: "roads", Attachment: "inline"}.overlaps(Rule{Layer: "roads"}))
	assert.False(t, Rule{Layer: "roads", Attachment: "inline"}.overlaps(Rule{Layer: "Baz"}))
	assert.True(t, Rule{Layer: "roads", Attachment: "inline"}.overlaps(Rule{Layer: "roads", Attachment: "inline"}))
	assert.False(t, Rule{Layer: "roads", Attachment: "inline"}.overlaps(Rule{Layer: "roads", Attachment: "coutline"}))

	assert.True(t, Rule{
		Layer: "roads", Attachment: "inline", Filters: []Filter{Filter{"highway", EQ, "path"}},
	}.overlaps(Rule{Layer: "roads", Attachment: "inline"}))

	assert.True(t, Rule{
		Layer: "roads", Attachment: "inline", Filters: []Filter{Filter{"highway", EQ, "path"}}, Zoom: newZoomRange(EQ, 5),
	}.overlaps(Rule{Layer: "roads", Attachment: "inline", Zoom: AllZoom}))

	assert.True(t, Rule{
		Layer: "roads", Attachment: "inline", Filters: []Filter{Filter{"highway", EQ, "path"}}, Zoom: newZoomRange(EQ, 5),
	}.overlaps(Rule{Layer: "roads", Attachment: "inline", Zoom: newZoomRange(LTE, 5)}))

	assert.False(t, Rule{
		Layer: "roads", Attachment: "inline", Filters: []Filter{Filter{"highway", EQ, "path"}}, Zoom: newZoomRange(EQ, 5),
	}.overlaps(Rule{Layer: "roads", Attachment: "inline", Zoom: newZoomRange(LT, 5)}))

	assert.True(t, Rule{
		Layer: "roads", Attachment: "inline", Filters: []Filter{Filter{"oneway", EQ, "yes"}, Filter{"highway", EQ, "path"}},
		Zoom: newZoomRange(EQ, 4),
	}.overlaps(Rule{
		Layer: "roads", Attachment: "inline", Filters: []Filter{Filter{"highway", EQ, "path"}},
		Zoom: newZoomRange(GT, 3)},
	))
	// same with different filter order
	assert.True(t, Rule{
		Layer: "roads", Attachment: "inline", Filters: []Filter{Filter{"highway", EQ, "path"}, Filter{"oneway", EQ, "yes"}},
		Zoom: newZoomRange(EQ, 4),
	}.overlaps(Rule{
		Layer: "roads", Attachment: "inline", Filters: []Filter{Filter{"highway", EQ, "path"}},
		Zoom: newZoomRange(GT, 3)},
	))

	assert.False(t, Rule{
		Layer: "roads", Attachment: "inline", Filters: []Filter{Filter{"oneway", EQ, "yes"}, Filter{"highway", EQ, "path"}},
		Zoom: newZoomRange(EQ, 4),
	}.overlaps(Rule{
		Layer: "roads", Attachment: "inline", Filters: []Filter{Filter{"highway", EQ, "path"}},
		Zoom: newZoomRange(LT, 4)},
	))
	// same with different filter order
	assert.False(t, Rule{
		Layer: "roads", Attachment: "inline", Filters: []Filter{Filter{"highway", EQ, "path"}, Filter{"oneway", EQ, "yes"}},
		Zoom: newZoomRange(EQ, 4),
	}.overlaps(Rule{
		Layer: "roads", Attachment: "inline", Filters: []Filter{Filter{"highway", EQ, "path"}},
		Zoom: newZoomRange(LT, 4)},
	))

	assert.True(t, Rule{Zoom: newZoomRange(EQ, 4)}.overlaps(Rule{Zoom: newZoomRange(LTE, 5)}))
	assert.True(t, Rule{Zoom: newZoomRange(EQ, 5)}.overlaps(Rule{Zoom: newZoomRange(LTE, 5)}))
	assert.False(t, Rule{Zoom: newZoomRange(EQ, 6)}.overlaps(Rule{Zoom: newZoomRange(LTE, 5)}))
}

func TestFilterIsSubset(t *testing.T) {
	assert.True(t, filterIsSubset(nil, nil))
	assert.True(t, filterIsSubset(nil, []Filter{Filter{"foo", EQ, "bar"}}))
	assert.True(t, filterIsSubset([]Filter{Filter{"foo", EQ, "bar"}}, []Filter{Filter{"foo", EQ, "bar"}}))
	assert.True(t, filterIsSubset([]Filter{Filter{"foo", EQ, "bar"}}, []Filter{Filter{"baz", EQ, "bar"}, Filter{"foo", EQ, "bar"}}))
	assert.False(t, filterIsSubset([]Filter{Filter{"foo", EQ, "barbaz"}}, []Filter{Filter{"baz", EQ, "bar"}, Filter{"foo", EQ, "bar"}}))
}

func TestFilterIsSubset_TODO(t *testing.T) {
	t.Skip("TODO: implement filterIsSubset for numerical comparsions")
	assert.True(t, filterIsSubset([]Filter{Filter{"foo", GTE, 5}}, []Filter{Filter{"foo", EQ, 5}}))
	assert.True(t, filterIsSubset([]Filter{Filter{"foo", GT, 4}}, []Filter{Filter{"foo", EQ, 5}}))
	assert.True(t, filterIsSubset([]Filter{Filter{"foo", LTE, 5}}, []Filter{Filter{"foo", EQ, 5}}))
	assert.True(t, filterIsSubset([]Filter{Filter{"foo", LT, 6}}, []Filter{Filter{"foo", EQ, 5}}))

	assert.False(t, filterIsSubset([]Filter{Filter{"foo", GTE, 6}}, []Filter{Filter{"foo", EQ, 5}}))
	assert.False(t, filterIsSubset([]Filter{Filter{"foo", GT, 5}}, []Filter{Filter{"foo", EQ, 5}}))
	assert.False(t, filterIsSubset([]Filter{Filter{"foo", LTE, 4}}, []Filter{Filter{"foo", EQ, 5}}))
	assert.False(t, filterIsSubset([]Filter{Filter{"foo", LT, 5}}, []Filter{Filter{"foo", EQ, 5}}))

	assert.True(t, filterIsSubset([]Filter{Filter{"foo", GTE, 5}}, []Filter{Filter{"foo", GT, 5}}))
	assert.True(t, filterIsSubset([]Filter{Filter{"foo", GT, 4}}, []Filter{Filter{"foo", GT, 5}}))
	assert.False(t, filterIsSubset([]Filter{Filter{"foo", LT, 10}}, []Filter{Filter{"foo", GT, 5}}))
	assert.False(t, filterIsSubset([]Filter{Filter{"foo", LTE, 10}}, []Filter{Filter{"foo", GT, 5}}))
	assert.False(t, filterIsSubset([]Filter{Filter{"foo", EQ, 6}}, []Filter{Filter{"foo", GT, 5}}))

	assert.True(t, filterIsSubset([]Filter{Filter{"foo", LTE, 5}}, []Filter{Filter{"foo", LT, 5}}))
	assert.True(t, filterIsSubset([]Filter{Filter{"foo", LT, 5}}, []Filter{Filter{"foo", LT, 5}}))
}

func BenchmarkFilterIsSubset(b *testing.B) {
	for i := 0; i < b.N; i++ {
		filterIsSubset(
			[]Filter{Filter{"foo1", EQ, "bar"}, Filter{"foo2", EQ, "bar"}, Filter{"foo3", EQ, "bar"}},
			[]Filter{Filter{"baz", EQ, "bar"}, Filter{"baz", EQ, "bar"}, Filter{"foo1", EQ, "bar"}, Filter{"foo2", EQ, "bar"}, Filter{"foo3", EQ, "bar"}},
		)
	}
}

func TestCombineRule(t *testing.T) {
	combined := combineRules(Rule{
		Layer: "roads", Attachment: "inline", Filters: []Filter{{"highway", EQ, "path"}, {"oneway", EQ, "yes"}},
		Zoom: newZoomRange(LT, 5), Properties: NewProperties("width", 1, "cap", "round"),
	}, Rule{
		Layer: "roads", Attachment: "inline", Filters: []Filter{{"bridge", EQ, 1}, {"highway", EQ, "path"}},
		Zoom: newZoomRange(GT, 3), Properties: NewProperties("width", 4, "fill", "red", "end", "butt"),
	})

	assertRuleEq(t, Rule{
		Layer: "roads", Attachment: "inline", Filters: []Filter{{"bridge", EQ, 1}, {"highway", EQ, "path"}, {"oneway", EQ, "yes"}},
		Zoom: newZoomRange(EQ, 4), Properties: NewProperties("width", 4, "fill", "red", "cap", "round", "end", "butt")},
		combined,
	)
}

func TestSortedRulesNoInfinitiveLoop(t *testing.T) {
	// check for bug were r.overlaps(o) returns rule identical with o which resulted
	// in more rules that returned identical rules -> infinite loop
	rules := []Rule{
		{Filters: []Filter{{"bar", EQ, "foo"}}, Zoom: AllZoom, Properties: NewProperties("width", 1)},
		{Filters: []Filter{}, Zoom: newZoomRange(GTE, 12), Properties: NewProperties("width", 1)},
		{Filters: []Filter{}, Zoom: newZoomRange(GTE, 13), Properties: NewProperties("width", 1)},
		{Filters: []Filter{}, Zoom: newZoomRange(GTE, 14), Properties: NewProperties("width", 1)},
		{Filters: []Filter{}, Zoom: newZoomRange(GTE, 15), Properties: NewProperties("width", 1)},
	}
	sorted := sortedRules(rules, nil, nil)
	assert.Len(t, sorted, 9)
}

func TestSortedRulesCombinationDuplicates(t *testing.T) {
	rules := []Rule{
		{Filters: []Filter{{"a", EQ, "1"}}, Properties: NewProperties("a", 1)},
		{Filters: []Filter{{"b", EQ, "1"}}, Properties: NewProperties("b", 1)},
		{Filters: []Filter{{"c", EQ, "1"}}, Properties: NewProperties("c", 1)},
	}
	sorted := sortedRules(rules, nil, nil)
	assert.Len(t, sorted, 7)

	rules = []Rule{
		{Filters: []Filter{{"a", EQ, "1"}}, Properties: NewProperties("a", 1)},
		{Filters: []Filter{{"b", EQ, "1"}}, Properties: NewProperties("b", 1)},
		{Filters: []Filter{{"c", EQ, "1"}}, Properties: NewProperties("c", 1)},
		{Filters: []Filter{{"a", EQ, "1"}, {"b", EQ, "1"}}, Properties: NewProperties("b", 2, "a", 2)},
	}
	sorted = sortedRules(rules, nil, nil)
	assert.Len(t, sorted, 7)
}

func TestSortedRulesMultipleClasses(t *testing.T) {
	// .A [a=1] { a: 1 }
	// .B [b=1] { b: 1 }
	rules := []Rule{
		{Class: "A", Filters: []Filter{{"a", EQ, 1}}, Properties: NewProperties("a", 1)},
		{Class: "B", Filters: []Filter{{"b", EQ, 1}}, Properties: NewProperties("b", 1)},
	}

	sorted := sortedRules(rules, nil, nil)
	assert.Len(t, sorted, 3)
	assertRuleEq(t, Rule{
		Class:      "A", // TODO A, B
		Filters:    []Filter{{"a", EQ, 1}, {"b", EQ, 1}},
		Properties: NewProperties("a", 1, "b", 1)},
		sorted[0],
	)
	assertRuleEq(t, Rule{
		Class:      "A",
		Filters:    []Filter{{"a", EQ, 1}},
		Properties: NewProperties("a", 1)},
		sorted[1],
	)
	assertRuleEq(t, Rule{
		Class:      "B",
		Filters:    []Filter{{"b", EQ, 1}},
		Properties: NewProperties("b", 1)},
		sorted[2],
	)
}

func TestSortedRulesMultipleClassesInstance(t *testing.T) {
	// .A::X [a=1] { a: 1}
	// .B::X [a=1][b=2] { b/b: 1}
	rules := []Rule{
		{Class: "A", Attachment: "X", Filters: []Filter{{"a", EQ, 1}}, Properties: NewPropertiesInstance("a", "", 1)},
		{Class: "B", Attachment: "X", Filters: []Filter{{"a", EQ, 1}, {"b", EQ, 2}}, Properties: NewPropertiesInstance("b", "b", 1)},
	}

	sorted := sortedRules(rules, nil, nil)
	assert.Len(t, sorted, 2)
	assertRuleEq(t, Rule{
		Class:      "B", // TODO A, B
		Attachment: "X",
		Filters:    []Filter{{"a", EQ, 1}, {"b", EQ, 2}},
		Properties: NewPropertiesInstance("a", "", 1, "b", "b", 1)},
		sorted[0],
	)
	assertRuleEq(t, Rule{
		Class:      "A",
		Attachment: "X",
		Filters:    []Filter{{"a", EQ, 1}},
		Properties: NewPropertiesInstance("a", "", 1)},
		sorted[1],
	)
}

func TestSortedRulesCombination(t *testing.T) {
	// disjoint filters will result in combination of all rules
	rules := []Rule{
		{Filters: []Filter{{"type", EQ, "road"}}, Properties: NewProperties("width", 1)},
		{Filters: []Filter{{"tunnel", EQ, 1}}, Properties: NewProperties("dash-array", 1)},
		{Filters: []Filter{{"access", EQ, "private"}}, Properties: NewProperties("color", "grey")},
	}
	sorted := sortedRules(rules, nil, nil)
	assert.Len(t, sorted, 7)

	assertRuleEq(t, Rule{
		Filters:    []Filter{{"access", EQ, "private"}, {"tunnel", EQ, 1}, {"type", EQ, "road"}},
		Properties: NewProperties("width", 1, "dash-array", 1, "color", "grey")},
		sorted[0],
	)

	assertRuleEq(t, Rule{
		Filters:    []Filter{{"tunnel", EQ, 1}, {"type", EQ, "road"}},
		Properties: NewProperties("width", 1, "dash-array", 1)},
		sorted[1],
	)

	assertRuleEq(t, Rule{
		Filters:    []Filter{{"access", EQ, "private"}, {"type", EQ, "road"}},
		Properties: NewProperties("width", 1, "color", "grey")},
		sorted[2],
	)

	assertRuleEq(t, Rule{
		Filters:    []Filter{{"type", EQ, "road"}},
		Properties: NewProperties("width", 1)},
		sorted[3],
	)

	assertRuleEq(t, Rule{
		Filters:    []Filter{{"access", EQ, "private"}, {"tunnel", EQ, 1}},
		Properties: NewProperties("dash-array", 1, "color", "grey")},
		sorted[4],
	)
	assertRuleEq(t, Rule{
		Filters:    []Filter{{"tunnel", EQ, 1}},
		Properties: NewProperties("dash-array", 1)},
		sorted[5],
	)

	assertRuleEq(t, Rule{
		Filters:    []Filter{{"access", EQ, "private"}},
		Properties: NewProperties("color", "grey")},
		sorted[6],
	)
}

func TestSortedRulesRedundantRuleRemoved(t *testing.T) {
	// [size > 1000],
	// [zoom >= 17]
	// {
	//   [size > 2000] { line-width: 2; }
	//   line-width: 1;
	// }
	rules := []Rule{
		{Filters: []Filter{{"size", GT, 1000}}, Properties: NewProperties("width", 1)},
		{Zoom: newZoomRange(GTE, 17), Properties: NewProperties("width", 1)},
		{Filters: []Filter{{"size", GT, 2000}}, Properties: NewProperties("width", 2)},
		{Zoom: newZoomRange(GTE, 17), Filters: []Filter{{"size", GT, 2000}}, Properties: NewProperties("width", 2)},
	}
	sorted := sortedRules(rules, nil, nil)
	assert.Len(t, sorted, 4)

	assertRuleEq(t, Rule{
		Filters:    []Filter{{"size", GT, 1000}},
		Properties: NewProperties("width", 1)},
		sorted[0],
	)
	assertRuleEq(t, Rule{
		Filters:    []Filter{{"size", GT, 2000}},
		Properties: NewProperties("width", 2)},
		sorted[1],
	)
	// TODO this rule is redundant since it is covered by the rule above
	assertRuleEq(t, Rule{
		Filters:    []Filter{{"size", GT, 2000}},
		Zoom:       newZoomRange(GTE, 17),
		Properties: NewProperties("width", 2)},
		sorted[2],
	)

	assertRuleEq(t, Rule{
		Zoom:       newZoomRange(GTE, 17),
		Properties: NewProperties("width", 1)},
		sorted[3],
	)
}

func TestMergeFilter(t *testing.T) {
	if f, ok := mergeFilter(Filter{"foo", EQ, "bar"}, Filter{"bar", EQ, "bar"}); ok {
		t.Error("different filters should not be merged to:", f)
	}
	if f, ok := mergeFilter(Filter{"foo", EQ, "bar"}, Filter{"foo", EQ, "bar1"}); ok {
		t.Error("different filters should not be merged to:", f)
	}
	if f, ok := mergeFilter(Filter{"foo", EQ, "bar"}, Filter{"foo", NEQ, "bar"}); ok {
		t.Error("different filters should not be merged to:", f)
	}

	if _, ok := mergeFilter(Filter{"foo", EQ, "bar"}, Filter{"foo", EQ, "bar"}); !ok {
		t.Error("same filters should merge")
	}

	if f, ok := mergeFilter(Filter{"foo", GT, 3.0}, Filter{"foo", GT, 1.0}); !ok || f.CompOp != GTE || f.Value.(float64) != 4 {
		t.Error("same filters should merge to", f)
	}
	if f, ok := mergeFilter(Filter{"foo", GTE, 4.0}, Filter{"foo", GT, 1.0}); !ok || f.CompOp != GTE || f.Value.(float64) != 4 {
		t.Error("same filters should merge to", f)
	}
	if f, ok := mergeFilter(Filter{"foo", LT, 3.0}, Filter{"foo", LT, 5.0}); !ok || f.CompOp != LTE || f.Value.(float64) != 2 {
		t.Error("same filters should merge to", f)
	}
	if f, ok := mergeFilter(Filter{"foo", LT, 3.0}, Filter{"foo", LTE, 6.0}); !ok || f.CompOp != LTE || f.Value.(float64) != 2 {
		t.Error("same filters should merge to", f)
	}
}

func TestMergeFilters(t *testing.T) {
	var result []Filter
	var ok bool

	result, ok = mergeFilters([]Filter{}, []Filter{})
	if !ok || len(result) != 0 {
		t.Error("error merging empty filters", result)
	}

	result, ok = mergeFilters([]Filter{{"A", EQ, "A"}}, []Filter{})
	if !ok || len(result) != 1 {
		t.Error("error merging filters", result)
	}
	result, ok = mergeFilters([]Filter{{"A", EQ, "A"}, {"B", EQ, "B"}, {"C", EQ, "C"}}, []Filter{{"A", EQ, "A"}})
	if !ok || len(result) != 3 || result[0].Field != "A" || result[1].Field != "B" || result[2].Field != "C" {
		t.Error("error merging filters", result)
	}

	result, ok = mergeFilters([]Filter{{"A", EQ, "A"}}, []Filter{{"A", EQ, "A"}, {"B", EQ, "B"}, {"C", EQ, "C"}})
	if !ok || len(result) != 3 || result[0].Field != "A" || result[1].Field != "B" || result[2].Field != "C" {
		t.Error("error merging filters", result)
	}

	result, ok = mergeFilters([]Filter{{"A", GTE, 2.0}}, []Filter{{"A", GTE, 4.0}, {"B", EQ, "B"}, {"C", EQ, "C"}})
	if !ok || len(result) != 3 || result[0].Field != "A" || result[0].CompOp != GTE || result[0].Value.(float64) != 4 || result[1].Field != "B" || result[2].Field != "C" {
		t.Error("error merging filters", result)
	}
}
