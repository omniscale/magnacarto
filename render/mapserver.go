package render

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/cgi"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func MapServer(bin, mapfile string, mapReq Request) ([]byte, error) {
	if !filepath.IsAbs(mapfile) {
		if wd, err := os.Getwd(); err == nil {
			mapfile = filepath.Join(wd, mapfile)
		}
	}

	q := url.Values{}
	q.Set("REQUEST", "GetMap")
	q.Set("SERVICE", "WMS")
	q.Set("VERSION", "1.1.1")
	q.Set("STYLES", "")
	q.Set("LAYERS", "map")
	q.Set("MAP", mapfile)

	q.Set("BBOX", fmt.Sprintf("%f,%f,%f,%f", mapReq.BBOX[0], mapReq.BBOX[1], mapReq.BBOX[2], mapReq.BBOX[3]))
	q.Set("WIDTH", fmt.Sprintf("%d", mapReq.Width))
	q.Set("HEIGHT", fmt.Sprintf("%d", mapReq.Height))
	q.Set("SRS", fmt.Sprintf("EPSG:%d", mapReq.EPSGCode))
	q.Set("FORMAT", mapReq.Format)

	if !filepath.IsAbs(bin) {
		for _, path := range strings.Split(os.Getenv("PATH"), string(filepath.ListSeparator)) {
			if fi, _ := os.Stat(filepath.Join(path, bin)); fi != nil {
				bin = filepath.Join(path, bin)
				break
			}
		}
	}

	wd := filepath.Dir(mapfile)
	handler := cgi.Handler{
		Path: bin,
		Dir:  wd,
	}

	w := &httptest.ResponseRecorder{
		Body: bytes.NewBuffer(nil),
	}

	req, err := http.NewRequest("GET", "/?"+q.Encode(), nil)
	if err != nil {
		return nil, err
	}

	handler.ServeHTTP(w, req)

	if w.Code != 200 {
		return nil, fmt.Errorf("error while calling mapserv CGI (status %d): %v", w.Code, string(w.Body.Bytes()))
	}
	if ct := w.Header().Get("Content-type"); ct != "" && !strings.HasPrefix(ct, "image") {
		return nil, fmt.Errorf(" mapserv CGI did not return image (%v)\n%v", w.Header(), string(w.Body.Bytes()))
	}
	return w.Body.Bytes(), nil
}
