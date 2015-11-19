// Package builder implements converter from CartoCSS to Mapnik/MapServer styles.
package builder

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/omniscale/magnacarto/color"
	"github.com/omniscale/magnacarto/config"
	"github.com/omniscale/magnacarto/mml"
	"github.com/omniscale/magnacarto/mss"
)

// Builder builds map styles from MML and MSS files.
type Builder struct {
	dstMap          Map
	mss             []string
	mml             string
	locator         config.Locator
	dumpRules       io.Writer
	includeInactive bool
}

// New returns a Builder
func New(mw Map) *Builder {
	return &Builder{dstMap: mw, includeInactive: true}
}

// AddMSS adds another mss file to this builder.
func (b *Builder) AddMSS(mss string) {
	b.mss = append(b.mss, mss)
}

// SetMML sets/overwirtes the mml file of this builder.
func (b *Builder) SetMML(mml string) {
	b.mml = mml
}

// SetDumpRulesDest enables internal debuging output.
func (b *Builder) SetDumpRulesDest(w io.Writer) {
	b.dumpRules = w
}

// SetIncludeInactive set whether status=off layers should be included in output.
func (b *Builder) SetIncludeInactive(includeInactive bool) {
	b.includeInactive = includeInactive
}

// Build parses MML, MSS files, builds all rules and adds them to the Map.
func (b *Builder) Build() error {
	layerIDs := []string{}
	layers := []mml.Layer{}

	if b.mml != "" {
		r, err := os.Open(b.mml)
		if err != nil {
			return err
		}
		defer r.Close()
		mml, err := mml.Parse(r)
		if err != nil {
			return err
		}
		if len(b.mss) == 0 {
			for _, s := range mml.Stylesheets {
				b.mss = append(b.mss, filepath.Join(filepath.Dir(b.mml), s))
			}
		}

		for _, l := range mml.Layers {
			layers = append(layers, l)
			layerIDs = append(layerIDs, l.ID)
		}
	}

	carto := mss.New()

	for _, mss := range b.mss {
		err := carto.ParseFile(mss)
		if err != nil {
			return err
		}
	}

	if err := carto.Evaluate(); err != nil {
		return err
	}

	if b.mml == "" {
		layerIDs = carto.MSS().Layers()
		for _, layerID := range layerIDs {
			layers = append(layers,
				// XXX assume we only have LineStrings for -mss only export
				mml.Layer{ID: layerID, Type: mml.LineString},
			)
		}
	}

	for _, l := range layers {
		rules := carto.MSS().LayerRules(l.ID, l.Classes...)

		if b.dumpRules != nil {
			for _, r := range rules {
				fmt.Fprintln(b.dumpRules, r.String())
			}
		}
		if len(rules) > 0 && (l.Active || b.includeInactive) {
			b.dstMap.AddLayer(l, rules)
		}
	}

	if m, ok := b.dstMap.(MapOptionsSetter); ok {
		if bgColor, ok := carto.MSS().Map().GetColor("background-color"); ok {
			m.SetBackgroundColor(bgColor)
		}
	}
	return nil
}

type MapOptionsSetter interface {
	SetBackgroundColor(color.RGBA)
}

type Writer interface {
	Write(io.Writer) error
	WriteFiles(basename string) error
}

type Map interface {
	AddLayer(mml.Layer, []mss.Rule)
}

type MapWriter interface {
	Writer
	Map
}
