// Package mss implements a CartoCSS parser and rule generator.
package mss

import "math"

type Value interface{}

type MSS struct {
	root  block
	stack []*block
	base  block
}

// Map returns properties of the single Map{} block.
func (m *MSS) Map() *Properties {
	if m.base.properties != nil {
		return m.base.properties
	}
	return &Properties{}
}

func (m *MSS) current() *block {
	return m.stack[len(m.stack)-1]
}

func newMSS() *MSS {
	m := MSS{}
	m.stack = []*block{&m.root}
	return &m
}

func (m *MSS) addLayer(layer string) {
	s := m.current().currentSelector()
	s.Layer = layer
}

func (m *MSS) addAttachment(attachment string) {
	s := m.current().currentSelector()
	s.Attachment = attachment
}

func (m *MSS) addClass(class string) {
	s := m.current().currentSelector()
	s.Class = class
}

func (m *MSS) addFilter(field string, compOp CompOp, value interface{}) {
	s := m.current().currentSelector()
	f := Filter{field, compOp, value}
	s.Filters = append(s.Filters, f)
}

func (m *MSS) addZoom(comp CompOp, level int64) {
	if level > math.MaxInt8 || level < 0 {
		// TODO
		panic("zoom not between 0 and 30")
	}
	s := m.current().currentSelector()
	if s.Zoom != InvalidZoom {
		s.Zoom = s.Zoom.add(comp, int8(level))
	} else {
		s.Zoom = newZoomRange(comp, level)
	}
}

func (m *MSS) pushSelector() {
	b := m.current()
	b.selectors = append(b.selectors, &Selector{Zoom: AllZoom})
}

func (m *MSS) setProperty(property string, val Value, pos position) {
	m.current().addProperty(property, val, pos)
}

func (m *MSS) setInstance(instance string) {
	m.current().instance = instance
}

func (m *MSS) pushMapBlock() {
	m.stack = append(m.stack, &m.base)
}

func (m *MSS) pushBlock() {
	b := &block{}
	current := m.stack[len(m.stack)-1]
	current.blocks = append(current.blocks, b)
	m.stack = append(m.stack, b)
}

func (m *MSS) popBlock() {
	m.stack = m.stack[:len(m.stack)-1]
}

type block struct {
	selectors  []*Selector
	properties *Properties
	instance   string
	blocks     []*block
}

func (b *block) addProperty(property string, val Value, pos position) {
	if b.properties == nil {
		b.properties = &Properties{}
	}
	b.properties.setPos(key{name: property, instance: b.instance}, val, pos)
	b.instance = ""
}

func (b *block) currentSelector() *Selector {
	return b.selectors[len(b.selectors)-1]
}
