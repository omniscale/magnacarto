package sql

import (
	"strings"

	"github.com/omniscale/magnacarto/mss"
)

func filterItems(rules []mss.Rule) map[string]map[string]struct{} {
	result := make(map[string]map[string]struct{})
	for _, r := range rules {
		if len(r.Filters) == 0 {
			return nil
		}

		// rule filters are conjunct (AND), it's ok if we can limit the rows
		// by matching at least one filter.
		found := false
		for _, f := range r.Filters {
			if f.CompOp != mss.EQ {
				continue
			}
			v, ok := f.Value.(string)
			if !ok {
				continue
			}
			found = true
			if result[f.Field] == nil {
				result[f.Field] = make(map[string]struct{})
			}
			result[f.Field][v] = struct{}{}
		}
		if !found {
			return nil
		}
	}
	return result
}

// FilterString returns an SQL WHERE statement that pre-filters rows
// for this set of rules. Only supports "equal strings" filters.
//
// Rules for:
//   #foo [type='bar'][level=2] {}
//   #foo [type='baz'][level=2] {}
// will return
//   "(type IN ('bar', 'baz'))"
//
// Rules for:
//   #foo [level=2] {}
//   #foo [type='bar'][level=2] {}
// will return an empty string, since any 'type' can match.
func FilterString(rules []mss.Rule) string {
	var vals, parts []string

	items := filterItems(rules)

	for key, values := range items {
		vals = vals[:0]
		for v, _ := range values {
			vals = append(vals, "'"+v+"'")
		}
		parts = append(parts, "\""+key+"\" IN ("+strings.Join(vals, ", ")+")")
	}
	if parts == nil {
		return ""
	}
	return "(" + strings.Join(parts, " OR ") + ")"
}

func WrapWhere(query, where string) string {
	if where == "" {
		return query
	}
	return "(SELECT * FROM " + query + " WHERE " + where + ") as filtered"
}
