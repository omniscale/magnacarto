package mapserver

import (
	"fmt"
	"regexp"
	"strconv"
)

type transformation struct {
	scale  float64
	rotate float64
}

var svgTransformRe = regexp.MustCompile(`(rotate|scale)\((-?\d*\.?\d+)\)`)

func parseTransform(transform string) (transformation, error) {
	tr := transformation{}
	for _, match := range svgTransformRe.FindAllStringSubmatch(transform, -1) {
		switch match[1] {
		case "rotate":
			tr.rotate, _ = strconv.ParseFloat(match[2], 64)
		case "scale":
			tr.scale, _ = strconv.ParseFloat(match[2], 64)
		default:
			return tr, fmt.Errorf("unsupported transform function %s in %s", match[1], transform)
		}
	}
	return tr, nil
}
