// Package color implements color functions.
package color

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"

	"github.com/husl-colors/husl-go"
)

type Color struct {
	H, S, L, A float64
	Perceptual bool
}

var hexRe = regexp.MustCompile(`[a-fA-F0-9]{6,6}`)

func Parse(colorStr string) (Color, error) {
	color := Color{}
	color.A = 1
	if len(colorStr) == 0 {
		return color, errors.New("empty color")
	}
	if colorStr[0] == '#' {
		return parseHex(colorStr)
	}

	if colorStr == "transparent" {
		return Color{0, 0, 0, 0, false}, nil
	}
	hex, ok := cssColors[colorStr]
	if ok {
		return parseHex(hex)
	}
	return color, errors.New("unknown color")
}

func parseHex(hex string) (Color, error) {
	color := Color{}
	color.A = 1
	if len(hex) > 0 && hex[0] == '#' {
		hex = hex[1:]
	}
	if len(hex) == 3 {
		hex = string(hex[0]) + string(hex[0]) + string(hex[1]) + string(hex[1]) + string(hex[2]) + string(hex[2])
	}

	if !hexRe.MatchString(hex) {
		return color, errors.New("invalid hex color")
	}

	if len(hex) == 6 {
		v, err := strconv.ParseInt(hex[0:2], 16, 32)
		if err != nil {
			return color, err
		}
		r := float64(v) / 255.0
		v, err = strconv.ParseInt(hex[2:4], 16, 32)
		if err != nil {
			return color, err
		}
		g := float64(v) / 255.0
		v, err = strconv.ParseInt(hex[4:6], 16, 32)
		if err != nil {
			return color, err
		}
		b := float64(v) / 255.0

		color.H, color.S, color.L = RgbToHsl(r, g, b)
	} else {
		return color, errors.New("hex color not 3 or 6 chars long")
	}
	return color, nil
}

func MustParse(hex string) Color {
	color, err := Parse(hex)
	if err != nil {
		panic(err)
	}
	return color
}

func (color Color) ToPerceptual() Color {
	if color.Perceptual {
		return color
	} else {
		// transition via RGB, because HSL values cannot be directly
		// transformed into HUSL values easily
		r, g, b := HslToRgb(color.H, color.S, color.L)
		color.H, color.S, color.L = husl.HuslFromRGB(r, g, b)
		color.Perceptual = true
		return color
	}
}

func (color Color) ToStandard() Color {
	if !color.Perceptual {
		return color
	} else {
		// transition via RGB, because HUSL values cannot be directly
		// transformed into HSL values easily
		r, g, b := husl.HuslToRGB(color.H, color.S, color.L)
		color.H, color.S, color.L = RgbToHsl(r, g, b)
		color.Perceptual = false
		return color
	}
}

func (color Color) String() string {
	var r, g, b float64
	if color.Perceptual {
		r, g, b = husl.HuslToRGB(color.H, color.S, color.L)
	} else {
		r, g, b = HslToRgb(color.H, color.S, color.L)
	}

	if color.A == 1.0 {
		return fmt.Sprintf("#%02x%02x%02x", round(r*255), round(g*255), round(b*255))
	} else {
		return fmt.Sprintf("rgba(%d, %d, %d, %.5f)", round(r*255), round(g*255), round(b*255), color.A)
	}
}

func (color Color) HexString() string {
	var r, g, b float64
	if color.Perceptual {
		r, g, b = husl.HuslToRGB(color.H, color.S, color.L)
	} else {
		r, g, b = HslToRgb(color.H, color.S, color.L)
	}

	if color.A == 1.0 {
		return fmt.Sprintf("#%02x%02x%02x", round(r*255), round(g*255), round(b*255))
	} else {
		return fmt.Sprintf("#%02x%02x%02x%02x", round(r*255), round(g*255), round(b*255), round(color.A*255))
	}
}

func round(f float64) int {
	if math.Abs(f) < 0.5 {
		return 0
	}
	return int(f + math.Copysign(0.5, f))
}
