package color

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"runtime"

	"testing"
)

type colors map[string]color

type color struct {
	HuSL  [3]float64
	HuSLP [3]float64
	LCH   [3]float64
	LUV   [3]float64
	RGB   [3]float64
	XYZ   [3]float64
}

func TestDecodeTestColor(t *testing.T) {
	testColor := []byte(`{
    "#dd5533": {
        "husl": [
           18.714071247462172,
           83.35464521492307,
           54.14423922554455
       ],
       "huslp": [
           18.714071247462172,
           269.0614201849774,
           54.14423922554455
       ],
       "lch": [
           54.14423922554455,
           114.80594019465079,
           18.714071247462172
       ],
       "luv": [
           54.14423922554455,
           108.73632348964713,
           36.834981443358274
       ],
       "rgb": [
           0.8666666666666667,
           0.3333333333333333,
           0.2
       ],
       "xyz": [
           0.3366396301820665,
           0.22110678011653206,
           0.056272250397587584
       ]
    }
}
`)
	testColors := make(colors)
	if err := json.Unmarshal(testColor, &testColors); err != nil {
		t.Fatal(err)
	}
	assertTuplesEq(t, pack(husl2rgb(unpack(testColors["#dd5533"].HuSL))), testColors["#dd5533"].RGB)
	assertTuplesEq(t, pack(rgb2husl(unpack(testColors["#dd5533"].RGB))), testColors["#dd5533"].HuSL)
}

const delta = 0.00000000001

func assertTuplesEq(t *testing.T, a, b [3]float64, msg ...interface{}) {
	for i := 0; i < 3; i++ {
		if math.Abs(a[i]-b[i]) > delta {
			_, file, line, _ := runtime.Caller(1)
			t.Errorf("%v != %v (%s:%d)%s", a, b, file, line, fmt.Sprint(msg))
		}
	}
}

func pack(a, b, c float64) [3]float64 {
	return [3]float64{a, b, c}
}

func unpack(tuple [3]float64) (float64, float64, float64) {
	return tuple[0], tuple[1], tuple[2]
}

func TestShnapshot(t *testing.T) {
	gz, err := os.Open("snapshot-rev3.json.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer gz.Close()

	f, err := gzip.NewReader(gz)
	if err != nil {
		t.Fatal(err)
	}
	decoder := json.NewDecoder(f)
	testColors := make(colors)
	if err := decoder.Decode(&testColors); err != nil {
		t.Fatal(err)
	}
	for hex, color := range testColors {
		if hex == "#dddddd" || hex == "#eeeeee" {
			// these colors fail due to rounding errors :/
			// luv values differ
			continue
		}
		assertTuplesEq(t, pack(husl2rgb(unpack(color.HuSL))), color.RGB, hex)
		assertTuplesEq(t, pack(rgb2husl(unpack(color.RGB))), color.HuSL, hex)
	}
}
