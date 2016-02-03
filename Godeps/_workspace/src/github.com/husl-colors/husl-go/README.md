[![Build Status](https://travis-ci.org/husl-colors/husl-go.svg?branch=master)](https://travis-ci.org/husl-colors/husl-go)

Go port of HUSL (revision 4), written by [Michael Glanznig](https://github.com/nebulon42)

More details about HUSL at http://www.husl-colors.org.

# API

**husl.HuslToHex(hue, saturation, lightness)**

*hue* is a number between 0 and 360, *saturation* and *lightness* are numbers between 0 and 100. This function returns the resulting color as a hex string.

**husl.HuslToRGB(hue, saturation, lightness)**

Like above, but returns 3 numbers between 0 and 1, for the r, g, and b channel.

**husl.HuslFromHex(hex)**

Takes a hex string and returns the HUSL color as 3 numbers for hue (0-360), saturation (0-100) and lightness (0-100).
_Note_: The result can have rounding errors. For example saturation can be 100.00000000000007

**husl.HuslFromRGB(red, green, blue)**

Like above, but *red*, *green* and *blue* are passed as numbers between 0 and 1.

Use **HuslpToHex**, **HuslpToRGB**, **HuslpFromHex** and **HuslpFromRGB** for the pastel variant (HUSLp).

# Testing

Run `go test`.

# Thanks

Testing was inspired by [omniscale/magnacarto](https://github.com/omniscale/magnacarto).
