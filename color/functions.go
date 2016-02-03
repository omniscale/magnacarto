package color

import "math"

func Lighten(c Color, v float64) Color {
	c.L += v
	c.L = clamp(c.L)
	return c
}

func LightenP(c Color, v float64) Color {
	if !c.Perceptual {
		c = c.ToPerceptual()
	}
	return Lighten(c, v)
}

func Darken(c Color, v float64) Color {
	c.L -= v
	c.L = clamp(c.L)
	return c
}

func DarkenP(c Color, v float64) Color {
	if !c.Perceptual {
		c = c.ToPerceptual()
	}
	return Darken(c, v)
}

func Saturate(c Color, v float64) Color {
	c.S += v
	c.S = clamp(c.S)
	return c
}

func SaturateP(c Color, v float64) Color {
	if !c.Perceptual {
		c = c.ToPerceptual()
	}
	return Saturate(c, v)
}

func Desaturate(c Color, v float64) Color {
	c.S -= v
	c.S = clamp(c.S)
	return c
}

func DesaturateP(c Color, v float64) Color {
	if !c.Perceptual {
		c = c.ToPerceptual()
	}
	return Desaturate(c, v)
}

func FadeIn(c Color, v float64) Color {
	c.A += v
	c.A = clamp(c.A)
	return c
}

func FadeOut(c Color, v float64) Color {
	c.A -= v
	c.A = clamp(c.A)
	return c
}

func Spin(c Color, v float64) Color {
	c.H += v
	if c.H < 0 {
		c.H += 360
	} else if c.H > 360 {
		c.H -= 360
	}
	return c
}

func SpinP(c Color, v float64) Color {
	if !c.Perceptual {
		c = c.ToPerceptual()
	}
	return Spin(c, v)
}

func Greyscale(c Color) Color {
	return Desaturate(c, 1.0)
}

func GreyscaleP(c Color) Color {
	return DesaturateP(c, 1.0)
}

func Hue(c Color) float64 {
	return c.H
}

func HueP(c Color) float64 {
	if !c.Perceptual {
		c = c.ToPerceptual()
	}
	return Hue(c)
}

func Saturation(c Color) float64 {
	return c.S
}

func SaturationP(c Color) float64 {
	if !c.Perceptual {
		c = c.ToPerceptual()
	}
	return Saturation(c)
}

func Lightness(c Color) float64 {
	return c.L
}

func LightnessP(c Color) float64 {
	if !c.Perceptual {
		c = c.ToPerceptual()
	}
	return Lightness(c)
}

func Alpha(c Color) float64 {
	return c.A
}

func Multiply(c Color, v float64) Color {
	c.H = math.Max(math.Min(c.H*v, 360.0), 0.0)
	c.S = clamp(c.S * v)
	c.L = clamp(c.L * v)
	return c
}

func Mix(c1, c2 Color, weight float64) Color {
	w := weight*2 - 1
	a := c1.A - c2.A
	perceptual := c1.Perceptual || c2.Perceptual

	if c1.Perceptual && !c2.Perceptual {
		c2 = c2.ToPerceptual()
	} else if !c1.Perceptual && c2.Perceptual {
		c1 = c1.ToPerceptual()
	}

	var w1 float64

	if w*a == -1 {
		w1 = (w + 1) / 2.0
	} else {
		w1 = ((w+a)/(1+w*a) + 1) / 2.0
	}
	w2 := 1 - w1

	return Color{
		H:          c1.H*w1 + c2.H*w2,
		S:          c1.S*w1 + c2.S*w2,
		L:          c1.L*w1 + c2.L*w2,
		A:          c1.A*weight + c2.A*(1-weight),
		Perceptual: perceptual,
	}
}

func SetHue(c, hue Color) Color {
	base := c.ToPerceptual()
	base.H = hue.ToPerceptual().H
	return base
}

func clamp(v float64) float64 {
	return math.Max(math.Min(v, 1.0), 0.0)
}
