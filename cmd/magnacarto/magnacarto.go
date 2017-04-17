// The magnacarto command converts CartoCSS to Mapnik/MapServer styles.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime/pprof"
	"time"

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
	dumpRules := flag.Bool("dumprules", false, "print calculated rules to stderr")
	builderType := flag.String("builder", "mapnik2", "builder type {mapnik2,mapnik3,mapserver}")
	outFile := flag.String("out", "", "out file")
	relPaths := flag.Bool("relpaths", false, "use relative paths in output style")
	version := flag.Bool("version", false, "print version and exit")
	benchmark := flag.Bool("benchmark", false, "print total compile time")

	cpuprofile := flag.String("cpuprofile", "", "write cpu profile to file")

	msNoMapBlock := flag.Bool("ms-no-map-block", false, "hide MAP block, only output layers/symbols for INCLUDE")

	flag.Parse()

	start := time.Now()

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
	case "mapnik2":
		mm := mapnik.New(locator)
		mm.SetMapnik2(true)
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

	if *benchmark {
		b.SetBenchmark(true)
	}

	if err := b.Build(); err != nil {
		log.Fatal("error building style: ", err)
	}

	if !*benchmark {
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
	} else {
		fmt.Println("")
		fmt.Println("TOTAL:", time.Since(start))
	}
}
