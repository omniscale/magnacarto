package mss

import (
	"regexp"
	"strings"

	"github.com/omniscale/magnacarto/color"
)

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

var expressionHasField = regexp.MustCompile(`^(?:[^\[\]]|\[[^\[\]]*\])*$`)
var expressionHasMath = regexp.MustCompile(`\d|[-*/+]`)

// isExpressionString checks whether val is a string that is likely an data
// expression with field ([name]) or math (2*2).
// Not a proper parser, just minimal checks.
func isExpressionString(val interface{}) bool {
	s, ok := val.(string)
	if !ok {
		return false
	}
	if strings.Contains(s, "[") {
		// checks whether all fields are closed ([field])
		return expressionHasField.MatchString(s)
	}
	return expressionHasMath.MatchString(s)
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

func isStringOrField(val interface{}) bool {
	switch vals := val.(type) {
	case string, Field, []FormatParameter, FormatEnd:
		return true
	case []Value:
		for _, v := range vals {
			if !isStringOrField(v) {
				return false
			}
		}
		return true
	default:
		return false
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

func isKeywordOr(other isValid, keywords ...string) isValid {
	return func(val interface{}) bool {
		if other(val) {
			return true
		}
		return isKeyword(keywords...)(val)
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

func isSimplifyAlgorithm(val interface{}) bool {
	return isKeyword(
		"radial-distance",
		"zhao-saalfeld",
		"visvalingam-whyatt",
	)(val)
}

func isRasterizer(val interface{}) bool {
	return isKeyword(
		"full",
		"fast",
	)(val)
}

func isVerticalAlignment(val interface{}) bool {
	return isKeyword(
		"top",
		"middle",
		"bottom",
		"auto",
	)(val)
}

func isJustifyAlignment(val interface{}) bool {
	return isKeyword(
		"left",
		"center",
		"right",
		"auto",
	)(val)
}

func isImageFilters(val interface{}) bool {
	if _, ok := val.(Function); ok {
		return isImageFilter(val)
	}
	funcs, ok := val.([]Value)
	if !ok {
		return false
	}
	for _, f := range funcs {
		if !isImageFilter(f) {
			return false
		}
	}
	return true
}

func isImageFilter(val interface{}) bool {
	f, ok := val.(Function)
	if !ok {
		return false
	}
	numParams, ok := imageFilterFuncs[f.Name]
	if !ok {
		return false
	}
	if numParams == -1 {
		// unlimited params
		return true
	}
	return len(f.Params) == numParams
}

var imageFilterFuncs map[string]int

func init() {
	imageFilterFuncs = map[string]int{
		"agg-stack-blur":          2,
		"emboss":                  0,
		"blur":                    0,
		"gray":                    0,
		"sobel":                   0,
		"edge-detect":             0,
		"x-gradient":              0,
		"y-gradient":              0,
		"invert":                  0,
		"sharpen":                 0,
		"color-blind-protanope":   0,
		"color-blind-deuteranope": 0,
		"color-blind-tritanope":   0,
		"colorize-alpha":          -1,
		"color-to-alpha":          1,
		"scale-hsla":              8,
	}
}

func init() {
	attributeTypes = map[string]isValid{
		"background-color": isColor,

		// layer attributes
		"opacity":              isNumber,
		"comp-op":              isCompOp,
		"image-filters":        isImageFilters,
		"direct-image-filters": isImageFilters,

		"building-fill":         isColor,
		"building-fill-opacity": isNumber,
		"building-height":       isNumber,

		"dot-fill":    isColor,
		"dot-opacity": isNumber,
		"dot-width":   isNumber,
		"dot-height":  isNumber,
		"dot-comp-op": isCompOp,

		"line-cap":                isKeyword("round", "butt", "square"),
		"line-clip":               isBool,
		"line-color":              isColor,
		"line-dasharray":          isNumbers,
		"line-dash-offset":        isNumbers,
		"line-gamma":              isNumber,
		"line-gamma-method":       isKeyword("power", "linear", "none", "threshold", "multiply"),
		"line-join":               isKeyword("miter", "miter-revert", "round", "bevel"),
		"line-miterlimit":         isNumber,
		"line-offset":             isNumber,
		"line-opacity":            isNumber,
		"line-rasterizer":         isRasterizer,
		"line-simplify":           isNumber,
		"line-simplify-algorithm": isSimplifyAlgorithm,
		"line-smooth":             isNumber,
		"line-width":              isNumber,
		"line-comp-op":            isCompOp,
		"line-geometry-transform": isString,

		"line-pattern-file":               isString,
		"line-pattern-clip":               isBool,
		"line-pattern-opacity":            isNumber,
		"line-pattern-simplify":           isNumber,
		"line-pattern-simplify-algorithm": isSimplifyAlgorithm,
		"line-pattern-smooth":             isNumber,
		"line-pattern-offset":             isNumber,
		"line-pattern-geometry-transform": isString,
		"line-pattern-comp-op":            isCompOp,

		"marker-allow-overlap":      isBool,
		"marker-file":               isString,
		"marker-fill":               isColor,
		"marker-fill-opacity":       isNumber,
		"marker-height":             isNumber,
		"marker-line-color":         isColor,
		"marker-line-width":         isNumber,
		"marker-line-opacity":       isNumber,
		"marker-opacity":            isNumber,
		"marker-placement":          isKeyword("point", "interior", "line", "vertex-first", "vertex-last"),
		"marker-spacing":            isNumber,
		"marker-transform":          isString,
		"marker-type":               isKeyword("arrow", "ellipse"),
		"marker-width":              isNumber,
		"marker-multi-policy":       isKeyword("each", "whole", "largest"),
		"marker-avoid-edges":        isBool,
		"marker-ignore-placement":   isBool,
		"marker-max-error":          isNumber,
		"marker-clip":               isBool,
		"marker-simplify":           isNumber,
		"marker-simplify-algorithm": isSimplifyAlgorithm,
		"marker-smooth":             isNumber,
		"marker-geometry-transform": isString,
		"marker-offset":             isNumber,
		"marker-comp-op":            isCompOp,
		"marker-direction":          isKeyword("auto", "auto-down", "left", "right", "left-only", "right-only", "up", "down"),

		"point-file":             isString,
		"point-allow-overlap":    isBool,
		"point-opacity":          isNumber,
		"point-transform":        isString,
		"point-ignore-placement": isBool,
		"point-placement":        isKeyword("centroid", "interior"),
		"point-comp-op":          isCompOp,

		"polygon-fill":               isColor,
		"polygon-gamma":              isNumber,
		"polygon-gamma-method":       isKeyword("power", "linear", "none", "threshold", "multiply"),
		"polygon-opacity":            isNumber,
		"polygon-clip":               isBool,
		"polygon-simplify":           isNumber,
		"polygon-simplify-algorithm": isSimplifyAlgorithm,
		"polygon-smooth":             isNumber,
		"polygon-geometry-transform": isString,
		"polygon-comp-op":            isCompOp,

		"polygon-pattern-alignment":          isKeyword("global", "local"),
		"polygon-pattern-file":               isString,
		"polygon-pattern-gamma":              isNumber,
		"polygon-pattern-opacity":            isNumber,
		"polygon-pattern-clip":               isBool,
		"polygon-pattern-simplify":           isNumber,
		"polygon-pattern-simplify-algorithm": isSimplifyAlgorithm,
		"polygon-pattern-smooth":             isNumber,
		"polygon-pattern-geometry-transform": isString,
		"polygon-pattern-comp-op":            isCompOp,

		"shield-allow-overlap":            isBool,
		"shield-avoid-edges":              isBool,
		"shield-character-spacing":        isNumber,
		"shield-clip":                     isBool,
		"shield-dx":                       isNumber,
		"shield-dy":                       isNumber,
		"shield-face-name":                isStringOrStrings,
		"shield-file":                     isString,
		"shield-fill":                     isColor,
		"shield-halo-fill":                isColor,
		"shield-halo-radius":              isNumber,
		"shield-halo-rasterizer":          isRasterizer,
		"shield-halo-transform":           isString,
		"shield-halo-comp-op":             isCompOp,
		"shield-halo-opacity":             isNumber,
		"shield-line-spacing":             isNumber,
		"shield-min-distance":             isNumber,
		"shield-min-padding":              isNumber,
		"shield-name":                     isString,
		"shield-opacity":                  isNumber,
		"shield-placement":                isKeyword("line", "point", "vertex", "interior"),
		"shield-placement-type":           isKeyword("dummy", "simple", "list"),
		"shield-placements":               isString,
		"shield-transform":                isString,
		"shield-simplify":                 isNumber,
		"shield-simplify-algorithm":       isSimplifyAlgorithm,
		"shield-smooth":                   isNumber,
		"shield-comp-op":                  isCompOp,
		"shield-size":                     isNumber,
		"shield-spacing":                  isNumber,
		"shield-text-dx":                  isNumber,
		"shield-text-dy":                  isNumber,
		"shield-text-opacity":             isNumber,
		"shield-text-transform":           isKeyword("none", "uppercase", "lowercase", "capitalize", "reverse"),
		"shield-wrap-before":              isBool,
		"shield-wrap-character":           isString,
		"shield-wrap-width":               isNumber,
		"shield-unlock-image":             isBool,
		"shield-margin":                   isNumber,
		"shield-repeat-distance":          isNumber,
		"shield-label-position-tolerance": isNumber,
		"shield-horizontal-alignment":     isKeyword("left", "middle", "right", "auto"),
		"shield-vertical-alignment":       isVerticalAlignment,
		"shield-justify-alignment":        isJustifyAlignment,

		"text-allow-overlap":            isBool,
		"text-avoid-edges":              isBool,
		"text-character-spacing":        isNumber,
		"text-clip":                     isBool,
		"text-dx":                       isNumber,
		"text-dy":                       isNumber,
		"text-face-name":                isStringOrStrings,
		"text-font-feature-settings":    isString,
		"text-fill":                     isColor,
		"text-halo-fill":                isColor,
		"text-halo-radius":              isNumber,
		"text-halo-opacity":             isNumber,
		"text-halo-rasterizer":          isRasterizer,
		"text-halo-transform":           isString,
		"text-halo-comp-op":             isCompOp,
		"text-line-spacing":             isNumber,
		"text-min-distance":             isNumber,
		"text-min-padding":              isNumber,
		"text-name":                     isStringOrField,
		"text-opacity":                  isNumber,
		"text-orientation":              isFieldOr(isNumber),
		"text-placement":                isKeyword("line", "point", "vertex", "interior"),
		"text-placement-type":           isKeyword("dummy", "simple", "list"),
		"text-placements":               isString,
		"text-placement-list":           nil, // not validated, as it's directly parsed in decode
		"text-size":                     isNumber,
		"text-spacing":                  isNumber,
		"text-transform":                isKeyword("none", "uppercase", "lowercase", "capitalize", "reverse"),
		"text-wrap-before":              isBool,
		"text-wrap-character":           isString,
		"text-wrap-width":               isNumber,
		"text-repeat-wrap-characater":   isBool,
		"text-ratio":                    isNumber,
		"text-label-position-tolerance": isNumber,
		"text-lang":                     isString,
		"text-max-char-angle-delta":     isNumber,
		"text-vertical-alignment":       isVerticalAlignment,
		"text-horizontal-alignment":     isKeyword("left", "middle", "right", "auto", "adjust"),
		"text-justify-alignment":        isJustifyAlignment,
		"text-margin":                   isNumber,
		"text-repeat-distance":          isNumber,
		"text-min-path-length":          isKeywordOr(isNumber, "auto"),
		"text-rotate-displacement":      isBool,
		"text-upgright":                 isKeyword("auto", "auto-down", "left", "right", "left-only", "right-only"),
		"text-simplify":                 isNumber,
		"text-simplify-algorithm":       isSimplifyAlgorithm,
		"text-smooth":                   isNumber,
		"text-comp-op":                  isCompOp,
		"text-largest-bbox-only":        isBool,

		"raster-opacity":                 isNumber,
		"raster-scaling":                 isScaling,
		"raster-colorizer-default-mode":  isKeyword("discrete", "linear", "exact"),
		"raster-colorizer-default-color": isColor,
		"raster-colorizer-stops":         isStops,
		"raster-comp-op":                 isCompOp,
		"raster-filter-factor":           isNumber,
		"raster-mesh-size":               isNumber,
		"raster-colorizer-epsilon":       isNumber,
	}
}

// validProperty returns whether the property and the value is valid.
func validProperty(property string, value interface{}) (bool, bool) {
	checkFunc, ok := attributeTypes[property]
	if !ok {
		return false, false
	}
	if checkFunc(value) {
		return true, true
	}
	return true, isExpressionString(value)
}
