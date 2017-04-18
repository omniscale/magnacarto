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
	baseDir         string
	dstMap          Map
	mss             []string
	mmlFile         string
	mmlData         *mml.MML
	locator         config.Locator
	dumpRules       io.Writer
	includeInactive bool
}

// New returns a Builder
func New(mw Map) *Builder {
	return &Builder{dstMap: mw, includeInactive: true}
}

// SetBaseDir sets the base directory for MSS files
func (b *Builder) SetBaseDir(baseDir string) {
	b.baseDir = baseDir
}

// AddMSS adds another mss file to this builder.
func (b *Builder) AddMSS(mss string) {
	b.mss = append(b.mss, mss)
}

// SetMML sets/overwrites the mml file of this builder.
func (b *Builder) SetMML(mml string) {
	b.mmlFile = mml
}

// SetMMLData sets/overwrites the mml data of this builder.
func (b *Builder) SetMMLData(mml *mml.MML) {
	b.mmlData = mml
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
	var err error

	if b.mmlFile != "" {
		r, err := os.Open(b.mmlFile)
		if err != nil {
			return err
		}
		defer r.Close()
		b.mmlData, err = mml.Parse(r)
		if err != nil {
			return err
		}
		if len(b.mss) == 0 {
			var basedir string
			if b.baseDir != "" {
				basedir = b.baseDir
			} else {
				basedir = filepath.Dir(b.mmlFile)
			}
			for _, s := range b.mmlData.Stylesheets {
				b.mss = append(b.mss, filepath.Join(basedir, s))
			}
		}
	}

	if b.mmlData.Layers != nil {
		for _, l := range b.mmlData.Layers {
			layers = append(layers, l)
			layerIDs = append(layerIDs, l.ID)
		}
	}

	carto := mss.New()

	for _, mss := range b.mss {
		if b.mmlFile != "" {
			err = carto.ParseFile(mss)
		} else {
			err = carto.ParseString(mss)
		}
		if err != nil {
			return err
		}
	}

	if err := carto.Evaluate(); err != nil {
		return err
	}

	if b.mmlData.Layers == nil {
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
	SetBackgroundColor(color.Color)
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

// BuildMapFromString parses the style from a string and adds all
// mml.Layers to the map.
func BuildMapFromString(m Map, mml *mml.MML, style string) error {
	carto := mss.New()

	err := carto.ParseString(style)
	if err != nil {
		return err
	}
	if err := carto.Evaluate(); err != nil {
		return err
	}

	for _, l := range mml.Layers {
		rules := carto.MSS().LayerRules(l.ID, l.Classes...)

		if len(rules) > 0 {
			m.AddLayer(l, rules)
		}
	}

	if m, ok := m.(MapOptionsSetter); ok {
		if bgColor, ok := carto.MSS().Map().GetColor("background-color"); ok {
			m.SetBackgroundColor(bgColor)
		}
	}
	return nil
}
