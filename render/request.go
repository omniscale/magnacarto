package render

import "image/color"

type Request struct {
	Width       int
	Height      int
	BBOX        [4]float64
	EPSGCode    int
	Format      string
	ScaleFactor float64
	BGColor     *color.RGBA
}
