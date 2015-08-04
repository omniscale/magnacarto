package render

import (
	"errors"
	"net/rpc"
	"os"
	"os/exec"

	"github.com/natefinch/pie"
)

type Mapnik struct {
	client *rpc.Client
}

func NewMapnik() (*Mapnik, error) {
	path, err := exec.LookPath("magnacarto-mapnik")
	if err != nil {
		return nil, err
	}
	client, err := pie.StartProvider(os.Stderr, path)
	if err != nil {
		return nil, err
	}
	return &Mapnik{client: client}, nil
}

func (m *Mapnik) Is3() (bool, error) {
	if m.client == nil {
		return false, errors.New("mapnik plugin not initialized")
	}
	var is3 bool
	err := m.client.Call("Mapnik.Is3", false /* not used */, &is3)
	return is3, err
}

func (m *Mapnik) RegisterFonts(fontDir string) error {
	if m.client == nil {
		return errors.New("mapnik plugin not initialized")
	}
	var tmp interface{}
	err := m.client.Call("Mapnik.RegisterFonts", fontDir, &tmp /* not used */)
	return err
}

func (m *Mapnik) Render(mapfile string, mapReq Request) ([]byte, error) {
	if m.client == nil {
		return nil, errors.New("mapnik plugin not initialized")
	}
	var buf []byte
	err := m.client.Call("Mapnik.Render", struct {
		Mapfile string
		Req     Request
	}{mapfile, mapReq}, &buf)
	return buf, err
}
