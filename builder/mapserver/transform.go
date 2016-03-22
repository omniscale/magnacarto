package mapserver

import (
	"fmt"
	"regexp"
	"strconv"
)

type transformation struct {
	scale           float64
	rotate          float64
	hasRotateAnchor bool
	rotateAnchor    [2]float64
}

var svgTransformRe = regexp.MustCompile(`(rotate|scale)\(((-?\d*\.?\d+),? ?(-?\d*\.?\d+)?,? ?(-?\d*\.?\d+)?)\)`)

func parseTransform(transform string) (transformation, error) {
	tr := transformation{}
	for _, match := range svgTransformRe.FindAllStringSubmatch(transform, -1) {
		switch match[1] {
		case "rotate":
			if len(match) == 6 {
				tr.rotate, _ = strconv.ParseFloat(match[3], 64)
				tr.rotateAnchor[0], _ = strconv.ParseFloat(match[4], 64)
				tr.rotateAnchor[1], _ = strconv.ParseFloat(match[5], 64)
				tr.hasRotateAnchor = true
			} else {
				tr.rotate, _ = strconv.ParseFloat(match[2], 64)
			}
			tr.rotate *= -1
		case "scale":
			tr.scale, _ = strconv.ParseFloat(match[2], 64)
		default:
			return tr, fmt.Errorf("unsupported transform function %s in %s", match[1], transform)
		}
	}
	return tr, nil
}
