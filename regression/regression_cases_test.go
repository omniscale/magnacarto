package regression

import (
    "testing"
)

func Test_010_linestrings_default(t *testing.T) { testIt(t, testCase{Name: "010-linestrings-default"}) }
func Test_011_linestrings_cap_end(t *testing.T) { testIt(t, testCase{Name: "011-linestrings-cap-end"}) }
func Test_012_linestrings_dasharray(t *testing.T) {
	testIt(t, testCase{Name: "012-linestrings-dasharray"})
}
func Test_013_linestrings_labels(t *testing.T) { testIt(t, testCase{Name: "013-linestrings-labels"}) }
func Test_020_polygons_default(t *testing.T)   { testIt(t, testCase{Name: "020-polygons-default"}) }
func Test_021_polygons(t *testing.T)           { testIt(t, testCase{Name: "021-polygons"}) }
func Test_022_polygon_patterns(t *testing.T)   { testIt(t, testCase{Name: "022-polygon-patterns"}) }
func Test_023_polygons_buildings(t *testing.T) { testIt(t, testCase{Name: "023-polygons-buildings"}) }
func Test_030_labels_sizes(t *testing.T)       { testIt(t, testCase{Name: "030-labels-sizes"}) }
func Test_031_label_point(t *testing.T)        { testIt(t, testCase{Name: "031-label-point"}) }
func Test_032_label_point_dxdy(t *testing.T)   { testIt(t, testCase{Name: "032-label-point-dxdy"}) }
func Test_033_label_fields(t *testing.T)       { testIt(t, testCase{Name: "033-label-fields"}) }
func Test_034_label_orientation(t *testing.T)  { testIt(t, testCase{Name: "034-label-orientation"}) }
func Test_040_point_svg(t *testing.T)          { testIt(t, testCase{Name: "040-point-svg"}) }
func Test_041_marker_svg(t *testing.T)         { testIt(t, testCase{Name: "041-marker-svg"}) }
func Test_042_marker_arrow(t *testing.T)       { testIt(t, testCase{Name: "042-marker-arrow"}) }
func Test_043_marker_ellipse(t *testing.T)     { testIt(t, testCase{Name: "043-marker-ellipse"}) }
func Test_050_shields_svg(t *testing.T)        { testIt(t, testCase{Name: "050-shields-svg"}) }
func Test_060_instances(t *testing.T)          { testIt(t, testCase{Name: "060-instances"}) }
func Test_061_classes(t *testing.T)            { testIt(t, testCase{Name: "061-classes"}) }
func Test_062_specifity(t *testing.T)          { testIt(t, testCase{Name: "062-specifity"}) }
