package main

import (
	"fmt"
	"image"
	"image/draw"
	"log"

	"github.com/natefinch/pie"

	"path/filepath"

	"github.com/omniscale/go-mapnik/v2"
	"github.com/omniscale/magnacarto/render"
)

func renderReq(mapfile string, mapReq render.Request) ([]byte, error) {
	style := filepath.Base(mapfile)
	style = style[:len(style)-len(filepath.Ext(style))] // wihout suffix

	m := mapnik.NewSized(mapReq.Width, mapReq.Height)
	defer m.Free()
	err := m.Load(mapfile)
	if err != nil {
		return nil, err
	}

	if useProj4 {
		m.SetSRS(fmt.Sprintf("+init=epsg:%d", mapReq.EPSGCode))
	} else {
		m.SetSRS(fmt.Sprintf("epsg:%d", mapReq.EPSGCode))
	}

	m.ZoomTo(mapReq.BBOX[0], mapReq.BBOX[1], mapReq.BBOX[2], mapReq.BBOX[3])

	renderOpts := mapnik.RenderOpts{}
	renderOpts.Format = mapReq.Format
	renderOpts.ScaleFactor = mapReq.ScaleFactor

	if mapReq.BGColor == nil {
		b, err := m.Render(renderOpts)
		if err != nil {
			return nil, err
		}
		return b, nil
	}

	// draw image on requested bgcolor
	img, err := m.RenderImage(renderOpts)
	if err != nil {
		return nil, err
	}

	result := image.NewNRGBA(img.Bounds())
	bg := image.NewUniform(mapReq.BGColor)
	draw.Draw(result, img.Bounds(), bg, image.ZP, draw.Src)
	draw.Draw(result, img.Bounds(), img, image.ZP, draw.Over)

	b, err := mapnik.Encode(result, mapReq.Format)
	if err != nil {
		return nil, err
	}
	return b, nil
}

type api struct {
}

type Args struct {
	Mapfile string
	Req     render.Request
}

func (api) Render(args *Args, response *[]byte) error {
	tmp, err := renderReq(args.Mapfile, args.Req)
	*response = tmp
	return err
}

func (api) Is3(args struct{}, response *bool) error {
	if mapnik.Version.Major == 3 {
		*response = true
	} else {
		*response = false
	}
	return nil
}

func (api) RegisterFonts(fontDir string, _ *interface{}) error {
	mapnik.RegisterFonts(fontDir)
	return nil
}

var useProj4 bool // useProj4 determines whether we use old +init syntax or not

func checkProjVersion() {
	m := mapnik.NewSized(1, 1)
	defer m.Free()
	m.SetSRS("epsg:4326") // this should fail with Proj4
	m.ZoomTo(-1, -1, 0, 0)
	_, err := m.Render(mapnik.RenderOpts{})
	if err != nil {
		useProj4 = true
	}
}

func main() {
	checkProjVersion()

	p := pie.NewProvider()
	if err := p.RegisterName("Mapnik", api{}); err != nil {
		log.Fatalf("failed to register Plugin: %s", err)
	}
	p.Serve()
}
