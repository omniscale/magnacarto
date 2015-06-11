// Package color implements color functions.
package color

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
)

type RGBA struct {
	R, G, B, A float64
}

var hexRe = regexp.MustCompile(`[a-fA-F0-9]{6,6}`)

func Parse(color string) (RGBA, error) {
	rgba := RGBA{}
	if len(color) == 0 {
		return rgba, errors.New("empty color")
	}
	if color[0] == '#' {
		return parseHex(color)
	}

	if color == "transparent" {
		return RGBA{0, 0, 0, 0}, nil
	}
	hex, ok := cssColors[color]
	if ok {
		return parseHex(hex)
	}
	return rgba, errors.New("unknown color")
}

func parseHex(hex string) (RGBA, error) {
	rgba := RGBA{}
	rgba.A = 1
	if len(hex) > 0 && hex[0] == '#' {
		hex = hex[1:]
	}
	if len(hex) == 3 {
		hex = string(hex[0]) + string(hex[0]) + string(hex[1]) + string(hex[1]) + string(hex[2]) + string(hex[2])
	}

	if !hexRe.MatchString(hex) {
		return rgba, errors.New("invalid hex color")
	}

	if len(hex) == 6 {
		v, err := strconv.ParseInt(hex[0:2], 16, 32)
		if err != nil {
			return rgba, err
		}
		rgba.R = float64(v) / 255
		v, err = strconv.ParseInt(hex[2:4], 16, 32)
		if err != nil {
			return rgba, err
		}
		rgba.G = float64(v) / 255
		v, err = strconv.ParseInt(hex[4:6], 16, 32)
		if err != nil {
			return rgba, err
		}
		rgba.B = float64(v) / 255
	} else {
		return rgba, errors.New("hex color not 3 or 6 chars long")
	}
	return rgba, nil
}

func MustParse(hex string) RGBA {
	rgba, err := Parse(hex)
	if err != nil {
		panic(err)
	}
	return rgba
}

func (rgba RGBA) String() string {
	if rgba.A == 1.0 {
		return rgba.Hex()
	}
	return fmt.Sprintf("rgba(%d, %d, %d, %.5f)", int(rgba.R*255), int(rgba.G*255), int(rgba.B*255), rgba.A)
}

func (rgba RGBA) Hex() string {
	return fmt.Sprintf("#%02x%02x%02x", int(rgba.R*255), int(rgba.G*255), int(rgba.B*255))
}

func (rgba RGBA) HuSL() HuSLA {
	husl := HuSLA{A: rgba.A}
	husl.H, husl.S, husl.L = rgb2husl(rgba.R, rgba.G, rgba.B)
	return husl
}

func (rgba RGBA) HSL() HSLA {
	r := rgba.R
	g := rgba.G
	b := rgba.B
	a := rgba.A
	max := math.Max(math.Max(r, g), b)
	min := math.Min(math.Min(r, g), b)
	h := (max + min) / 2.0
	s := h
	l := h

	if max == min {
		h = 0
		s = 0
	} else {
		d := max - min
		if l > 0.5 {
			s = d / (2 - max - min)
		} else {
			s = d / (max + min)
		}

		switch max {
		case r:
			h = (g - b) / d
			if g < b {
				h += 6
			}
		case g:
			h = (b-r)/d + 2
		case b:
			h = (r-g)/d + 4
		}
		h /= 6
	}
	return HSLA{H: h * 360, S: s, L: l, A: a}
}

type HuSLA struct {
	H, S, L, A float64
}

func (husl HuSLA) RGB() RGBA {
	rgba := RGBA{A: husl.A}
	rgba.R, rgba.G, rgba.B = husl2rgb(husl.H, husl.S, husl.L)
	return rgba
}

type HSLA struct {
	H, S, L, A float64
}

func (hsl HSLA) RGB() RGBA {
	h := float64(int(hsl.H)%360) / 360.0
	s := hsl.S
	l := hsl.L
	a := hsl.A

	m2 := l + s - l*s
	if l <= 0.5 {
		m2 = l * (s + 1.0)
	}
	m1 := l*2.0 - m2

	hue := func(h float64) float64 {
		if h < 0.0 {
			h = h + 1.0
		} else if h > 1.0 {
			h = h - 1.0
		}

		if h*6.0 < 1.0 {
			return m1 + (m2-m1)*h*6.0
		} else if h*2.0 < 1.0 {
			return m2
		} else if h*3.0 < 2.0 {
			return m1 + (m2-m1)*(2.0/3.0-h)*6.0
		} else {
			return m1
		}
	}

	return RGBA{
		hue(h + 1.0/3.0),
		hue(h),
		hue(h - 1.0/3.0),
		a,
	}
}
