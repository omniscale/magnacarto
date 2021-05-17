package mapserver

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/omniscale/magnacarto/color"
	"github.com/omniscale/magnacarto/mss"
)

func quote(v string) string {
	return `"` + v + `"`
}

func fmtKeyword(v mss.Value, ok bool) *string {
	if !ok {
		return nil
	}
	switch v := v.(type) {
	case string:
		r := strings.ToUpper(v)
		return &r
	case nil:
		return nil
	default:
		r := fmt.Sprintf("unknown type %T for %v", v, v)
		return &r
	}
}

// fmtFloatProp formats a scaled float property. Returns nil if property is not set or ont a float.
func fmtFloatProp(p *mss.Properties, name string, scale float64) *string {
	v, ok := p.GetFloat(name)
	if !ok {
		return nil
	}
	r := strconv.FormatFloat(v*scale, 'f', -1, 64)
	return &r
}

func fmtFloat(v float64, ok bool) *string {
	if !ok {
		return nil
	}
	r := strconv.FormatFloat(v, 'f', -1, 64)
	return &r
}

func fmtString(v string, ok bool) *string {
	if !ok {
		return nil
	}
	if len(v) > 2 && v[0] == '\'' && v[len(v)-1] == '\'' {
		v = v[1 : len(v)-1]
	}
	r := "'" + escapeSingleQuote(v) + "'"
	return &r
}

func fmtBool(v bool, ok bool) *string {
	if !ok {
		return nil
	}
	var r string
	if v {
		r = "true"
	} else {
		r = "false"
	}
	return &r
}

func fmtColor(v color.Color, ok bool) *string {
	if !ok {
		return nil
	}
	var r string
	r = "\"" + v.HexString() + "\""
	return &r
}

func fmtFilters(filters []mss.Filter) string {
	parts := []string{}
	for _, f := range filters {
		field := "[" + f.Field + "]"

		var value string
		switch v := f.Value.(type) {
		case nil:
			value = "null"
		case string:
			value = "'" + escapeSingleQuote(v) + "'"
			// field needs to be quoted if we compare strings
			// e.g. ('[field]' = "foo"), but ([field] = 5)
			field = "'" + field + "'"
		case float64:
			value = string(*fmtFloat(v, true))
		case mss.ModuloComparsion:
			value = fmt.Sprintf("%d %s %d", v.Div, v.CompOp, v.Value)
		default:
			log.Printf("unknown type of filter value: %s", v)
			value = ""
		}
		if f.CompOp == mss.REGEX {
			parts = append(parts, "("+field+" ~ "+value+")")
		} else {
			parts = append(parts, "("+field+" "+f.CompOp.String()+" "+value+")")
		}
	}

	s := strings.Join(parts, " AND ")
	if len(filters) > 1 {
		s = "(" + s + ")"
	}
	return s
}

func fmtPattern(v []float64, scale float64, ok bool) *Block {
	if !ok {
		return nil
	}
	b := NewBlock("PATTERN")
	for i := range v {
		b.Add("", *fmtFloat(v[i]*LineWidthFactor*scale, true))
	}
	return &b
}

// fmtFieldString formats a list of fields as single quoted string, eg. 'Foo: [name]'
func fmtFieldString(vals []interface{}, ok bool) *string {
	if !ok {
		return nil
	}
	parts := []string{}
	// TODO: improve testing for this, i'm sure this will fail with more complex field expressions
	for _, v := range vals {
		switch v.(type) {
		case mss.Field:
			parts = append(parts, escapeSingleQuote(string(v.(mss.Field))))
		case string:
			parts = append(parts, escapeSingleQuote(v.(string)))
		}
	}
	r := "'" + strings.Join(parts, "") + "'"
	return &r
}

// fmtField formats a field as a single attribute, eg. [angle]
func fmtField(vals []interface{}, ok bool) *string {
	if !ok {
		return nil
	}
	var r string
	if len(vals) != 1 {
		return nil
	}
	switch vals[0].(type) {
	case mss.Field:
		r = escapeSingleQuote(string(vals[0].(mss.Field)))
	case string:
		r = escapeSingleQuote(vals[0].(string))
	}
	return &r
}

func escapeSingleQuote(str string) string {
	return strings.Replace(str, "'", "\\'", -1)
}

// indent inserts prefix at the beginning of each non-empty line of s. The
// end-of-line marker is NL.
func indent(s, prefix string) string {
	return string(indentBytes([]byte(s), []byte(prefix)))
}

// indentBytes inserts prefix at the beginning of each non-empty line of b.
// The end-of-line marker is NL.
func indentBytes(b, prefix []byte) []byte {
	var res []byte
	bol := true
	for _, c := range b {
		if bol && c != '\n' {
			res = append(res, prefix...)
		}
		res = append(res, c)
		bol = c == '\n'
	}
	return res
}
