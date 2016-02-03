package color

import (
	"math"
)

/* from https://github.com/gka/chroma.js written by Gregor Aisch
under BSD license */

func HslToRgb(h, s, l float64) (r, g, b float64) {
	if s == 0 {
		r = l
		g = l
		b = l
	} else {
		t3 := [3]float64{0, 0, 0}
		c := [3]float64{0, 0, 0}
		var t2 float64
		if l < 0.5 {
			t2 = l * (1.0 + s)
		} else {
			t2 = l + s - l*s
		}
		t1 := 2.0*l - t2
		h /= 360.0
		t3[0] = h + 1.0/3.0
		t3[1] = h
		t3[2] = h - 1.0/3.0
		for i := 0; i < 3; i++ {
			if t3[i] < 0 {
				t3[i] += 1.0
			}
			if t3[i] > 1 {
				t3[i] -= 1.0
			}
			if 6.0*t3[i] < 1 {
				c[i] = t1 + (t2-t1)*6.0*t3[i]
			} else if 2.0*t3[i] < 1 {
				c[i] = t2
			} else if 3.0*t3[i] < 2 {
				c[i] = t1 + (t2-t1)*((2.0/3.0)-t3[i])*6.0
			} else {
				c[i] = t1
			}
		}
		r = c[0]
		g = c[1]
		b = c[2]
	}
	return
}

func RgbToHsl(r, g, b float64) (h, s, l float64) {
	min := math.Min(r, math.Min(g, b))
	max := math.Max(r, math.Max(g, b))
	l = (max + min) / 2.0
	if max == min {
		s = 0
		h = 0
	} else {
		if l < 0.5 {
			s = (max - min) / (max + min)
		} else {
			s = (max - min) / (2.0 - max - min)
		}
		if r == max {
			h = (g - b) / (max - min)
		} else if g == max {
			h = 2.0 + (b-r)/(max-min)
		} else if b == max {
			h = 4.0 + (r-g)/(max-min)
		}
		h *= 60.0
		if h < 0 {
			h += 360.0
		}
	}
	return
}
