package color

import "math"

func Lighten(c RGBA, v float64) RGBA {
	hsl := c.HSL()
	hsl.L += v
	hsl.L = clamp(hsl.L)
	return hsl.RGB()
}

func Darken(c RGBA, v float64) RGBA {
	hsl := c.HSL()
	hsl.L -= v
	hsl.L = clamp(hsl.L)
	return hsl.RGB()
}

func Saturate(c RGBA, v float64) RGBA {
	hsl := c.HSL()
	hsl.S += v
	hsl.S = clamp(hsl.S)
	return hsl.RGB()
}

func Desaturate(c RGBA, v float64) RGBA {
	hsl := c.HSL()
	hsl.S -= v
	hsl.S = clamp(hsl.S)
	return hsl.RGB()
}

func FadeIn(c RGBA, v float64) RGBA {
	hsl := c.HSL()
	hsl.A += v
	hsl.A = clamp(hsl.A)
	return hsl.RGB()
}

func FadeOut(c RGBA, v float64) RGBA {
	hsl := c.HSL()
	hsl.A -= v
	hsl.A = clamp(hsl.A)
	return hsl.RGB()
}

// Change hue of color by v centi-degrees (0-3.6)
func Spin(c RGBA, v float64) RGBA {
	hsl := c.HSL()
	v *= 100
	hsl.H += v
	if hsl.H < 0 {
		hsl.H += 360
	} else if hsl.H > 360 {
		hsl.H -= 360
	}
	return hsl.RGB()
}

func Multiply(c RGBA, v float64) RGBA {
	c.R = clamp(c.R * v)
	c.G = clamp(c.G * v)
	c.B = clamp(c.B * v)
	return c
}

func Mix(c1, c2 RGBA, weight float64) RGBA {
	w := weight*2 - 1
	a := c1.A - c2.A

	var w1 float64

	if w*a == -1 {
		w1 = (w + 1) / 2.0
	} else {
		w1 = ((w+a)/(1+w*a) + 1) / 2.0
	}
	w2 := 1 - w1

	return RGBA{
		R: c1.R*w1 + c2.R*w2,
		G: c1.G*w1 + c2.G*w2,
		B: c1.B*w1 + c2.B*w2,
		A: c1.A*weight + c2.A*(1-weight),
	}
}

func SetHue(c, hue RGBA) RGBA {
	base := c.HuSL()
	base.H = hue.HuSL().H
	return base.RGB()
}

func clamp(v float64) float64 {
	return math.Max(math.Min(v, 1.0), 0.0)
}
