package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/omniscale/magnacarto/mml"

	"golang.org/x/net/websocket"

	"github.com/omniscale/magnacarto"
	"github.com/omniscale/magnacarto/builder"
	mapnikBuilder "github.com/omniscale/magnacarto/builder/mapnik"
	"github.com/omniscale/magnacarto/builder/mapserver"
	"github.com/omniscale/magnacarto/config"
	"github.com/omniscale/magnacarto/maps"
	mssPkg "github.com/omniscale/magnacarto/mss"
	"github.com/omniscale/magnacarto/render"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type magnaserv struct {
	config            *config.Magnacarto
	builderCache      *builder.Cache
	mapnikMaker       builder.MapMaker
	defaultMaker      builder.MapMaker
	mapnikRenderer    *render.Mapnik
	mapserverRenderer *render.MapServer
}

func (s *magnaserv) styleParams(r *http.Request) (mml string, mss []string) {
	baseDir := s.config.StylesDir
	base := r.FormValue("base")
	if base != "" {
		baseDir = filepath.Join(baseDir, base)
	}

	mml = r.FormValue("mml")
	if mml != "" {
		mml = filepath.Join(baseDir, mml)
	}

	mssList := r.FormValue("mss")
	if mssList != "" {
		for _, f := range strings.Split(mssList, ",") {
			mss = append(mss, filepath.Join(baseDir, f))
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
	switch renderer {
	case "mapnik":
		maker = s.mapnikMaker
	case "mapserver":
		maker = mapserver.Maker
	default:
		maker = s.defaultMaker
		if mapserver.Maker == s.defaultMaker {
			renderer = "mapserver"
		}
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

	w.Header().Add("Content-Type", "image/png")
	if renderer == "mapserver" {
		mapReq.Format = mapReq.Query.Get("FORMAT") // use requested format, not internal mapnik format
		if s.mapserverRenderer == nil {
			err = errors.New("mapserver not initialized")
		}
		err = s.mapserverRenderer.Render(styleFile, w, renderReq(mapReq))
	} else {
		if s.mapnikRenderer == nil {
			err = errors.New("mapnik not initialized")
		} else {
			err = s.mapnikRenderer.Render(styleFile, w, renderReq(mapReq))
		}
	}

	if err != nil {
		s.internalError(w, r, err)
		return
	}
}

func (s *magnaserv) projects(w http.ResponseWriter, r *http.Request) {
	projects, err := findProjects(s.config.StylesDir)
	if err != nil {
		s.internalError(w, r, err)
		return
	}

	sort.Sort(sort.Reverse(byLastChange(projects)))
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

func (s *magnaserv) mml(w http.ResponseWriter, r *http.Request) {
	path := mux.Vars(r)["path"]
	// mux returns safe path (e.g no /-root or ../ tricks)
	path = filepath.FromSlash(path)

	fileName, err := filepath.Abs(filepath.Join(s.config.StylesDir, path+".mml"))
	if err != nil {
		s.internalError(w, r, err)
		return
	}

	if r.Method == "POST" {
		if err := writeCheckedMML(r.Body, fileName); err != nil {
			s.internalError(w, r, err)
			return
		}
		return
	}
	http.ServeFile(w, r, fileName)
}

func (s *magnaserv) mcp(w http.ResponseWriter, r *http.Request) {
	path := mux.Vars(r)["path"]
	// mux returns safe path (e.g no /-root or ../ tricks)
	path = filepath.FromSlash(path)

	mcpFile, err := filepath.Abs(filepath.Join(s.config.StylesDir, path+".mcp"))
	if err != nil {
		s.internalError(w, r, err)
		return
	}

	if r.Method == "POST" {
		if err := writeCheckedJSON(r.Body, mcpFile); err != nil {
			s.internalError(w, r, err)
			return
		}
		return
	}

	// return mcp if exists
	if _, err := os.Stat(mcpFile); err == nil {
		http.ServeFile(w, r, mcpFile)
	} else {
		mmlFile := mcpFile[:len(mcpFile)-3] + "mml"
		// return empty JSON if mml exists
		if _, err := os.Stat(mmlFile); err == nil {
			w.Header().Add("Content-type", "application/json")
			w.Write([]byte("{}\n"))
		} else { // otherwise 404
			http.NotFound(w, r)
		}
	}
}

// writeCheckedMML writes MML from io.ReadCloser to fileName.
// Checks if r is a valid MML before (over)writing file.
func writeCheckedMML(r io.ReadCloser, fileName string) error {
	return writeCheckedFile(r, fileName, func(r io.Reader) error {
		_, err := mml.Parse(r)
		return err
	})
	return nil
}

// writeCheckedMML writes JSON from io.ReadCloser to fileName.
// Checks if r is a valid JSON before (over)writing file.
func writeCheckedJSON(r io.ReadCloser, fileName string) error {
	return writeCheckedFile(r, fileName, func(r io.Reader) error {
		d := json.NewDecoder(r)
		tmp := map[string]interface{}{}
		return d.Decode(&tmp)
	})
	return nil
}

func (s *magnaserv) internalError(w http.ResponseWriter, r *http.Request, err error) {
	log.Print(err)
	w.Header().Set("Content-type", "text/plain")
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("internal error"))
}

type fileChecker func(io.Reader) error

func writeCheckedFile(r io.ReadCloser, fileName string, checker fileChecker) error {
	tmpFile := fileName + ".tmp-" + strconv.FormatInt(int64(rand.Int31()), 16)
	f, err := os.Create(tmpFile)
	if err != nil {
		return err
	}
	defer r.Close()
	defer os.Remove(tmpFile) // make sure temp file always gets removed

	_, err = io.Copy(f, r)
	f.Close()
	if err != nil {
		return err
	}

	f, err = os.Open(tmpFile)
	if err != nil {
		return err
	}
	if err := checker(f); err != nil {
		f.Close()
		return err
	}
	f.Close()
	if err := os.Rename(tmpFile, fileName); err != nil {
		return err
	}
	return nil
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
	result.ScaleFactor = r.ScaleFactor
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
	switch renderer {
	case "mapnik":
		maker = s.mapnikMaker
	case "mapserver":
		maker = mapserver.Maker
	default:
		maker = s.defaultMaker
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
				} else if missingFilesErr, ok := update.Err.(*builder.FilesMissingError); ok {
					msg, err = json.Marshal(struct {
						Error string   `json:"error"`
						Files []string `json:"files"`
					}{"missing files", missingFilesErr.Files})
				} else {
					msg, err = json.Marshal(struct {
						Error string `json:"error"`
					}{update.Err.Error()})
				}
			} else {
				msg, err = json.Marshal(struct {
					UpdatedAt  time.Time `json:"updated_at"`
					UpdatedMML bool      `json:"updated_mml"`
				}{update.Time, update.UpdatedMML})

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

func findAppDir() string {
	// relative to the binary for our own binary builds
	binDir := filepath.Dir(os.Args[0])
	appDir := filepath.Join(binDir, "app")
	if _, err := os.Stat(appDir); err == nil {
		return appDir
	}

	// inside source code for developers (when GOPATH is set)
	if gopath := os.Getenv("GOPATH"); gopath != "" {
		appDir := filepath.Join(gopath, "src", "github.com", "omniscale", "magnacarto", "app")
		if _, err := os.Stat(appDir); err == nil {
			return appDir
		}
	}

	// relative to the working dir
	here, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	appDir = filepath.Join(here, "app")
	if _, err := os.Stat(appDir); err == nil {
		return appDir
	}
	log.Fatal("magnacarto ./app dir not found")
	return ""
}

const DefaultConfigFile = "magnaserv.tml"

func main() {
	if os.Getenv("GOMAXPROCS") == "" {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}

	var listenAddr = flag.String("listen", "localhost:7070", "listen address")
	var configFile = flag.String("config", DefaultConfigFile, "config")
	var builderType = flag.String("builder", "mapnik", "builder type {mapnik,mapserver}")
	var version = flag.Bool("version", false, "print version and exit")

	flag.Parse()

	if *version {
		fmt.Println(magnacarto.Version)
		os.Exit(0)
	}

	conf, err := config.Load(*configFile)
	if *configFile == DefaultConfigFile && os.IsNotExist(err) {
		// ignore error for missing default config
	} else if err != nil {
		log.Fatal(err)
	}

	builderCache := builder.NewCache(conf.Locator)
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
	handler.mapnikMaker = mapnikBuilder.Maker2

	mapnikRenderer, err := render.NewMapnik()
	if err != nil {
		log.Print("Mapnik plugin: ", err)
	} else {
		log.Print("Mapnik plugin available")
		for _, fontDir := range conf.Mapnik.FontDirs {
			mapnikRenderer.RegisterFonts(fontDir)
		}
		if is3, _ := mapnikRenderer.Is3(); is3 {
			handler.mapnikMaker = mapnikBuilder.Maker3
		}
		handler.mapnikRenderer = mapnikRenderer
	}

	mapserverRenderer, err := render.NewMapServer()
	if err != nil {
		log.Print("MapServer plugin: ", err)
	} else {
		log.Print("MapServer plugin available")
		handler.mapserverRenderer = mapserverRenderer
	}

	switch *builderType {
	case "mapnik", "mapnik2", "mapnik3":
		handler.defaultMaker = handler.mapnikMaker
	case "mapserver":
		handler.defaultMaker = mapserver.Maker
	default:
		log.Fatal("unknown -builder ", *builderType)
	}

	v1 := r.PathPrefix("/api/v1").Subrouter()
	v1.HandleFunc("/map", handler.render)
	v1.HandleFunc("/projects/{path:.*?}.mml", handler.mml)
	v1.HandleFunc("/projects/{path:.*?}.mcp", handler.mcp)
	v1.HandleFunc("/projects", handler.projects)
	v1.Handle("/changes", websocket.Handler(handler.changes))

	appDir := findAppDir()
	r.Handle("/magnacarto/{path:.*}", http.StripPrefix("/magnacarto/", http.FileServer(http.Dir(appDir))))

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/magnacarto/", 302)
	})

	log.Printf("listening on http://%s", *listenAddr)

	log.Fatal(http.ListenAndServe(*listenAddr, handlers.LoggingHandler(os.Stdout, r)))
}
