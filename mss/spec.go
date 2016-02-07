package mss

import "github.com/omniscale/magnacarto/color"

var attributeTypes map[string]isValid

type isValid func(interface{}) bool

func isNumber(val interface{}) bool {
	_, ok := val.(float64)
	return ok
}

func isNumbers(val interface{}) bool {
	vals, ok := val.([]Value)
	if !ok {
		return false
	}
	for _, v := range vals {
		if !isNumber(v) {
			return false
		}
	}
	return true
}

func isString(val interface{}) bool {
	_, ok := val.(string)
	return ok
}

func isStrings(val interface{}) bool {
	vals, ok := val.([]Value)
	if !ok {
		return false
	}
	for _, v := range vals {
		if !isString(v) {
			return false
		}
	}
	return true
}

func isStringOrStrings(val interface{}) bool {
	return isString(val) || isStrings(val)
}

func isFieldOr(other isValid) isValid {
	return func(val interface{}) bool {
		if s, ok := val.(string); ok {
			if len(s) > 2 && s[0] == '[' && s[len(s)-1] == ']' {
				return true
			}
			return false
		}
		return other(val)
	}
}

func isColor(val interface{}) bool {
	_, ok := val.(color.Color)
	return ok
}

func isBool(val interface{}) bool {
	_, ok := val.(bool)
	return ok
}

func isKeyword(keywords ...string) isValid {
	return func(val interface{}) bool {
		k, ok := val.(string)
		if !ok {
			return false
		}
		for _, expected := range keywords {
			if k == expected {
				return true
			}
		}
		return false
	}
}

func isStops(val interface{}) bool {
	vals, ok := val.([]Value)
	if !ok {
		return false
	}
	for _, v := range vals {
		if _, ok := v.(Stop); !ok {
			return false
		}
	}
	return true
}

func isCompOp(val interface{}) bool {
	return isKeyword(
		"clear",
		"src",
		"dst",
		"src-over",
		"dst-over",
		"src-in",
		"dst-in",
		"src-out",
		"dst-out",
		"src-atop",
		"dst-atop",
		"xor",
		"plus",
		"minus",
		"multiply",
		"divide",
		"screen",
		"overlay",
		"darken",
		"lighten",
		"color-dodge",
		"color-burn",
		"hard-light",
		"soft-light",
		"difference",
		"exclusion",
		"contrast",
		"invert",
		"invert-rgb",
		"grain-merge",
		"grain-extract",
		"hue",
		"saturation",
		"color",
		"value",
	)(val)
}

func isScaling(val interface{}) bool {
	return isKeyword(
		"near",
		"fast",
		"bilinear",
		"bicubic",
		"spline16",
		"spline36",
		"hanning",
		"hamming",
		"hermite",
		"kaiser",
		"quadric",
		"catrom",
		"gaussian",
		"bessel",
		"mitchell",
		"sinc",
		"lanczos",
		"blackman",
	)(val)
}
func init() {
	attributeTypes = map[string]isValid{
		"background-color": isColor,

		"building-fill":   isColor,
		"building-fill-opacity": isNumber,
		"building-height": isNumber,

		"line-cap":          isKeyword("round", "butt", "square"),
		"line-clip":         isBool,
		"line-color":        isColor,
		"line-dasharray":    isNumbers,
		"line-gamma-method": isKeyword("power", "linear", "none", "threshold", "multiply"),
		"line-join":         isKeyword("miter", "round", "bevel"),
		"line-offset":       isNumber,
		"line-opacity":      isNumber,
		"line-rastersizer":  isKeyword("full", "fast"),
		"line-simplify":     isNumber,
		"line-smooth":       isNumber,
		"line-width":        isNumber,

		"marker-allow-overlap": isBool,
		"marker-file":          isString,
		"marker-fill":          isColor,
		"marker-height":        isNumber,
		"marker-line-color":    isColor,
		"marker-line-width":    isNumber,
		"marker-opacity":       isNumber,
		"marker-placement":     isKeyword("point", "interior", "line"),
		"marker-spacing":       isNumber,
		"marker-transform":     isString,
		"marker-type":          isKeyword("arrow", "ellipse"),
		"marker-width":         isNumber,

		"point-file":             isString,
		"point-allow-overlap":    isBool,
		"point-opacity":          isNumber,
		"point-transform":        isString,
		"point-ignore-placement": isBool,

		"polygon-fill":              isColor,
		"polygon-gamma":             isNumber,
		"polygon-gamma-method":      isKeyword("power", "linear", "none", "threshold", "multiply"),
		"polygon-opacity":           isNumber,
		"polygon-pattern-alignment": isKeyword("global", "local"),
		"polygon-pattern-file":      isString,

		"shield-allow-overlap":     isBool,
		"shield-avoid-edges":       isBool,
		"shield-character-spacing": isNumber,
		"shield-clip":              isBool,
		"shield-dx":                isNumber,
		"shield-dy":                isNumber,
		"shield-face-name":         isStringOrStrings,
		"shield-file":              isString,
		"shield-fill":              isColor,
		"shield-halo-fill":         isColor,
		"shield-halo-radius":       isNumber,
		"shield-line-spacing":      isNumber,
		"shield-min-distance":      isNumber,
		"shield-min-padding":       isNumber,
		"shield-name":              isString,
		"shield-opacity":           isNumber,
		"shield-placement":         isKeyword("line", "point", "vertex", "interior"),
		"shield-placement-type":    isKeyword("dummy", "simple", "list"),
		"shield-placements":        isString,
		"shield-size":              isNumber,
		"shield-spacing":           isNumber,
		"shield-text-dx":           isNumber,
		"shield-text-dy":           isNumber,
		"shield-transform":         isKeyword("none", "uppercase", "lowercase", "capitalize"),
		"shield-wrap-before":       isBool,
		"shield-wrap-character":    isString,
		"shield-wrap-width":        isNumber,
		"shield-unlock-image":      isBool,

		"text-allow-overlap":     isBool,
		"text-avoid-edges":       isBool,
		"text-character-spacing": isNumber,
		"text-clip":              isBool,
		"text-dx":                isNumber,
		"text-dy":                isNumber,
		"text-face-name":         isStringOrStrings,
		"text-fill":              isColor,
		"text-halo-fill":         isColor,
		"text-halo-radius":       isNumber,
		"text-line-spacing":      isNumber,
		"text-min-distance":      isNumber,
		"text-min-padding":       isNumber,
		"text-name":              isString,
		"text-opacity":           isNumber,
		"text-orientation":       isFieldOr(isNumber),
		"text-placement":         isKeyword("line", "point", "vertex", "interior"),
		"text-placement-type":    isKeyword("dummy", "simple", "list"),
		"text-placements":        isString,
		"text-size":              isNumber,
		"text-spacing":           isNumber,
		"text-transform":         isKeyword("none", "uppercase", "lowercase", "capitalize"),
		"text-wrap-before":       isBool,
		"text-wrap-character":    isString,
		"text-wrap-width":        isNumber,

		"raster-opacity":                 isNumber,
		"raster-scaling":                 isScaling,
		"raster-colorizer-default-mode":  isKeyword("discrete", "linear", "exact"),
		"raster-colorizer-default-color": isColor,
		"raster-colorizer-stops":         isStops,
		"raster-comp-op":                 isCompOp,
		"raster-filter-factor":           isNumber,
		"raster-mesh-size":               isNumber,
		"raster-epsilon":                 isNumber,
	}
}

func validProperty(property string, value interface{}) bool {
	checkFunc, ok := attributeTypes[property]
	if !ok {
		return false
	}
	return checkFunc(value)
}
