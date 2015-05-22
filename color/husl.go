package color

import "math"

func rgb2husl(r, g, b float64) (float64, float64, float64) {
	return lch2husl(luv2lch(xyz2luv(rgb2xyz(r, g, b))))
}

func husl2rgb(h, s, l float64) (float64, float64, float64) {
	return xyz2rgb(luv2xyz(lch2luv(husl2lch(h, s, l))))
}

func rgb2xyz(r, g, b float64) (x, y, z float64) {
	r = toLinear(r)
	g = toLinear(g)
	b = toLinear(b)
	return dotProduct(m_inv[0], r, g, b), dotProduct(m_inv[1], r, g, b), dotProduct(m_inv[2], r, g, b)
}

func xyz2rgb(x, y, z float64) (float64, float64, float64) {
	r := dotProduct(m[0], x, y, z)
	g := dotProduct(m[1], x, y, z)
	b := dotProduct(m[2], x, y, z)
	return fromLinear(r), fromLinear(g), fromLinear(b)
}

func xyz2luv(x, y, z float64) (l, u, v float64) {
	if x == 0.0 && y == 0.0 && z == 0.0 {
		return 0.0, 0.0, 0.0
	}
	varU := (4.0 * x) / (x + (15.0 * y) + (3.0 * z))
	varV := (9.0 * y) / (x + (15.0 * y) + (3.0 * z))
	l = f(y)
	// Black will create a divide-by-zero error
	if l == 0.0 {
		return 0.0, 0.0, 0.0
	}
	u = 13.0 * l * (varU - refU)
	v = 13.0 * l * (varV - refV)
	return l, u, v
}

func luv2xyz(l, u, v float64) (float64, float64, float64) {
	if l == 0 {
		return 0.0, 0.0, 0.0
	}
	varY := fInv(l)
	varU := u/(13.0*l) + refU
	varV := v/(13.0*l) + refV
	y := varY * refY
	x := 0.0 - (9.0*y*varU)/((varU-4.0)*varV-varU*varV)
	z := (9.0*y - (15.0 * varV * y) - (varV * x)) / (3.0 * varV)
	return x, y, z
}

func luv2lch(l, u, v float64) (float64, float64, float64) {
	c := math.Pow(math.Pow(u, 2)+math.Pow(v, 2), (1.0 / 2.0))
	hrad := math.Atan2(v, u)
	h := hrad * rad2deg
	if h < 0.0 {
		h = 360.0 + h
	}
	return l, c, h
}

func lch2luv(l, c, h float64) (float64, float64, float64) {
	hrad := h * deg2rad
	u := (math.Cos(hrad) * c)
	v := (math.Sin(hrad) * c)
	return l, u, v
}

func lch2husl(l, c, h float64) (float64, float64, float64) {
	if l > 99.9999999 {
		return h, 0.0, 100.0
	}
	if l < 0.00000001 {
		return h, 0.0, 0.0
	}
	mx := maxChromaForLH(l, h)
	s := c / mx * 100.0
	return h, s, l
}

func husl2lch(h, s, l float64) (float64, float64, float64) {
	if l > 99.9999999 {
		return 100, 0.0, h
	}
	if l < 0.00000001 {
		return 0.0, 0.0, h
	}
	mx := maxChromaForLH(l, h)
	c := mx / 100.0 * s
	return l, c, h
}

func maxChromaForLH(l, h float64) float64 {
	hrad := h * deg2rad
	minLength := math.MaxFloat64
	for _, line := range getBounds(l) {
		l := lengthOfRayUntilIntersect(hrad, line[0], line[1])
		if l >= 0.0 && l < minLength {
			minLength = l
		}
	}
	return minLength
}

func lengthOfRayUntilIntersect(theta, m, b float64) float64 {
	// returns <0 for invalid values
	return b / (math.Sin(theta) - m*math.Cos(theta))
}

func getBounds(l float64) [6][2]float64 {
	var sub1, sub2 float64
	sub1 = math.Pow(l+16.0, 3.0) / 1560896.0
	if sub1 > epsilon {
		sub2 = sub1
	} else {
		sub2 = l / kappa
	}

	var ret [6][2]float64
	for i := range m {
		for t := 0; t < 2; t++ { // [0, 1]
			top1 := (284517.0*m[i][0] - 94839.0*m[i][2]) * sub2
			top2 := (838422.0*m[i][2]+769860.0*m[i][1]+731718.0*m[i][0])*l*sub2 - 769860.0*float64(t)*l
			bottom := (632260.0*m[i][2]-126452.0*m[i][1])*sub2 + 126452.0*float64(t)
			ret[i*2+t][0] = top1 / bottom
			ret[i*2+t][1] = top2 / bottom
		}
	}
	return ret
}

func dotProduct(row [3]float64, x, y, z float64) float64 {
	return row[0]*x + row[1]*y + row[2]*z
}

func f(t float64) float64 {
	if t > epsilon {
		return 116*math.Pow(t/refY, 1.0/3.0) - 16.0
	}
	return (t / refY) * kappa
}

func fInv(t float64) float64 {
	if t > 8 {
		return refY * math.Pow((t+16.0)/116.0, 3.0)
	}
	return refY * t / kappa
}

func toLinear(c float64) float64 {
	a := 0.055
	if c > 0.04045 {
		return math.Pow((c+a)/(1.0+a), 2.4)
	} else {
		return c / 12.92
	}
}

func fromLinear(c float64) float64 {
	if c <= 0.0031308 {
		return 12.92 * c
	}
	return (1.055*math.Pow(c, 1.0/2.4) - 0.055)
}

const deg2rad = math.Pi / 180
const rad2deg = 180 / math.Pi

const (
	refX    = 0.95045592705167
	refY    = 1.0
	refZ    = 1.089057750759878
	refU    = 0.19783000664283
	refV    = 0.46831999493879
	kappa   = 903.2962962
	epsilon = 0.0088564516
)

var m = [3][3]float64{
	{3.240969941904521, -1.537383177570093, -0.498610760293},
	{-0.96924363628087, 1.87596750150772, 0.041555057407175},
	{0.055630079696993, -0.20397695888897, 1.056971514242878},
}

var m_inv = [3][3]float64{
	{0.41239079926595, 0.35758433938387, 0.18048078840183},
	{0.21263900587151, 0.71516867876775, 0.072192315360733},
	{0.019330818715591, 0.11919477979462, 0.95053215224966},
}
