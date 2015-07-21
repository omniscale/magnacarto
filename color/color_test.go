package color

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseColor(t *testing.T) {
	var c RGBA
	var err error
	c, err = Parse("#f63")
	assert.NoError(t, err)
	assert.Equal(t, c, RGBA{1.0, 0.4, 0.2, 1.0})

	c, err = Parse("#ff6633")
	assert.NoError(t, err)
	assert.Equal(t, c, RGBA{1.0, 0.4, 0.2, 1.0})

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
	assert.Equal(t, "#ff6633", MustParse("#f63").String())
	assert.Equal(t, "#fa623d", MustParse("#fA623D").String())
	assert.Equal(t, "rgba(0, 0, 0, 0.00000)", RGBA{0, 0, 0, 0}.String())
	assert.Equal(t, "rgba(204, 255, 153, 0.40000)", RGBA{0.8, 1.0, 0.6, 0.4}.String())
	assert.Equal(t, "#ccff99", RGBA{0.8, 1.0, 0.6, 1}.String())
}

func TestMustParseColor(t *testing.T) {
	var c RGBA
	c = MustParse("#f63")
	assert.Equal(t, c, RGBA{1.0, 0.4, 0.2, 1.0})

	c = MustParse("#ff6633")
	assert.Equal(t, c, RGBA{1.0, 0.4, 0.2, 1.0})

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

func TestHSLColor(t *testing.T) {
	var c HSLA
	c = RGBA{1.0, 1.0, 1.0, 0.0}.HSL()
	assert.Equal(t, c, HSLA{0.0, 0.0, 1.0, 0.0})
	c = RGBA{0.0, 0.0, 0.0, 0.0}.HSL()
	assert.Equal(t, c, HSLA{0.0, 0.0, 0.0, 0.0})
	c = RGBA{0.5, 0.5, 0.5, 0.0}.HSL()
	assert.Equal(t, c, HSLA{0.0, 0.0, 0.5, 0.0})

	c = RGBA{1.0, 0.0, 0.0, 0.0}.HSL()
	assert.Equal(t, c, HSLA{0.0, 1.0, 0.5, 0.0})
	c = RGBA{0.0, 1.0, 0.0, 0.0}.HSL()
	assert.Equal(t, c, HSLA{120.0, 1.0, 0.5, 0.0})
	c = RGBA{0.0, 0.0, 1.0, 0.0}.HSL()
	assert.Equal(t, c, HSLA{240.0, 1.0, 0.5, 0.0})
}

func TestRGBColor(t *testing.T) {
	var c RGBA
	c = HSLA{0, 1.0, 0.5, 0.0}.RGB()
	assert.Equal(t, c, RGBA{1.0, 0.0, 0.0, 0.0})
	c = HSLA{120, 1.0, 0.5, 0.0}.RGB()
	assert.Equal(t, c, RGBA{0.0, 1.0, 0.0, 0.0})
	c = HSLA{240, 1.0, 0.5, 0.0}.RGB()
	assert.Equal(t, c, RGBA{0.0, 0.0, 1.0, 0.0})
}

func TestRGBHSLColor(t *testing.T) {
	assert.Equal(t, RGBA{1.0, 1.0, 1.0, 0.0}.HSL().RGB(), RGBA{1.0, 1.0, 1.0, 0.0})
	assert.Equal(t, RGBA{0.0, 0.0, 0.0, 0.0}.HSL().RGB(), RGBA{0.0, 0.0, 0.0, 0.0})
	assert.Equal(t, RGBA{1.0, 0.0, 0.0, 0.0}.HSL().RGB(), RGBA{1.0, 0.0, 0.0, 0.0})
	assert.Equal(t, RGBA{0.0, 1.0, 0.0, 0.0}.HSL().RGB(), RGBA{0.0, 1.0, 0.0, 0.0})
	assert.Equal(t, RGBA{0.0, 0.0, 1.0, 0.0}.HSL().RGB(), RGBA{0.0, 0.0, 1.0, 0.0})
}

func TestFunctionsColor(t *testing.T) {
	assert.Equal(t, "#ed2b1a", Lighten(MustParse("#dd2211"), .05).Hex())
	assert.Equal(t, "#dc1e11", Darken(MustParse("#ed2b1a"), .05).Hex())

	assert.Equal(t, "#f32213", Saturate(MustParse("#ed2b1a"), .05).Hex())
	assert.Equal(t, "#ec2719", Desaturate(MustParse("#f32213"), .05).Hex())

	assert.Equal(t, RGBA{0.5, 0.5, 0.5, 0.05}, FadeIn(RGBA{0.5, 0.5, 0.5, 0.0}, .05))
	assert.Equal(t, RGBA{0.5, 0.5, 0.5, 0.5}, FadeIn(RGBA{0.5, 0.5, 0.5, 0.0}, .50))
	assert.Equal(t, RGBA{0.5, 0.5, 0.5, 0.7}, FadeIn(RGBA{0.5, 0.5, 0.5, 0.2}, .50))

	assert.Equal(t, RGBA{0.5, 0.5, 0.5, 0.45}, FadeOut(RGBA{0.5, 0.5, 0.5, 0.5}, .05))
	assert.Equal(t, RGBA{0.5, 0.5, 0.5, 0.0}, FadeOut(RGBA{0.5, 0.5, 0.5, 0.5}, .50))
	assert.Equal(t, RGBA{0.5, 0.5, 0.5, 0.4}, FadeOut(RGBA{0.5, 0.5, 0.5, 0.9}, .50))

	assert.Equal(t, "#00ff00", Spin(MustParse("#ff0000"), 120).Hex())
	assert.Equal(t, "#0000ff", Spin(MustParse("#00ff00"), 120).Hex())
	assert.Equal(t, "#ff0000", Spin(MustParse("#0000ff"), 120).Hex())

	assert.Equal(t, "#7f6633", Multiply(MustParse("#ffcc66"), 0.5).Hex())
	assert.Equal(t, "#fecc66", Multiply(MustParse("#7f6633"), 2).Hex())
}

func TestSetHue(t *testing.T) {
	assert.Equal(t, SetHue(MustParse("#737373"), MustParse("red")).Hex(), "#727372") //still grey

	// SetHue uses HuSL color space to keep saturation and lightning
	assert.Equal(t, SetHue(MustParse("#996644"), MustParse("red")).Hex(), "#c64444")
	// should differ from calculations with HSL
	hsl := MustParse("#996644").HSL()
	hsl.H = MustParse("red").HSL().H
	assert.Equal(t, hsl.RGB().Hex(), "#994444")
}
