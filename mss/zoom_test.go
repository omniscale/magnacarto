package mss

import (
	"fmt"
	"math"

	"github.com/stretchr/testify/assert"

	"testing"
)

func checkZoomIncludes(t *testing.T, zoom ZoomRange, includes []int) {
	for _, l := range includes {
		assert.True(t, zoom.validFor(l), fmt.Sprintf("%d in %v", l, zoom))
	}
excludeCheck:
	for l := 0; l < 30; l++ {
		for _, incl := range includes {
			if incl == l {
				continue excludeCheck
			}
		}
		assert.False(t, zoom.validFor(l))
	}
}

func checkZoomExcludes(t *testing.T, zoom ZoomRange, excludes []int) {
	for _, l := range excludes {
		assert.False(t, zoom.validFor(l))
	}
includeCheck:
	for l := 0; l < 30; l++ {
		for _, excl := range excludes {
			if excl == l {
				continue includeCheck
			}
		}
		assert.True(t, zoom.validFor(l))
	}
}

func TestZoomRangeadd(t *testing.T) {
	checkZoomIncludes(t, AllZoom.add(EQ, 15), []int{15})
	checkZoomIncludes(t, AllZoom.add(GTE, 28), []int{28, 29, 30})
	checkZoomIncludes(t, AllZoom.add(GT, 28), []int{29, 30})
	checkZoomIncludes(t, AllZoom.add(LT, 4), []int{0, 1, 2, 3})
	checkZoomIncludes(t, AllZoom.add(LTE, 4), []int{0, 1, 2, 3, 4})
	checkZoomExcludes(t, AllZoom.add(NEQ, 4).add(NEQ, 8), []int{4, 8})

	assert.Equal(t, InvalidZoom.add(EQ, 15), InvalidZoom)
	assert.Equal(t, InvalidZoom.add(GTE, 15), InvalidZoom)
	assert.Equal(t, InvalidZoom.add(GT, 15), InvalidZoom)
	assert.Equal(t, InvalidZoom.add(LTE, 15), InvalidZoom)
	assert.Equal(t, InvalidZoom.add(LT, 15), InvalidZoom)
}

func TestZoomRange(t *testing.T) {
	var z ZoomRange
	z = ZoomRange(math.MaxInt32)
	z = z.add(EQ, 5)
	assert.False(t, z.validFor(4))
	assert.True(t, z.validFor(5))
	assert.False(t, z.validFor(6))
	assert.Equal(t, 1, z.Levels())
	assert.Equal(t, 5, z.First())
	assert.Equal(t, 5, z.Last())

	z = ZoomRange(math.MaxInt32)
	z = z.add(NEQ, 5)
	assert.True(t, z.validFor(4))
	assert.False(t, z.validFor(5))
	assert.True(t, z.validFor(6))
	assert.Equal(t, 30, z.Levels())
	assert.Equal(t, 0, z.First())
	assert.Equal(t, 30, z.Last())

	z = ZoomRange(math.MaxInt32)
	z = z.add(LT, 5)
	assert.True(t, z.validFor(4))
	assert.False(t, z.validFor(5))
	assert.False(t, z.validFor(6))
	assert.Equal(t, 5, z.Levels())
	assert.Equal(t, 0, z.First())
	assert.Equal(t, 4, z.Last())

	z = ZoomRange(math.MaxInt32)
	z = z.add(LTE, 5)
	assert.True(t, z.validFor(4))
	assert.True(t, z.validFor(5))
	assert.False(t, z.validFor(6))
	assert.Equal(t, 6, z.Levels())
	assert.Equal(t, 0, z.First())
	assert.Equal(t, 5, z.Last())

	z = ZoomRange(math.MaxInt32)
	z = z.add(GT, 5)
	assert.False(t, z.validFor(4))
	assert.False(t, z.validFor(5))
	assert.True(t, z.validFor(6))
	assert.Equal(t, 25, z.Levels())
	assert.Equal(t, 6, z.First())
	assert.Equal(t, 30, z.Last())

	z = ZoomRange(math.MaxInt32)
	z = z.add(GTE, 5)
	assert.False(t, z.validFor(4))
	assert.True(t, z.validFor(5))
	assert.True(t, z.validFor(6))
	assert.Equal(t, 26, z.Levels())
	assert.Equal(t, 5, z.First())
	assert.Equal(t, 30, z.Last())
}
