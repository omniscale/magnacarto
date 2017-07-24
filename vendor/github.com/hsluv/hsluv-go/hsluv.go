package hsluv

import (
    "fmt"
    "math"
    "strconv"
    "strings"
)

func HsluvToHex(h, s, l float64) (string) {
    return convRgbHex(convHsluvRgb(h, s, l))
}

func HsluvToRGB(h, s, l float64) (float64, float64, float64) {
    return convHsluvRgb(h, s, l)
}

func HsluvFromHex(hex string) (float64, float64, float64) {
    return convRgbHsluv(convHexRgb(hex))
}

func HsluvFromRGB(r, g, b float64) (float64, float64, float64) {
    return convRgbHsluv(r, g, b)
}

func HpluvToHex(h, s, l float64) (string) {
    return convRgbHex(convXyzRgb(convLuvXyz(convLchLuv(convHpluvLch(h, s, l)))))
}

func HpluvToRGB(h, s, l float64) (float64, float64, float64) {
    return convXyzRgb(convLuvXyz(convLchLuv(convHpluvLch(h, s, l))))
}

func HpluvFromHex(hex string) (float64, float64, float64) {
    return convLchHpluv(convLuvLch(convXyzLuv(convRgbXyz(convHexRgb(hex)))))
}

func HpluvFromRGB(r, g, b float64) (float64, float64, float64) {
    return convLchHpluv(convLuvLch(convXyzLuv(convRgbXyz(r, g, b))))
}

var m = [3][3]float64 {
    {3.2409699419045214, -1.5373831775700935, -0.49861076029300328},
    {-0.96924363628087983, 1.8759675015077207, 0.041555057407175613},
    {0.055630079696993609, -0.20397695888897657, 1.0569715142428786},
}

var m_inv =  [3][3]float64 {
    {0.41239079926595948, 0.35758433938387796, 0.18048078840183429},
    {0.21263900587151036, 0.71516867876775593, 0.072192315360733715},
    {0.019330818715591851, 0.11919477979462599, 0.95053215224966058},
}

const refU = 0.19783000664283681
const refV = 0.468319994938791
const kappa = 903.2962962962963
const epsilon = 0.0088564516790356308

func convLchRgb(l, c, h float64) (float64, float64, float64) {
    return convXyzRgb(convLuvXyz(convLchLuv(l, c, h)))
}

func convRgbLch(r, g, b float64) (float64, float64, float64) {
    return convLuvLch(convXyzLuv(convRgbXyz(r, g, b)))
}

func convHsluvRgb(h, s, l float64) (float64, float64, float64) {
    return convLchRgb(convHsluvLch(h, s, l))
}

func convRgbHsluv(r, g, b float64) (float64, float64, float64) {
    return convLchHsluv(convRgbLch(r, g, b))
}

func convXyzLuv(x, y, z float64) (float64, float64, float64) {
    var l, u, v float64
    if y == 0 {
        return l, u, v
    }
    l = yToL(y)
    varU := (4.0 * x) / (x + (15.0 * y) + (3.0 * z))
    varV := (9.0 * y) / (x + (15.0 * y) + (3.0 * z))
    u = 13.0 * l * (varU - refU)
    v = 13.0 * l * (varV - refV)
    return l, u, v
}

func convLuvXyz(l, u, v float64) (float64, float64, float64) {
    var x, y, z float64
    if (l == 0) {
        return x, y, z
    }
    varU := u / (13.0 * l) + refU
    varV := v / (13.0 * l) + refV
    y = lToY(l)
    x = 0.0 - (9.0 * y * varU) / ((varU - 4.0) * varV - varU * varV)
    z = (9.0 * y - (15.0 * varV * y) - (varV * x)) / (3.0 * varV)
    return x, y, z
}

func convLuvLch(l, u, v float64) (float64, float64, float64) {
    var hRad, h float64
    c := math.Sqrt(math.Pow(u, 2) + math.Pow(v, 2))
    if c >= 0.00000001 {
        hRad = math.Atan2(v, u)
        h = hRad * 360.0 / 2.0 / math.Pi
        if h < 0.0 {
            h = 360.0 + h
        }
    }
    return l, c, h
}

func convLchLuv(l, c, h float64) (float64, float64, float64) {
    hRad := h / 360.0 * 2.0 * math.Pi
    u := math.Cos(hRad) * c
    v := math.Sin(hRad) * c
    return l, u, v
};

func convHsluvLch(h, s, l float64) (float64, float64, float64) {
    var c, max float64
    if l > 99.9999999 || l < 0.00000001 {
        c = 0.0
    } else {
        max = maxChromaForLH(l, h)
        c = max / 100.0 * s;
    }
    return l, c, h
}

func convLchHsluv(l, c, h float64) (float64, float64, float64) {
    var s, max float64
    if l > 99.9999999 || l < 0.00000001 {
        s = 0.0
    } else {
        max = maxChromaForLH(l, h)
        s = c / max * 100.0
    }
    return h, s, l
}

func convHpluvLch(h, s, l float64) (float64, float64, float64) {
    var c, max float64
    if l > 99.9999999 || l < 0.00000001 {
        c = 0.0
    } else {
        max = maxSafeChromaForL(l)
        c = max / 100.0 * s
    }
    return l, c, h
}

func convLchHpluv(l, c, h float64) (float64, float64, float64) {
    var s, max float64
    if l > 99.9999999 || l < 0.00000001 {
        s = 0.0
    } else {
        max = maxSafeChromaForL(l)
        s = c / max * 100.0
    }
    return h, s, l
}

func convRgbHex(r, g, b float64) (hex string) {
    rV := round(math.Max(0.0, math.Min(r, 1)) * 255.0)
    gV := round(math.Max(0.0, math.Min(g, 1)) * 255.0)
    bV := round(math.Max(0.0, math.Min(b, 1)) * 255.0)

    hex = fmt.Sprintf("#%02x%02x%02x", rV, gV, bV)
    return
}

func convHexRgb(hex string) (r float64, g float64, b float64) {
    if strings.HasPrefix(hex, "#") {
        hex = hex[1:len(hex)]
    }
    rV, err := strconv.ParseInt(hex[0:2], 16, 0)
    gV, err := strconv.ParseInt(hex[2:4], 16, 0)
    bV, err := strconv.ParseInt(hex[4:6], 16, 0)

    if err == nil {
        r = float64(rV) / 255.0
        g = float64(gV) / 255.0
        b = float64(bV) / 255.0
    }
    return
}

func convXyzRgb(x, y, z float64) (r, g, b float64) {
  r = fromLinear(dotProduct(m[0], [3]float64{x, y, z}))
  g = fromLinear(dotProduct(m[1], [3]float64{x, y, z}))
  b = fromLinear(dotProduct(m[2], [3]float64{x, y, z}))
  return
}

func convRgbXyz(r, g, b float64) (x, y, z float64) {
    r = toLinear(r)
    g = toLinear(g)
    b = toLinear(b)
    x = dotProduct(m_inv[0], [3]float64{r, g, b})
    y = dotProduct(m_inv[1], [3]float64{r, g, b})
    z = dotProduct(m_inv[2], [3]float64{r, g, b})
    return
}

func fromLinear(c float64) (float64) {
    if c <= 0.0031308 {
        return 12.92 * c
    } else {
        return 1.055 * math.Pow(c, 1.0 / 2.4) - 0.055
    }
}

func toLinear(c float64) (float64) {
    const a = 0.055
    if c > 0.04045 {
        return math.Pow((c + a) / (1.0 + a), 2.4)
    } else {
        return c / 12.92
    }
}

func yToL(y float64) (float64) {
    if y <= epsilon {
        return y * kappa
    } else {
        return 116.0 * math.Pow(y, 1.0 / 3.0) - 16.0
    }
}

func lToY(l float64) (float64) {
    if l <= 8 {
        return l / kappa;
    } else {
        return math.Pow((l + 16.0) / 116.0, 3.0)
    }
}

func maxSafeChromaForL(l float64) (float64) {
    minLength := math.MaxFloat64
    for _, line := range getBounds(l) {
      m1 := line[0]
      b1 := line[1]
      x := intersectLineLine(m1, b1, -1.0 / m1, 0.0)
      dist := distanceFromPole(x, b1 + x * m1)
      if dist < minLength {
          minLength = dist
      }
    }
    return minLength
}

func maxChromaForLH(l, h float64) (float64) {
    hRad := h / 360.0 * math.Pi * 2.0
    minLength := math.MaxFloat64
    for _, line := range getBounds(l) {
        length := lengthOfRayUntilIntersect(hRad, line[0], line[1])
        if length > 0.0 && length < minLength {
            minLength = length
        }
    }
    return minLength
}

func getBounds(l float64) [6][2]float64 {
    var sub2 float64
    var ret [6][2]float64
    sub1 := math.Pow(l + 16.0, 3.0) / 1560896.0
    if sub1 > epsilon {
        sub2 = sub1
    } else {
        sub2 = l / kappa
    }
    for i := range m {
        for k := 0; k < 2; k++ {
            top1 := (284517.0 * m[i][0] - 94839.0 * m[i][2]) * sub2
            top2 := (838422.0 * m[i][2] + 769860.0 * m[i][1] + 731718.0 * m[i][0]) * l * sub2 - 769860.0 * float64(k) * l
            bottom := (632260.0 * m[i][2] - 126452.0 * m[i][1]) * sub2 + 126452.0 * float64(k)
            ret[i * 2 + k][0] = top1 / bottom
            ret[i * 2 + k][1] = top2 / bottom
        }
    }
    return ret
}

func intersectLineLine(x1, y1, x2, y2 float64) (float64) {
    return (y1 - y2) / (x2 - x1)
}

func distanceFromPole(x, y float64) (float64) {
    return math.Sqrt(math.Pow(x, 2.0) + math.Pow(y, 2.0))
}

func lengthOfRayUntilIntersect(theta, x, y float64) (length float64) {
    length = y / (math.Sin(theta) - x * math.Cos(theta))
    return
}

func dotProduct(a, b [3]float64) (dot float64) {
    for i := 0; i < 3; i++ {
        dot += a[i] * b[i]
    }
    return
}

func round(f float64) int {
    if math.Abs(f) < 0.5 {
        return 0
    }
    return int(f + math.Copysign(0.5, f))
}
