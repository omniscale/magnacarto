package mss

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

func newZoomRange(comp CompOp, zoom int64) ZoomRange {
	if zoom < 0 {
		return InvalidZoom
	}
	if zoom > 30 {
		return InvalidZoom
	}

	return AllZoom.add(comp, int8(zoom))
}

type CompOp int

const (
	Unknown CompOp = iota
	GT
	LT
	GTE
	LTE
	EQ
	NEQ
)

func (c CompOp) String() string {
	switch c {
	case EQ:
		return "="
	case GTE:
		return ">="
	case GT:
		return ">"
	case LTE:
		return "<="
	case LT:
		return "<"
	case NEQ:
		return "!="
	default:
		return "?"
	}
}

func parseCompOp(comp string) (CompOp, error) {
	switch comp {
	case "=":
		return EQ, nil
	case ">=":
		return GTE, nil
	case ">":
		return GT, nil
	case "<=":
		return LTE, nil
	case "<":
		return LT, nil
	case "!=":
		return NEQ, nil
	default:
		return Unknown, fmt.Errorf("unknown comparsion '%s'", comp)
	}
}

var AllZoom = ZoomRange(math.MaxInt32)
var InvalidZoom = ZoomRange(0)

type ZoomRange int32

func (z ZoomRange) validFor(level int) bool {
	return z>>uint8(level)&1 > 0
}

func (z ZoomRange) add(comp CompOp, level int8) ZoomRange {
	l := uint(level)
	switch comp {
	case EQ:
		return z & 1 << l
	case NEQ:
		return z & ^(1 << l)
	case LT:
		return z & ^(math.MaxInt32 << l)
	case LTE:
		return z & ^(math.MaxInt32 << (l + 1))
	case GT:
		return z & (math.MaxInt32 << (l + 1))
	case GTE:
		return z & (math.MaxInt32 << l)
	default:
		panic("unknown CompOp")
	}
}

func (z ZoomRange) combine(other ZoomRange) ZoomRange {
	return ZoomRange(other & z)
}

func (z ZoomRange) Levels() (n int) {
	// n accumulates the total bits set in x, counting only set bits
	for ; z > 0; n++ {
		// clear the least significant bit set
		z &= z - 1
	}
	return
}

func (z ZoomRange) String() string {
	if z == AllZoom {
		return "Zoom{*}"
	}
	op, l := z.simplify()
	if op != Unknown {
		return fmt.Sprintf("Zoom{%s%d}", op.String(), l)
	}
	zooms := []string{}
	for i := 0; i < 31; i++ {
		if z.validFor(i) {
			zooms = append(zooms, strconv.FormatInt(int64(i), 10))
		}
	}

	return "Zoom{" + strings.Join(zooms, " ") + "}"
}

func (z ZoomRange) First() int {
	first := 0
	for l := 0; l <= 30; l++ {
		if z>>uint8(l)&1 > 0 {
			first = l
			break
		}
	}
	return first
}
func (z ZoomRange) Last() int {
	last := 30
	for l := 30; l >= 0; l-- {
		if z>>uint8(l)&1 > 0 {
			last = l
			break
		}
	}
	return last
}

func (z ZoomRange) simplify() (CompOp, int) {
	if z == InvalidZoom || z == AllZoom {
		return Unknown, 0
	}

	first := z.First()
	last := z.Last()

	if last == first {
		return EQ, last
	}
	if first == 0 {
		return LTE, last
	}
	if last == 30 {
		return GTE, first
	}
	return Unknown, 0
}
