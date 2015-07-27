// The magnacarto command converts CartoCSS to Mapnik/MapServer styles.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/pprof"

	"github.com/omniscale/magnacarto"
	"github.com/omniscale/magnacarto/builder"
	"github.com/omniscale/magnacarto/builder/mapnik"
	"github.com/omniscale/magnacarto/builder/mapserver"
	"github.com/omniscale/magnacarto/config"
)

type files []string

func (f *files) String() string {
	return fmt.Sprintf("%v", *f)
}

func (f *files) Set(value string) error {
	*f = append(*f, value)
	return nil
}

func main() {
	mmlFilename := flag.String("mml", "", "mml file")
	var mssFilenames files

	flag.Var(&mssFilenames, "mss", "mss file")
	confFile := flag.String("config", "", "config")
	sqliteDir := flag.String("sqlite-dir", "", "sqlite directory")
	shapeDir := flag.String("shape-dir", "", "shapefile directory")
	imageDir := flag.String("image-dir", "", "image/marker directory")
	fontDir := flag.String("font-dir", "", "fonts directory")
	dumpRules := flag.Bool("dumprules", false, "print calculated rules to stderr")
	builderType := flag.String("builder", "mapnik2", "builder type {mapnik2,mapnik3,mapserver}")
	outFile := flag.String("out", "", "out file")
	version := flag.Bool("version", false, "print version and exit")
	noCheckFiles := flag.Bool("no-check-files", false, "do not check if images/shps/etc exists")

	cpuprofile := flag.String("cpuprofile", "", "write cpu profile to file")

	flag.Parse()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	if *version {
		fmt.Println(magnacarto.Version)
		os.Exit(0)
	}

	conf := config.Magnacarto{}
	if *confFile != "" {
		if err := conf.Load(*confFile); err != nil {
			log.Fatal(err)
		}
	}

	// overwrite config with command line args
	if *sqliteDir != "" {
		conf.Datasources.SQLiteDirs = []string{*sqliteDir}
	}
	if *fontDir != "" {
		conf.Mapnik.FontDirs = []string{*fontDir}
	}
	if *shapeDir != "" {
		conf.Datasources.ShapefileDirs = []string{*shapeDir}
	}
	if *imageDir != "" {
		conf.Datasources.ImageDirs = []string{*imageDir}
	}

	if *noCheckFiles {
		conf.Datasources.NoCheckFiles = true
	}

	locator := conf.Locator()

	var m builder.MapWriter

	switch *builderType {
	case "mapserver":
		m = mapserver.New(locator)
	case "mapnik2":
		m = mapnik.New(locator)
		m.(*mapnik.Map).SetMapnik2(true)
	case "mapnik3":
		m = mapnik.New(locator)
	default:
		log.Fatal("unknown -builder ", *builderType)
	}

	b := builder.New(m)
	b.SetMML(*mmlFilename)
	for _, mss := range mssFilenames {
		b.AddMSS(mss)
	}
	if *dumpRules {
		b.SetDumpRulesDest(os.Stderr)
	}

	if err := b.Build(); err != nil {
		log.Fatal("error building map: ", err)
	}

	if *outFile == "" || *outFile == "-" {
		if err := m.Write(os.Stdout); err != nil {
			log.Fatal("error writing map to stdout: ", err)
		}
	} else {
		if err := m.WriteFiles(*outFile); err != nil {
			log.Fatal("error writing map: ", err)
		}
	}
}
