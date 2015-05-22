package builder

import "github.com/omniscale/magnacarto/mss"

func ZoomRange(rules []mss.Rule) mss.ZoomRange {
	z := mss.InvalidZoom
	for _, r := range rules {
		z = mss.ZoomRange(r.Zoom | z)
	}
	return z
}
