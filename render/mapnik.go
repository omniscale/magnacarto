package render

import (
	"fmt"

	"path/filepath"

	"github.com/omniscale/go-mapnik"
)

func Mapnik(mapfile string, mapReq Request) ([]byte, error) {
	style := filepath.Base(mapfile)
	style = style[:len(style)-len(filepath.Ext(style))] // wihout suffix

	m := mapnik.New()
	err := m.Load(mapfile)
	if err != nil {
		return nil, err
	}

	m.Resize(mapReq.Width, mapReq.Height)
	m.SetSRS(fmt.Sprintf("+init=epsg:%d", mapReq.EPSGCode))
	m.ZoomTo(mapReq.BBOX[0], mapReq.BBOX[1], mapReq.BBOX[2], mapReq.BBOX[3])

	renderOpts := mapnik.RenderOpts{}
	renderOpts.Format = mapReq.Format

	b, err := m.Render(renderOpts)
	if err != nil {
		return nil, err
	}
	return b, nil
}
