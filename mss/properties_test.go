package mss

import "testing"

func TestMinPrefixPos(t *testing.T) {
	p := Properties{}
	p.setPos(key{name: "line-width"}, 1, position{line: 10, filenum: 1, index: 1})
	p.setPos(key{name: "text-size"}, 1, position{line: 11, filenum: 2, index: 3})
	p.setPos(key{name: "polygon-fill"}, 1, position{line: 20, filenum: 2, index: 4})
	p.setPos(key{name: "polygon-gamma"}, 1, position{line: 23, filenum: 1, index: 2})

	if pos := p.minPrefixPos("polygon-gamma"); len(pos) != 1 || pos[0].index != 2 {
		t.Error("minPos:", pos)
	}
	if pos := p.minPrefixPos("text-size"); len(pos) != 1 || pos[0].index != 3 {
		t.Error("minPos:", pos)
	}
	if pos := p.minPrefixPos("marker-type"); len(pos) != 0 {
		t.Error("minPos:", pos)
	}
}

func TestCombinedProperties(t *testing.T) {
	p1 := &Properties{}
	p1.setPos(key{name: "line-width"}, 0.5, position{line: 10, filenum: 2, index: 5})
	p1.setPos(key{name: "text-size"}, 1, position{line: 11, filenum: 2, index: 6})
	p1.setPos(key{name: "line-cap"}, "butt", position{line: 11, filenum: 2, index: 7})

	p2 := &Properties{}
	p2.setPos(key{name: "line-width"}, 1, position{line: 20, filenum: 1, index: 1})
	p2.setPos(key{name: "text-size"}, 1, position{line: 21, filenum: 1, index: 2})
	p2.setPos(key{name: "line-color"}, "white", position{line: 21, filenum: 1, index: 3})
	p2.setPos(key{name: "line-width"}, 2, position{line: 21, filenum: 1, index: 4})

	p3 := combineProperties(p1, p2)
	if len(p3.values) != 4 {
		t.Error("length of combined properties does not match")
	}

	// value from p1
	if w, _ := p3.GetFloat("line-width"); w != 0.5 {
		t.Error("line-width not from p1", p3)
	}
	// pos from p1
	if pos := p3.minPrefixPos("line-width"); len(pos) != 1 || pos[0].index != 5 {
		t.Error("min-pos", pos)
	}

	// value from p2
	if c, _ := p3.GetString("line-color"); c != "white" {
		t.Error("line-color not from p2", p3)
	}

}
