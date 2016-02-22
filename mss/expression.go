package mss

import (
	"fmt"

	"github.com/omniscale/magnacarto/color"
)

type codeType uint8

const (
	typeUnknown codeType = iota
	typeVar
	typeNum
	typePercent
	typeColor
	typeBool
	typeFunction
	typeFunctionEnd
	typeURL
	typeKeyword
	typeField
	typeFieldExpr
	typeString
	typeList
	typeStop

	typeNegation
	typeAdd
	typeSubtract
	typeMultiply
	typeDivide
)

func (t codeType) String() string {
	switch t {
	case typeNegation:
		return "!"
	case typeAdd:
		return "+"
	case typeSubtract:
		return "-"
	case typeMultiply:
		return "*"
	case typeDivide:
		return "/"
	case typeVar:
		return "v"
	case typeNum:
		return "n"
	case typePercent:
		return "%"
	case typeColor:
		return "c"
	case typeBool:
		return "b"
	case typeFunction:
		return "{"
	case typeFunctionEnd:
		return "}"
	case typeURL:
		return "@"
	case typeKeyword:
		return "#"
	case typeField:
		return "["
	case typeList:
		return "L"
	case typeString:
		return "\""
	case typeStop:
		return "S"
	case typeUnknown:
		return "?"
	default:
		return fmt.Sprintf("%#v", t)
	}
}

type expression struct {
	code []code
	pos  position
}

func (e *expression) addOperator(t codeType) {
	e.code = append(e.code, code{T: t})
}

func (e *expression) addValue(val interface{}, t codeType) {
	e.code = append(e.code, code{T: t, Value: val})
}

func (e *expression) clear() {
	e.code = e.code[:0]
}

func (e *expression) evaluate() (Value, error) {
	codes, _, err := evaluate(e.code)
	if err != nil {
		return nil, err
	}
	if len(codes) == 0 {
		return nil, fmt.Errorf("unable to evaluate expression")
	}
	if len(codes) > 1 {
		// create copy since c points to internal code slice
		l := make([]Value, 0, len(codes))
		for _, c := range codes {
			l = append(l, c.Value)
		}
		return l, nil
	}
	return codes[0].Value, nil
}

type Field string

func evaluate(codes []code) ([]code, int, error) {
	top := 0
	for i := 0; i < len(codes); i++ {
		c := codes[i]
		switch c.T {
		case typeNum, typeColor, typePercent, typeString, typeKeyword, typeURL, typeBool, typeField, typeList:
			codes[top] = c
			top++
			continue
		case typeNegation:
			a := codes[top-1]
			a.Value = -a.Value.(float64)
			codes[top-1] = a
			continue
		case typeFunction:
			v, parsed, err := evaluate(codes[i+1:])
			i += parsed + 1
			if err != nil {
				return v, 0, err
			}
			// TODO refactor parameter checking
			if colorF, ok := colorFuncs[c.Value.(string)]; ok {
				if len(v) != 2 {
					return nil, 0, fmt.Errorf("function %s takes exactly two arguments, got %d", c.Value.(string), len(v))
				}
				if v[0].T != typeColor {
					return nil, 0, fmt.Errorf("function %s requires color as first argument, got %v", c.Value.(string), v[0])
				}
				if v[1].T != typeNum && v[1].T != typePercent {
					return nil, 0, fmt.Errorf("function %s requires number/percent as second argument, got %v", c.Value.(string), v[1])
				}
				v = []code{{Value: colorF(v[0].Value.(color.Color), v[1].Value.(float64)/100), T: typeColor}}
			} else if colorP, ok := colorParams[c.Value.(string)]; ok {
				if len(v) != 1 {
					return nil, 0, fmt.Errorf("function %s takes exactly one argument, got %d", c.Value.(string), len(v))
				}
				if v[0].T != typeColor {
					return nil, 0, fmt.Errorf("function %s requires color as argument, got %v", c.Value.(string), v[0])
				}
				v = []code{{Value: colorP(v[0].Value.(color.Color)), T: typeNum}}
			} else if c.Value.(string) == "mix" {
				if len(v) != 3 {
					return nil, 0, fmt.Errorf("function mix takes exactly three arguments, got %d", len(v))
				}
				if v[0].T != typeColor || v[1].T != typeColor {
					return nil, 0, fmt.Errorf("function mix requires color as first and second argument, got %v and %v", v[0], v[1])
				}
				if v[2].T != typeNum && v[2].T != typePercent {
					return nil, 0, fmt.Errorf("function mix requires number/percent as third argument, got %v", v[2])
				}
				v = []code{{Value: color.Mix(v[0].Value.(color.Color), v[1].Value.(color.Color), v[2].Value.(float64)/100), T: typeColor}}
			} else if c.Value.(string) == "-mc-set-hue" {
				if len(v) != 2 {
					return nil, 0, fmt.Errorf("function %s takes exactly two arguments, got %d", c.Value.(string), len(v))
				}
				if v[0].T != typeColor {
					return nil, 0, fmt.Errorf("function %s requires color as first argument, got %v", c.Value.(string), v[0])
				}
				if v[1].T != typeColor {
					return nil, 0, fmt.Errorf("function %s requires color as second argument, got %v", c.Value.(string), v[1])
				}
				v = []code{{Value: color.SetHue(v[0].Value.(color.Color), v[1].Value.(color.Color)), T: typeColor}}
			} else if c.Value.(string) == "greyscale" || c.Value.(string) == "greyscalep" {
				if len(v) != 1 {
					return nil, 0, fmt.Errorf("function %s takes exactly one argument, got %d", c.Value.(string), len(v))
				}
				if v[0].T != typeColor {
					return nil, 0, fmt.Errorf("function %s requires color as argument, got %v", c.Value.(string), v[0])
				}
				if c.Value.(string) == "greyscale" {
					v = []code{{Value: color.Greyscale(v[0].Value.(color.Color)), T: typeColor}}
				} else {
					v = []code{{Value: color.GreyscaleP(v[0].Value.(color.Color)), T: typeColor}}
				}
			} else if c.Value.(string) == "rgb" || c.Value.(string) == "rgba" {
				if c.Value.(string) == "rgb" && len(v) != 3 {
					return nil, 0, fmt.Errorf("rgb takes exactly three arguments, got %d", len(v))
				}
				if c.Value.(string) == "rgba" && len(v) != 4 {
					return nil, 0, fmt.Errorf("rgba takes exactly four arguments, got %d", len(v))
				}
				c := [4]float64{1, 1, 1, 1}
				for i := range v {
					if v[i].T == typeNum {
						if i < 3 {
							c[i] = v[i].Value.(float64) / 255
						} else {
							c[i] = v[i].Value.(float64) // alpha value is from 0.0-1.0
							if c[i] > 1.0 {
								c[i] /= 255 // TODO or clamp? compat with Carto?
							}
						}
					} else if v[i].T == typePercent {
						c[i] = v[i].Value.(float64) / 100
					} else {
						return nil, 0, fmt.Errorf("rgb/rgba takes float or percent arguments only, got %v", v[i])
					}
					if c[i] < 0 {
						c[i] = 0
					} else if c[i] > 255 {
						c[i] = 255
					}
				}
				v = []code{{Value: color.FromRgba(c[0], c[1], c[2], c[3]), T: typeColor}}
			} else if c.Value.(string) == "hsl" || c.Value.(string) == "hsla" {
				if c.Value.(string) == "hsl" && len(v) != 3 {
					return nil, 0, fmt.Errorf("hsl takes exactly three arguments, got %d", len(v))
				}
				if c.Value.(string) == "hsla" && len(v) != 4 {
					return nil, 0, fmt.Errorf("hsla takes exactly four arguments, got %d", len(v))
				}
				c := [4]float64{1, 1, 1, 1}
				for i := range v {
					if v[i].T == typeNum {
						if i == 0 {
							c[i] = v[i].Value.(float64)
						} else {
							c[i] = v[i].Value.(float64) // saturation, lightness, alpha values are from 0.0-1.0
							if c[i] > 1.0 {
								c[i] = 1.0
							} else if c[i] < 0 {
								c[i] = 0
							}
						}
					} else if v[i].T == typePercent {
						if i == 0 {
							c[i] = v[i].Value.(float64) / 360
							if c[i] < 0 {
								c[i] = 0
							} else if c[i] > 100 {
								c[i] = 1.0
							}
						} else {
							c[i] = v[i].Value.(float64) / 100
							if c[i] < 0 {
								c[i] = 0
							} else if c[i] > 100 {
								c[i] = 1.0
							}
						}
					} else {
						return nil, 0, fmt.Errorf("hsl/hsla takes float or percent arguments only, got %v", v[i])
					}
				}
				v = []code{{Value: color.FromHsla(c[0], c[1], c[2], c[3]), T: typeColor}}
			} else if c.Value.(string) == "husl" || c.Value.(string) == "husla" {
				if c.Value.(string) == "husl" && len(v) != 3 {
					return nil, 0, fmt.Errorf("husl takes exactly three arguments, got %d", len(v))
				}
				if c.Value.(string) == "husla" && len(v) != 4 {
					return nil, 0, fmt.Errorf("husla takes exactly four arguments, got %d", len(v))
				}
				c := [4]float64{1, 1, 1, 1}
				for i := range v {
					if v[i].T == typeNum {
						if i == 0 {
							c[i] = v[i].Value.(float64)
						} else {
							c[i] = v[i].Value.(float64) // saturation, lightness, alpha values are from 0.0-1.0
							if c[i] > 1.0 {
								c[i] = 1.0
							} else if c[i] < 0 {
								c[i] = 0
							}
						}
					} else if v[i].T == typePercent {
						if i == 0 {
							c[i] = v[i].Value.(float64) / 360
							if c[i] < 0 {
								c[i] = 0
							} else if c[i] > 100 {
								c[i] = 1.0
							}
						} else {
							c[i] = v[i].Value.(float64) / 100
							if c[i] < 0 {
								c[i] = 0
							} else if c[i] > 100 {
								c[i] = 1.0
							}
						}
					} else {
						return nil, 0, fmt.Errorf("husl/husla takes float or percent arguments only, got %v", v[i])
					}
				}
				v = []code{{Value: color.FromHusl(c[0], c[1], c[2], c[3]), T: typeColor}}
			} else if c.Value.(string) == "stop" {
				if len(v) != 2 {
					return nil, 0, fmt.Errorf("stop takes exactly two arguments, got %d", len(v))
				}
				if v[0].T != typeNum {
					return nil, 0, fmt.Errorf("stop takes int as first argument only, got %v", v[i])
				}
				if v[1].T != typeColor {
					return nil, 0, fmt.Errorf("stop takes color as second argument only, got %v", v[i])
				}
				val := int(v[0].Value.(float64))
				c := v[1].Value.(color.Color)
				v = []code{{
					Value: Stop{Value: val, Color: c},
					T:     typeStop},
				}
			} else if c.Value.(string) == "__echo__" {
				// pass
			} else {
				return nil, 0, fmt.Errorf("unknown function %s", c.Value.(string))
			}
			for i, v := range v {
				codes[top+i] = v
			}
			top += len(v)
		case typeFunctionEnd:
			return codes[0:top], i, nil
		case typeAdd, typeSubtract, typeMultiply, typeDivide:
			a, b := codes[top-2], codes[top-1]
			top -= 2
			if a.T == typeNum && b.T == typeNum {
				switch c.T {
				case typeAdd:
					codes[top] = code{T: typeNum, Value: a.Value.(float64) + b.Value.(float64)}
				case typeSubtract:
					codes[top] = code{T: typeNum, Value: a.Value.(float64) - b.Value.(float64)}
				case typeMultiply:
					codes[top] = code{T: typeNum, Value: a.Value.(float64) * b.Value.(float64)}
				case typeDivide:
					codes[top] = code{T: typeNum, Value: a.Value.(float64) / b.Value.(float64)}
				}
			} else if c.T == typeAdd && a.T == typeString && b.T == typeString {
				// string concatenation
				codes[top] = code{T: typeString, Value: a.Value.(string) + b.Value.(string)}
			} else if c.T == typeAdd && a.T == typeString && b.T == typeField {
				codes[top] = code{T: typeFieldExpr, Value: []Value{a.Value.(string), Field(b.Value.(string))}}
			} else if c.T == typeAdd && a.T == typeField && b.T == typeString {
				codes[top] = code{T: typeFieldExpr, Value: []Value{Field(a.Value.(string)), b.Value.(string)}}
			} else if c.T == typeAdd && a.T == typeField && b.T == typeField {
				codes[top] = code{T: typeFieldExpr, Value: []Value{Field(a.Value.(string)), Field(b.Value.(string))}}
			} else if c.T == typeAdd && a.T == typeFieldExpr && b.T == typeField {
				codes[top] = code{T: typeFieldExpr, Value: append(a.Value.([]Value), Field(b.Value.(string)))}
			} else if c.T == typeAdd && a.T == typeFieldExpr && b.T == typeString {
				codes[top] = code{T: typeFieldExpr, Value: append(a.Value.([]Value), b.Value.(string))}
			} else if c.T == typeMultiply && a.T == typeColor && b.T == typeNum {
				c := a.Value.(color.Color)
				f := b.Value.(float64)
				c = color.Multiply(c, f)
				codes[top] = code{T: typeColor, Value: c}
			} else {
				return nil, 0, fmt.Errorf("unsupported operation %v for %v and %v", c, a, b)
			}
			top++
		}
	}
	return codes[:top], 0, nil
}

type Stop struct {
	Value int
	Color color.Color
}

type functype func(args []code) ([]code, error)

var colorFuncs map[string]colorFunc
var colorParams map[string]colorParam

type colorFunc func(color.Color, float64) color.Color
type colorParam func(color.Color) float64

func init() {
	colorFuncs = map[string]colorFunc{
		"lighten":     color.Lighten,
		"lightenp":    color.LightenP,
		"darken":      color.Darken,
		"darkenp":     color.DarkenP,
		"saturate":    color.Saturate,
		"saturatep":   color.SaturateP,
		"desaturate":  color.Desaturate,
		"desaturatep": color.DesaturateP,
		"fadein":      color.FadeIn,
		"fadeout":     color.FadeOut,
		"spin":        color.Spin,
		"spinp":       color.SpinP,
	}

	colorParams = map[string]colorParam{
		"hue":         color.Hue,
		"huep":        color.HueP,
		"lightness":   color.Lightness,
		"lightnessp":  color.LightnessP,
		"saturation":  color.Saturation,
		"saturationp": color.SaturationP,
		"alpha":       color.Alpha,
	}
}

type code struct {
	T     codeType
	Value interface{}
}
