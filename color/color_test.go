package color

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseColor(t *testing.T) {
	var c Color
	var err error
	c, err = Parse("#f63")
	assert.NoError(t, err)
	assert.Equal(t, c, Color{15.0, 1.0, 0.6, 1.0, false})

	c, err = Parse("#ff6633")
	assert.NoError(t, err)
	assert.Equal(t, c, Color{15.0, 1.0, 0.6, 1.0, false})

	c, err = Parse("")
	assert.Error(t, err)
	c, err = Parse("ff66")
	assert.Error(t, err)
	c, err = Parse("ff6633f")
	assert.Error(t, err)

	for name, hex := range cssColors {
		c, err = Parse(name)
		assert.NoError(t, err)
		assert.Equal(t, hex, c.String())
	}
}

func TestColorString(t *testing.T) {
	assert.Equal(t, "#ff6633", Color{15.0, 1.0, 0.6, 1.0, false}.String())
	assert.Equal(t, "#fa623d", Color{11.71, 0.95, 0.61, 1.0, false}.String())
	assert.Equal(t, "rgba(0, 0, 0, 0.00000)", Color{0, 0, 0, 0, false}.String())
	assert.Equal(t, "rgba(204, 255, 153, 0.40000)", Color{90.0, 1.0, 0.8, 0.4, false}.String())
	assert.Equal(t, "#ccff99", Color{90.0, 1.0, 0.8, 1.0, false}.String())
}

func TestMustParseColor(t *testing.T) {
	var c Color
	c = MustParse("#f63")
	assert.Equal(t, c, Color{15.0, 1.0, 0.6, 1.0, false})

	c = MustParse("#ff6633")
	assert.Equal(t, c, Color{15.0, 1.0, 0.6, 1.0, false})

	assert.Panics(t, func() {
		MustParse("")
	})
	assert.Panics(t, func() {
		MustParse("#ff66")
	})
	assert.Panics(t, func() {
		MustParse("#ff6633f")
	})
	assert.Panics(t, func() {
		MustParse("#rottentomato")
	})
}

func TestFunctionsColor(t *testing.T) {
	assert.Equal(t, Color{5.0, 0.85, 0.5, 1.0, false}, Lighten(Color{5.0, 0.85, 0.45, 1.0, false}, .05))
	assert.Equal(t, Color{13.309706722717369, 0.9581840942803506, 0.5090338151746395, 1.0, true}, LightenP(Color{5.0, 0.85, 0.45, 1.0, false}, .05))
	assert.Equal(t, Color{5.0, 0.85, 0.4, 1.0, false}, Darken(Color{5.0, 0.85, 0.45, 1.0, false}, .05))
	assert.Equal(t, Color{13.309706722717369, 0.9581840942803506, 0.4090338151746395, 1.0, true}, DarkenP(Color{5.0, 0.85, 0.45, 1.0, false}, .05))

	assert.Equal(t, Color{5.0, 0.9, 0.45, 1.0, false}, Saturate(Color{5.0, 0.85, 0.45, 1.0, false}, .05))
	assert.Equal(t, Color{13.309706722717369, 1.0, 0.4590338151746395, 1.0, true}, SaturateP(Color{5.0, 0.85, 0.45, 1.0, false}, .05))
	assert.Equal(t, Color{5.0, 0.75, 0.45, 1.0, false}, Desaturate(Color{5.0, 0.8, 0.45, 1.0, false}, .05))
	assert.Equal(t, Color{13.536246577831788, 0.8851800810869034, 0.4517462301819672, 1.0, true}, DesaturateP(Color{5.0, 0.8, 0.45, 1.0, false}, .05))

	assert.Equal(t, Color{0.0, 0.0, 0.5, 0.05, false}, FadeIn(Color{0.0, 0.0, 0.5, 0.0, false}, .05))
	assert.Equal(t, Color{0.0, 0.0, 0.5, 0.5, false}, FadeIn(Color{0.0, 0.0, 0.5, 0.0, false}, .50))
	assert.Equal(t, Color{0.0, 0.0, 0.5, 0.7, false}, FadeIn(Color{0.0, 0.0, 0.5, 0.2, false}, .50))

	assert.Equal(t, Color{0.0, 0.0, 0.5, 0.45, false}, FadeOut(Color{0.0, 0.0, 0.5, 0.5, false}, .05))
	assert.Equal(t, Color{0.0, 0.0, 0.5, 0.0, false}, FadeOut(Color{0.0, 0.0, 0.5, 0.5, false}, .50))
	assert.Equal(t, Color{0.0, 0.0, 0.5, 0.4, false}, FadeOut(Color{0.0, 0.0, 0.5, 0.9, false}, .50))

	assert.Equal(t, Color{125.0, 0.85, 0.45, 1.0, false}, Spin(Color{5.0, 0.85, 0.45, 1.0, false}, 120))
	assert.Equal(t, Color{133.30970672271738, 0.9581840942803506, 0.4590338151746395, 1.0, true}, SpinP(Color{5.0, 0.85, 0.45, 1.0, false}, 120))

	assert.Equal(t, Color{2.5, 0.425, 0.225, 1.0, false}, Multiply(Color{5.0, 0.85, 0.45, 1.0, false}, 0.5))

	assert.Equal(t, Color{5.0, 0, 0.45, 1.0, false}, Greyscale(Color{5.0, 0.85, 0.45, 1.0, false}))
	assert.Equal(t, Color{13.309706722717369, 0, 0.4590338151746395, 1.0, true}, GreyscaleP(Color{5.0, 0.85, 0.45, 1.0, false}))

	assert.Equal(t, 5.0, Hue(Color{5.0, 0.85, 0.45, 1.0, false}))
	assert.Equal(t, 13.309706722717369, HueP(Color{5.0, 0.85, 0.45, 1.0, false}))
	assert.Equal(t, 0.45, Lightness(Color{5.0, 0.85, 0.45, 1.0, false}))
	assert.Equal(t, 0.4590338151746395, LightnessP(Color{5.0, 0.85, 0.45, 1.0, false}))
	assert.Equal(t, 0.85, Saturation(Color{5.0, 0.85, 0.45, 1.0, false}))
	assert.Equal(t, 0.9581840942803506, SaturationP(Color{5.0, 0.85, 0.45, 1.0, false}))
	assert.Equal(t, 0.6, Alpha(Color{5.0, 0.85, 0.45, 0.6, false}))
}

func TestSetHue(t *testing.T) {
	assert.Equal(t, "#737373", SetHue(MustParse("#737373"), MustParse("red")).String()) //still grey

	// SetHue uses HuSL color space to keep saturation and lightning
	assert.Equal(t, "#c64545", SetHue(MustParse("#996644"), MustParse("red")).String())
	// should differ from calculations with HSL
	hsl := MustParse("#996644")
	hsl.H = MustParse("red").H
	assert.Equal(t, hsl.String(), "#994444")
}
