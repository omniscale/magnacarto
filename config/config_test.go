package config

import (
	"reflect"
	"testing"
)

func TestFontVariations(t *testing.T) {
	variations := fontVariations("Foo Sans Bold Oblique", ".ttf")
	if !reflect.DeepEqual(
		variations,
		[]string{
			"FooSansBoldOblique.ttf",
			"Foo-SansBoldOblique.ttf",
			"FooSans-BoldOblique.ttf",
			"FooSansBold-Oblique.ttf",
			"FooSansBold.ttf",
		}) {
		t.Fatal(variations)
	}
}
