package husl

import (
    "encoding/json"
    "fmt"
    "math"
    "os"
    "testing"
)

type mapping map[string]values

type values struct {
    Rgb [3]float64
    Xyz [3]float64
    Luv [3]float64
    Lch [3]float64
    Husl [3]float64
    Huslp [3]float64
}

func pack(a, b, c float64) [3]float64 {
	return [3]float64{a, b, c}
}

func unpack(tuple [3]float64) (float64, float64, float64) {
	return tuple[0], tuple[1], tuple[2]
}

const delta = 0.00000000001

func compareTuple(t *testing.T, result, expected [3]float64, method string, hex string) {
    var error bool
    var errs [3]bool
    for i := 0; i < 3; i++ {
        if math.Abs(result[i] - expected[i]) > delta {
            error = true
            errs[i] = true
        }
    }
    if error {
        resultOutput := "["
        for i := 0; i < 3; i++ {
            resultOutput += fmt.Sprintf("%f", result[i])
            if errs[i] {
                resultOutput += " *"
            }
            if i < 2 {
                resultOutput += ", "
            }
        }
        resultOutput += "]"
        t.Errorf("result: %s expected: %v, testing %s with test case %s", resultOutput, expected, method, hex)
    }
}

func compareHex(t *testing.T, result, expected string, method string, hex string) {
    if result != expected {
        t.Errorf("result: %v expected: %v, testing %s with test case %s", result, expected, method, hex)
    }
}

func TestSnapshot(t *testing.T) {
    snapshotFile, err := os.Open("snapshot-rev4.json")
    if err != nil {
        t.Fatal(err)
    }
    defer snapshotFile.Close()

    jsonParser := json.NewDecoder(snapshotFile)
    snapshot := make(mapping)
    if err = jsonParser.Decode(&snapshot); err != nil {
        t.Fatal(err)
    }

    for hex, colorValues := range snapshot {
        // tests for public methods
        if testing.Verbose() {
            t.Logf("Testing public methods for test case %s", hex)
        }

        compareHex(t, HuslToHex(unpack(colorValues.Husl)), hex, "HuslToHex", hex)
        compareTuple(t, pack(HuslToRGB(unpack(colorValues.Husl))), colorValues.Rgb, "HuslToRGB", hex)
        compareTuple(t, pack(HuslFromHex(hex)), colorValues.Husl, "HuslFromHex", hex)
        compareTuple(t, pack(HuslFromRGB(unpack(colorValues.Rgb))), colorValues.Husl, "HuslFromRGB", hex)
        compareHex(t, HuslpToHex(unpack(colorValues.Huslp)), hex, "HuslpToHex", hex)
        compareTuple(t, pack(HuslpToRGB(unpack(colorValues.Huslp))), colorValues.Rgb, "HuslpToRGB", hex)
        compareTuple(t, pack(HuslpFromHex(hex)), colorValues.Huslp, "HuslpFromHex", hex)
        compareTuple(t, pack(HuslpFromRGB(unpack(colorValues.Rgb))), colorValues.Huslp, "HuslpFromRGB", hex)

        if !testing.Short() {
            // internal tests
            if testing.Verbose() {
                t.Logf("Testing internal methods for test case %s", hex)
            }

            compareTuple(t, pack(convLchRgb(unpack(colorValues.Lch))), colorValues.Rgb, "convLchRgb", hex)
            compareTuple(t, pack(convRgbLch(unpack(colorValues.Rgb))), colorValues.Lch, "convRgbLch", hex)
            compareTuple(t, pack(convXyzLuv(unpack(colorValues.Xyz))), colorValues.Luv, "convXyzLuv", hex)
            compareTuple(t, pack(convLuvXyz(unpack(colorValues.Luv))), colorValues.Xyz, "convLuvXyz", hex)
            compareTuple(t, pack(convLuvLch(unpack(colorValues.Luv))), colorValues.Lch, "convLuvLch", hex)
            compareTuple(t, pack(convLchLuv(unpack(colorValues.Lch))), colorValues.Luv, "convLchLuv", hex)
            compareTuple(t, pack(convHuslLch(unpack(colorValues.Husl))), colorValues.Lch, "convHuslLch", hex)
            compareTuple(t, pack(convLchHusl(unpack(colorValues.Lch))), colorValues.Husl, "convLchHusl", hex)
            compareTuple(t, pack(convHuslpLch(unpack(colorValues.Huslp))), colorValues.Lch, "convHuslpLch", hex)
            compareTuple(t, pack(convLchHuslp(unpack(colorValues.Lch))), colorValues.Huslp, "convLchHuslp", hex)
            compareHex(t, convRgbHex(unpack(colorValues.Rgb)), hex, "convRgbHex", hex)
            compareTuple(t, pack(convHexRgb(hex)), colorValues.Rgb, "convHexRgb", hex)
            compareTuple(t, pack(convXyzRgb(unpack(colorValues.Xyz))), colorValues.Rgb, "convXyzRgb", hex)
            compareTuple(t, pack(convRgbXyz(unpack(colorValues.Rgb))), colorValues.Xyz, "convRgbXyz", hex)
        }
    }
}
