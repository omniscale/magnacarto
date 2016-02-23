package mapserver

import (
	"github.com/omniscale/magnacarto/color"
	"github.com/omniscale/magnacarto/mss"
	"github.com/stretchr/testify/assert"

	"testing"
)

func TestFmtColor(t *testing.T) {
	assert.Nil(t, fmtColor(color.Color{}, false))
	assert.Equal(t, `"#ff0000"`, *fmtColor(color.MustParse("red"), true))
	c := color.MustParse("red")
	c.A = 0.5
	assert.Equal(t, `"#ff000080"`, *fmtColor(c, true))
}

func TestFmtKeyword(t *testing.T) {
	assert.Nil(t, fmtKeyword("", false))
	assert.Equal(t, `AUTO`, *fmtKeyword("auto", true))
}

func TestFmtFloat(t *testing.T) {
	assert.Nil(t, fmtFloat(1.23, false))
	assert.Equal(t, `1.23`, *fmtFloat(1.23, true))
}

func TestFmtString(t *testing.T) {
	assert.Nil(t, fmtString("test", false))
	assert.Equal(t, `'test'`, *fmtString("test", true))
	assert.Equal(t, `'test'`, *fmtString("'test'", true))
	assert.Equal(t, `'"\'te"st\''`, *fmtString(`"'te"st'`, true))
}

func TestFmtFilters(t *testing.T) {
	assert.Empty(t, fmtFilters(nil))
	assert.Equal(t, `([name] = 5)`, fmtFilters([]mss.Filter{{Field: "name", CompOp: mss.EQ, Value: 5.0}}))
	assert.Equal(t, `('[type]' = 'residential')`, fmtFilters([]mss.Filter{{Field: "type", CompOp: mss.EQ, Value: "residential"}}))
	assert.Equal(t, `([type] = null)`, fmtFilters([]mss.Filter{{Field: "type", CompOp: mss.EQ, Value: nil}}))
	assert.Equal(t, `([foo] >= 2)`, fmtFilters([]mss.Filter{{Field: "foo", CompOp: mss.GTE, Value: 2.0}}))
	assert.Equal(t, `('[foo]' ~ '^bar')`, fmtFilters([]mss.Filter{{Field: "foo", CompOp: mss.REGEX, Value: "^bar"}}))

	assert.Equal(t, `(('[type]' = 'residential') AND ('[foo]' ~ '^bar'))`, fmtFilters(
		[]mss.Filter{
			{Field: "type", CompOp: mss.EQ, Value: "residential"},
			{Field: "foo", CompOp: mss.REGEX, Value: "^bar"},
		},
	))
}

func TestFmtFieldString(t *testing.T) {
	assert.Nil(t, fmtFieldString(nil, false))
	assert.Equal(t, `'Name: [name]'`, *fmtFieldString([]interface{}{"Name: ", mss.Field("[name]")}, true))
	assert.Equal(t, `'[foo] = \'[name]\''`, *fmtFieldString([]interface{}{mss.Field("[foo]"), " = '", mss.Field("[name]"), "'"}, true))

}
