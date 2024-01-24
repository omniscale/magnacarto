// The magnacarto command converts CartoCSS to Mapnik/MapServer styles.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
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
	mmlFile := flag.String("mml", "", "mml file")
	var mssFilenames files

	flag.Var(&mssFilenames, "mss", "mss file")
	confFile := flag.String("config", "", "config")
	sqliteDir := flag.String("sqlite-dir", "", "sqlite directory")
	shapeDir := flag.String("shape-dir", "", "shapefile directory")
	imageDir := flag.String("image-dir", "", "image/marker directory")
	fontDir := flag.String("font-dir", "", "fonts directory")
	dataDir := flag.String("data-dir", "", "data directory for OGR/GDAL files, also fallback for sqlite/shape/image/font-dir")
	dumpRules := flag.Bool("dumprules", false, "print calculated rules to stderr")
	builderType := flag.String("builder", "mapnik3", "builder type {mapnik3,mapserver}")
	outFile := flag.String("out", "", "out file")
	relPaths := flag.Bool("relpaths", false, "use relative paths in output style")
	version := flag.Bool("version", false, "print version and exit")

	cpuprofile := flag.String("cpuprofile", "", "write cpu profile to file")

	msNoMapBlock := flag.Bool("ms-no-map-block", false, "hide MAP block, only output layers/symbols for INCLUDE")

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
		conf.Datasources.SQLiteDirs = filepath.SplitList(*sqliteDir)
	}
	if *fontDir != "" {
		conf.Mapnik.FontDirs = filepath.SplitList(*fontDir)
	}
	if *shapeDir != "" {
		conf.Datasources.ShapefileDirs = filepath.SplitList(*shapeDir)
	}
	if *imageDir != "" {
		conf.Datasources.ImageDirs = filepath.SplitList(*imageDir)
	}
	if *dataDir != "" {
		conf.Datasources.DataDirs = filepath.SplitList(*dataDir)
	}

	locator := conf.Locator()
	if *mmlFile != "" {
		locator.SetBaseDir(filepath.Dir(*mmlFile))
	}
	if *outFile != "" {
		locator.SetOutDir(filepath.Dir(*outFile))
	} else if *mmlFile != "" {
		locator.SetOutDir(filepath.Dir(*mmlFile))
	} else {
		wd, _ := os.Getwd()
		locator.SetOutDir(wd)
	}
	if *relPaths {
		locator.UseRelPaths(*relPaths)
	}

	var m builder.MapWriter

	switch *builderType {
	case "mapserver":
		mm := mapserver.New(locator)
		mm.SetNoMapBlock(*msNoMapBlock)
		m = mm
	case "mapnik3":
		m = mapnik.New(locator)
	default:
		log.Fatal("unknown -builder ", *builderType)
	}

	b := builder.New(m)
	b.SetMML(*mmlFile)
	for _, mss := range mssFilenames {
		b.AddMSS(mss)
	}
	if *dumpRules {
		b.SetDumpRulesDest(os.Stderr)
	}

	if err := b.Build(); err != nil {
		log.Fatal("error building style: ", err)
	}

	if *outFile == "" || *outFile == "-" {
		if err := m.Write(os.Stdout); err != nil {
			log.Fatal("error writing style to stdout: ", err)
		}
	} else {
		if err := m.WriteFiles(*outFile); err != nil {
			log.Fatal("error writing style: ", err)
		}
	}

	if mf := locator.MissingFiles(); mf != nil {
		for _, f := range mf {
			log.Println("File not found:", f)
		}
	}
}
