package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"code.google.com/p/go.net/websocket"

	"github.com/omniscale/magnacarto"
	"github.com/omniscale/magnacarto/builder"
	mapnikBuilder "github.com/omniscale/magnacarto/builder/mapnik"
	"github.com/omniscale/magnacarto/builder/mapserver"
	"github.com/omniscale/magnacarto/config"
	"github.com/omniscale/magnacarto/maps"
	mssPkg "github.com/omniscale/magnacarto/mss"
	"github.com/omniscale/magnacarto/render"

	"github.com/omniscale/go-mapnik"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type magnaserv struct {
	config       *config.Magnacarto
	builderCache *builder.Cache
	mapnikMaker  builder.MapMaker
}

func (s *magnaserv) styleParams(r *http.Request) (mml string, mss []string) {
	mml = r.FormValue("mml")
	if mml != "" {
		mml = filepath.Join(s.config.StylesDir, mml)
	}

	mssList := r.FormValue("mss")
	if mssList != "" {
		for _, f := range strings.Split(mssList, ",") {
			mss = append(mss, filepath.Join(s.config.StylesDir, f))
		}
	}

	return mml, mss
}

func (s *magnaserv) render(w http.ResponseWriter, r *http.Request) {
	mapReq, err := maps.ParseMapRequest(r)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var maker builder.MapMaker
	renderer := mapReq.Query.Get("RENDERER")
	if renderer == "mapserver" {
		maker = mapserver.Maker
	} else {
		maker = s.mapnikMaker
	}

	styleFile := mapReq.Query.Get("FILE")
	if styleFile == "" {
		mml, mss := s.styleParams(r)
		if mml == "" {
			log.Println("missing mml param in request")
			http.Error(w, "missing mml param", http.StatusBadRequest)
			return
		}

		styleFile, err = s.builderCache.StyleFile(maker, mml, mss)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	var b []byte
	if renderer == "mapserver" {
		mapReq.Format = mapReq.Query.Get("FORMAT") // use requested format, not internal mapnik format
		b, err = render.MapServer(s.config.MapServer.DevBin, styleFile, renderReq(mapReq))
	} else {
		b, err = render.Mapnik(styleFile, renderReq(mapReq))

	}
	if err != nil {
		s.internalError(w, r, err)
		return
	}

	w.Header().Add("Content-Type", "image/png")
	w.Header().Add("Content-Length", strconv.FormatUint(uint64(len(b)), 10))

	io.Copy(w, bytes.NewBuffer(b))
}

func (s *magnaserv) projects(w http.ResponseWriter, r *http.Request) {
	projects, err := findProjects(s.config.StylesDir)
	if err != nil {
		s.internalError(w, r, err)
		return
	}

	w.Header().Add("Content-Type", "application/json")

	enc := json.NewEncoder(w)
	err = enc.Encode(struct {
		Projects []project `json:"projects"`
	}{Projects: projects})
	if err != nil {
		s.internalError(w, r, err)
		return
	}
}

func (s *magnaserv) internalError(w http.ResponseWriter, r *http.Request, err error) {
	log.Print(err)
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("internal error"))
}

func renderReq(r *maps.Request) render.Request {
	result := render.Request{}
	result.BBOX[0] = r.BBOX.MinX
	result.BBOX[1] = r.BBOX.MinY
	result.BBOX[2] = r.BBOX.MaxX
	result.BBOX[3] = r.BBOX.MaxY
	result.Width = r.Width
	result.Height = r.Height
	result.EPSGCode = r.EPSGCode
	result.Format = r.Format
	return result
}

func (s *magnaserv) changes(ws *websocket.Conn) {
	mml, mss := s.styleParams(ws.Request())
	if mml == "" {
		log.Println("missing mml param in request")
		ws.Close()
		return
	}

	var maker builder.MapMaker

	renderer := ws.Request().Form.Get("renderer")
	if renderer == "mapserver" {
		maker = mapserver.Maker
	} else {
		maker = s.mapnikMaker
	}

	done := make(chan struct{})
	updatec, errc := s.builderCache.Notify(maker, mml, mss, done)

	// read and discard anything from client, signal close on any error
	closeWs := make(chan struct{})
	go func() {
		readbuf := make([]byte, 0, 16)
		for {
			if _, err := ws.Read(readbuf); err != nil {
				if err != io.EOF {
					log.Println("ws read err: ", err)
				}
				closeWs <- struct{}{}
				return
			}
		}
	}()

	var lastMsgSent time.Time
	var lastMsg []byte
	for {
		select {
		case <-closeWs:
			done <- struct{}{}
			ws.Close()
			return
		case update := <-updatec:
			var msg []byte
			var err error
			if update.Err != nil {
				if parseErr, ok := update.Err.(*mssPkg.ParseError); ok {
					msg, err = json.Marshal(struct {
						FullError string `json:"full_error"`
						Error     string `json:"error"`
						Filename  string `json:"filename"`
						Line      int    `json:"line"`
						Column    int    `json:"column"`
					}{parseErr.Error(), parseErr.Err, parseErr.Filename, parseErr.Line, parseErr.Column})
				} else {
					msg, err = json.Marshal(struct {
						Error string `json:"error"`
					}{update.Err.Error()})
				}
			} else {
				msg, err = json.Marshal(struct {
					UpdatedAt time.Time `json:"updated_at"`
				}{update.Time})

			}
			if err != nil {
				log.Println("error encoding json", err)
				ws.Close()
				return
			}
			// debounce notifications
			if !lastMsgSent.Add(2*time.Second).After(time.Now()) || bytes.Compare(msg, lastMsg) != 0 {
				if _, err := ws.Write(msg); err != nil {
					done <- struct{}{}
					ws.Close()
					return
				}
				lastMsg = msg
				lastMsgSent = time.Now()
			}
		case err := <-errc:
			log.Println("error:", err)
			ws.Close()
			return
		}
	}
}

func main() {
	if os.Getenv("GOMAXPROCS") == "" {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}

	var listenAddr = flag.String("listen", "localhost:7070", "listen address")
	var configFile = flag.String("config", "magnaserv.tml", "config")
	var version = flag.Bool("version", false, "print version and exit")

	flag.Parse()

	if *version {
		fmt.Println(magnacarto.Version)
		os.Exit(0)
	}

	conf, err := config.Load(*configFile)
	if err != nil {
		log.Fatal(err)
	}

	locator := conf.Locator()

	builderCache := builder.NewCache(locator)
	if conf.OutDir != "" {
		if err := os.MkdirAll(conf.OutDir, 0755); err != nil && !os.IsExist(err) {
			log.Fatal("unable to create outdir: ", err)
		}
		builderCache.SetDestination(conf.OutDir)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for _ = range c {
			builderCache.ClearAll()
			os.Exit(1)
		}
	}()

	r := mux.NewRouter()
	handler := magnaserv{config: conf, builderCache: builderCache}

	log.Println("using Mapnik", mapnik.Version.String)
	if mapnik.Version.Major == 2 {
		handler.mapnikMaker = mapnikBuilder.Maker2
	} else {
		handler.mapnikMaker = mapnikBuilder.Maker3
	}
	v1 := r.PathPrefix("/api/v1").Subrouter()
	v1.HandleFunc("/map", handler.render)
	v1.HandleFunc("/projects", handler.projects)
	v1.Handle("/changes", websocket.Handler(handler.changes))

	r.Handle("/magnacarto/{path:.*}", http.StripPrefix("/magnacarto/", http.FileServer(http.Dir("app"))))

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/magnacarto/", 302)
	})

	for _, fontDir := range conf.Mapnik.FontDirs {
		mapnik.RegisterFonts(fontDir)
	}
	log.Printf("listening on http://%s", *listenAddr)

	log.Fatal(http.ListenAndServe(*listenAddr, handlers.LoggingHandler(os.Stdout, r)))
}
