package render

import (
	"errors"
	"log"
	"net/rpc"
	"os"

	"github.com/natefinch/pie"
)

var mapnikClient *rpc.Client

func StartMapnik() error {
	var err error
	mapnikClient, err = pie.StartProvider(os.Stderr, "./magnacarto-mapnik")
	log.Println("start mapnik", mapnikClient, err)
	return err
}

func MapnikIs3() (bool, error) {
	if mapnikClient == nil {
		return false, errors.New("mapnik plugin not initialized")
	}
	var is3 bool
	err := mapnikClient.Call("Mapnik.Is3", false /* not used */, &is3)
	return is3, err
}

func MapnikRegisterFonts(fontDir string) error {
	if mapnikClient == nil {
		return errors.New("mapnik plugin not initialized")
	}
	var tmp interface{}
	err := mapnikClient.Call("Mapnik.RegisterFonts", fontDir, &tmp /* not used */)
	return err
}

func Mapnik(mapfile string, mapReq Request) ([]byte, error) {
	if mapnikClient == nil {
		return nil, errors.New("mapnik plugin not initialized")
	}
	var buf []byte
	err := mapnikClient.Call("Mapnik.Render", struct {
		Mapfile string
		Req     Request
	}{mapfile, mapReq}, &buf)
	return buf, err
}
