package render

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/cgi"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type MapServer struct {
	bin string
}

func NewMapServer() (*MapServer, error) {
	bin, err := exec.LookPath("mapserv")
	if err != nil {
		return nil, err
	}
	return &MapServer{bin: bin}, nil
}

func (m *MapServer) Render(mapfile string, mapReq Request) ([]byte, error) {
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
	q.Set("TRANSPARENT", "false")

	q.Set("BBOX", fmt.Sprintf("%f,%f,%f,%f", mapReq.BBOX[0], mapReq.BBOX[1], mapReq.BBOX[2], mapReq.BBOX[3]))
	q.Set("WIDTH", fmt.Sprintf("%d", mapReq.Width))
	q.Set("HEIGHT", fmt.Sprintf("%d", mapReq.Height))
	q.Set("SRS", fmt.Sprintf("EPSG:%d", mapReq.EPSGCode))
	q.Set("FORMAT", mapReq.Format)

	wd := filepath.Dir(mapfile)
	handler := cgi.Handler{
		Path: m.bin,
		Dir:  wd,
	}

	w := &responseRecorder{
		Body: &bytes.Buffer{},
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

// responseRecorder from net/http/httptest
// copied here to work around global -httptest.server flag from httptest package

// responseRecorder is an implementation of http.ResponseWriter that
// records its mutations for later inspection in tests.
type responseRecorder struct {
	Code      int           // the HTTP response code from WriteHeader
	HeaderMap http.Header   // the HTTP response headers
	Body      *bytes.Buffer // if non-nil, the bytes.Buffer to append written data to
	Flushed   bool

	wroteHeader bool
}

func (rw *responseRecorder) Header() http.Header {
	m := rw.HeaderMap
	if m == nil {
		m = make(http.Header)
		rw.HeaderMap = m
	}
	return m
}

func (rw *responseRecorder) Write(buf []byte) (int, error) {
	if !rw.wroteHeader {
		rw.WriteHeader(200)
	}
	if rw.Body != nil {
		rw.Body.Write(buf)
	}
	return len(buf), nil
}

func (rw *responseRecorder) WriteHeader(code int) {
	if !rw.wroteHeader {
		rw.Code = code
	}
	rw.wroteHeader = true
}

func (rw *responseRecorder) Flush() {
	if !rw.wroteHeader {
		rw.WriteHeader(200)
	}
	rw.Flushed = true
}
