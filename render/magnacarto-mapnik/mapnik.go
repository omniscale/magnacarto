package main

import (
	"fmt"
	"image"
	"image/draw"
	"log"
	"os"
	"sync"
	"time"

	"github.com/natefinch/pie"

	"github.com/omniscale/magnacarto/render"
	"github.com/omniscale/magnacarto/render/mapnik"
)

func renderReq(m *mapnik.Map, mapReq render.Request) ([]byte, error) {
	if useProj4 {
		m.SetSRS(fmt.Sprintf("+init=epsg:%d", mapReq.EPSGCode))
	} else {
		m.SetSRS(fmt.Sprintf("epsg:%d", mapReq.EPSGCode))
	}

	m.Resize(mapReq.Width, mapReq.Height)
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

type API struct {
	// cacheEnabled enables caching of a single mapfile.
	cacheEnabled     bool
	cacheWaitTimeout time.Duration
	mapfile          string
	mu               sync.Mutex
	cache            *mapnik.Map
	lastMTime        time.Time
}

type Args struct {
	Mapfile string
	Req     render.Request
}

func (a *API) Render(args *Args, response *[]byte) error {
	if a.cacheEnabled {
		m, err := a.getMap(args.Mapfile)
		if err != nil {
			return err
		}

		if m != nil {
			tmp, err := renderReq(m, args.Req)
			*response = tmp
			a.mu.Unlock()
			return err
		}
	}
	// cache not enabled or cacheWaitTimeout

	m, err := a.loadMap(args.Mapfile)
	if err != nil {
		return err
	}

	tmp, err := renderReq(m, args.Req)
	if err != nil {
		return err
	}

	*response = tmp
	return nil
}

// getMap returns a cached mapnik.Map. Loads a new map if the mapfile changed
// (filename and/or timestamp of mapfile). Returns nil if the lock for the
// cached map could not be obtained within cacheWaitTimeout.
// Caller needs to mu.Unlock if returned map is not nil.
func (a *API) getMap(mapfile string) (*mapnik.Map, error) {
	fi, err := os.Stat(mapfile)
	if err != nil {
		return nil, fmt.Errorf("checking time of mapfile (%s): %w", mapfile, err)
	}

	deadline := time.Now().Add(a.cacheWaitTimeout)
	for !a.mu.TryLock() {
		if time.Now().After(deadline) {
			return nil, nil
		}
		time.Sleep(10 * time.Millisecond)
	}

	if a.cache != nil && (mapfile != a.mapfile || fi.ModTime() != a.lastMTime) {
		// different mapfile or mtime, release cache
		a.cache.Free()
		a.cache = nil
	}

	// return map from cache
	if a.cache != nil {
		return a.cache, nil
	}

	m, err := a.loadMap(mapfile)
	if err != nil {
		a.mu.Unlock()
		return nil, err
	}

	a.cache = m
	a.mapfile = mapfile
	a.lastMTime = fi.ModTime()
	return m, nil
}

func (a *API) loadMap(mapfile string) (*mapnik.Map, error) {
	m := mapnik.New()
	err := m.Load(mapfile)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// CacheLoadedMap enables caching of a single loaded map object. Render will
// wait for the cached map for up to cacheWaitTimeout, will load a new map
// after timeout.
func (a *API) CacheLoadedMap(cacheWaitTimeout time.Duration, _ *interface{}) error {
	a.cacheEnabled = true
	a.cacheWaitTimeout = cacheWaitTimeout
	return nil
}

func (a *API) Is3(args struct{}, response *bool) error {
	if mapnik.Version.Major == 3 {
		*response = true
	} else {
		*response = false
	}
	return nil
}

func (a *API) RegisterFonts(fontDir string, _ *interface{}) error {
	return mapnik.RegisterFonts(fontDir)
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
	api := API{}
	p := pie.NewProvider()
	if err := p.RegisterName("Mapnik", &api); err != nil {
		log.Fatalf("failed to register Plugin: %s", err)
	}
	p.Serve()
}
