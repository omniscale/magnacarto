//go:generate bash mk_test_funcs.sh

package regression

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"

	"github.com/omniscale/magnacarto/builder"
	"github.com/omniscale/magnacarto/builder/mapnik"
	"github.com/omniscale/magnacarto/builder/mapserver"
	"github.com/omniscale/magnacarto/config"
	"github.com/omniscale/magnacarto/render"

	"github.com/BurntSushi/toml"

	"testing"
)

type testCase struct {
	Name             string
	Zoom             int
	Long             float64
	Lat              float64
	Width            int
	Height           int
	CartoTest        bool
	CartoCompare     bool
	CartoFuzz        float64
	CartoPxDiff      int64
	MapServerTest    bool
	MapServerCompare bool
	MapServerFuzz    float64
	MapServerPxDiff  int64
}

var mapnikRenderer *render.Mapnik
var mapserverRenderer *render.MapServer

func init() {
	mapnikRenderer, _ = render.NewMapnik()
	if mapnikRenderer != nil {
		wd, _ := os.Getwd()
		mapnikRenderer.RegisterFonts(wd)
	}
	mapserverRenderer, _ = render.NewMapServer()
}

var cmdsChecked bool

func checkTestCmds(t *testing.T) {
	if cmdsChecked {
		return
	}
	cmdsFound := true
	_, err := exec.LookPath("carto")
	if err != nil {
		t.Error("carto command not found; make sure carto is installed (with npm) and in your PATH")
		cmdsFound = false
	}

	_, err = exec.LookPath("compare")
	if err != nil {
		t.Error("compare command not found; make sure Image-Magick is installed and in your PATH")
		cmdsFound = false
	}

	_, err = exec.LookPath("mapserv")
	if err != nil {
		t.Error("mapserv command not found; make sure mapserv is installed and in your PATH")
		cmdsFound = false
	} else {
		out, err := exec.Command("mapserv", "-v").Output()
		if err != nil {
			t.Error("unable to get mapserv version:", err)
		}
		versionMatch := regexp.MustCompile(`version (\d)`).FindStringSubmatch(string(out))
		if len(versionMatch) < 2 {
			t.Error("unable to parse version:", string(out))
		} else {
			majVersion, _ := strconv.ParseInt(versionMatch[1], 10, 32)
			if majVersion < 7 {
				t.Error("mapserver >= 7 required", string(out))
			}
		}
	}

	if !cmdsFound {
		t.Fatal("not all commands found for regression tests.")
	}

	cmdsChecked = true
}

func (t *testCase) load() error {
	if _, err := toml.DecodeFile("config.tml", &t); err != nil {
		return err
	}
	if _, err := toml.DecodeFile(filepath.Join("cases", t.Name, "config.tml"), &t); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func testIt(t *testing.T, c testCase) {
	if testing.Short() {
		t.Skip("skipping regression test in short mode")
	}
	checkTestCmds(t)
	t.Parallel()

	if err := c.load(); err != nil {
		t.Fatal(err)
	}

	fmt.Printf("%#v\n", c)
	prepare(t, c)
	buildMagnacarto(t, c, false)
	renderMapnik(t, c, "style-mapnik")

	if c.CartoTest {
		buildCarto(t, c)
		renderMapnik(t, c, "style-carto")
	}
	if c.MapServerTest {
		buildMagnacarto(t, c, true)
		renderMapserver(t, c)
	}
	compare(t, c)
}

// prepare copies all files to the build directory
func prepare(t *testing.T, c testCase) {
	caseBuildDir := filepath.Join("build", c.Name)
	if err := os.MkdirAll(caseBuildDir, 0755); err != nil && !os.IsExist(err) {
		t.Fatal(err)
	}
	srcFiles, err := filepath.Glob(filepath.Join("cases", c.Name, "*"))
	if err != nil {
		t.Fatal(err)
	}
	for _, src := range srcFiles {
		if err := cpFile(filepath.Join(caseBuildDir, filepath.Base(src)), src); err != nil {
			t.Fatal(err)
		}
	}
}

// buildCarto builds the mapnik XML with carto
func buildCarto(t *testing.T, c testCase) {
	caseBuildDir := filepath.Join("build", c.Name)

	cmd := exec.Command("carto", "test.mml")
	cmd.Dir = caseBuildDir

	dst, err := os.Create(filepath.Join(caseBuildDir, "style-carto.xml"))
	if err != nil {
		t.Fatal(err)
	}
	defer dst.Close()
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	defer stdout.Close()
	stderr, err := cmd.StderrPipe()
	if err != nil {
		t.Fatal(err)
	}
	defer stderr.Close()
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	if _, err := io.Copy(dst, stdout); err != nil {
		t.Fatal("err copying stdout:", err)
	}

	errContent, _ := ioutil.ReadAll(stderr)

	if err := cmd.Wait(); err != nil {
		// carto outputs errors to stdout, read content of file to display err
		dst.Close()
		if dst, err := os.Open(filepath.Join(caseBuildDir, "style-carto.xml")); err == nil {
			if err != nil {
				t.Fatal(err)
			}
			content, _ := ioutil.ReadAll(dst)
			dst.Close()
			t.Log(string(content))
			t.Log(string(errContent))
		}
		t.Fatalf("error calling carto for %s: %s", c.Name, err)
	}
}

func wgsToMerc(long, lat float64) (x, y float64) {
	x = long * 6378137 * math.Pi / 180.0
	y = math.Log(math.Tan((90.0+lat)*math.Pi/360.0)) / math.Pi * 6378137 * math.Pi
	return x, y
}

func extent(c testCase, fixOffset bool) [4]float64 {
	x, y := wgsToMerc(c.Long, c.Lat)
	equatorWidth := 6378137 * math.Pi * 2
	res := (equatorWidth / 256.0) / math.Pow(2, float64(c.Zoom))
	width, height := c.Width, c.Height
	if fixOffset {
		// fix for center/outer pixel differences
		return [4]float64{x - float64(width)/2*res + 0.5*res, y - float64(height)/2*res - 0.5*res, x + float64(width)/2*res + 0.5*res, y + float64(height)/2*res - 0.5*res}
	}
	return [4]float64{x - float64(width)/2*res, y - float64(height)/2*res, x + float64(width)/2*res, y + float64(height)/2*res}
}

func renderMapnik(t *testing.T, c testCase, name string) {
	caseBuildDir := filepath.Join("build", c.Name)

	mapReq := render.Request{}
	mapReq.EPSGCode = 3857
	mapReq.BBOX = extent(c, true)
	mapReq.Width = c.Width
	mapReq.Height = c.Height
	mapReq.Format = "png24"

	if mapnikRenderer == nil {
		t.Skip("mapnik not initialized")
	}

	f, err := os.Create(filepath.Join(caseBuildDir, "render-"+strings.TrimPrefix(name, "style-")+".png"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	err = mapnikRenderer.Render(filepath.Join(caseBuildDir, name+".xml"), f, mapReq)
	if err != nil {
		t.Fatal(err)
	}
}

func renderMapserver(t *testing.T, c testCase) {
	caseBuildDir := filepath.Join("build", c.Name)

	mapReq := render.Request{}
	mapReq.EPSGCode = 3857
	mapReq.BBOX = extent(c, false)
	mapReq.Width = c.Width
	mapReq.Height = c.Height
	mapReq.Format = "image/png"

	if mapserverRenderer == nil {
		t.Skip("mapserver not initialized")
	}

	f, err := os.Create(filepath.Join(caseBuildDir, "render-mapserver.png"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	_, err = mapserverRenderer.Render(filepath.Join(caseBuildDir, "style-mapserver.map"), f, mapReq)
	if err != nil {
		t.Fatal(err)
	}
}

// buildMagnacarto builds the mapnik XML or mapserver MAP (when mapfile is true) with magnacarto
func buildMagnacarto(t *testing.T, c testCase, mapfile bool) {
	caseBuildDir := filepath.Join("build", c.Name)

	suffix := "mapnik.xml"
	if mapfile {
		suffix = "mapserver.map"
	}

	conf := config.Magnacarto{BaseDir: caseBuildDir}
	here, _ := os.Getwd()
	conf.Mapnik.FontDirs = []string{here}
	locator := conf.Locator()

	var m builder.MapWriter
	if mapfile {
		m = mapserver.New(locator)
	} else {
		m = mapnik.New(locator)
	}

	b := builder.New(m)
	b.SetMML(filepath.Join(caseBuildDir, "test.mml"))

	if err := b.Build(); err != nil {
		log.Fatal("error building map: ", err)
	}

	if err := m.WriteFiles(filepath.Join(caseBuildDir, "style-"+suffix)); err != nil {
		log.Fatal("error writing map: ", err)
	}
}

func compare(t *testing.T, c testCase) {
	dir := filepath.Join("build", c.Name)

	if c.CartoTest && c.CartoCompare {
		compareImg(t, dir, "render-carto.png", "render-mapnik.png", c.CartoFuzz, c.CartoPxDiff)
	}
	if c.MapServerTest && c.MapServerCompare {
		compareImg(t, dir, "render-mapnik.png", "render-mapserver.png", c.MapServerFuzz, c.MapServerPxDiff)
	}
}

func compareImg(t *testing.T, dir, fileA, fileB string, fuzz float64, expected int64) {
	strip := func(f string) string {
		return strings.TrimPrefix(strings.TrimSuffix(f, ".png"), "render-")
	}

	fileDiff := "diff-" + strip(fileA) + "-" + strip(fileB) + ".png"
	cmd := exec.Command(
		"compare", "-metric", "AE", "-fuzz", fmt.Sprintf("%.2f%%", fuzz), fileA, fileB, fileDiff,
	)
	cmd.Dir = dir

	out, err := cmd.CombinedOutput()
	if err != nil {
		// ignore exit status 1 if images differ
		if exit, ok := err.(*exec.ExitError); ok && exit.Sys().(syscall.WaitStatus).ExitStatus() == 2 {
			t.Logf("%#v", err)
			t.Log(string(out))
			t.Fatalf("error calling compare for %s (%s/%s): %s", dir, fileA, fileB, err)
		}
	}

	px := regexp.MustCompile(`\((\d+)\)`).FindStringSubmatch(string(out))
	diff, err := strconv.ParseInt(px[1], 10, 32)
	if err != nil {
		t.Fatal("found no diff output: ", px)
	}
	if diff > expected {
		t.Errorf("diff for %s and %s is too large (%d>%d), see %s", filepath.Join(dir, fileA), filepath.Join(dir, fileB), diff, expected, filepath.Join(dir, fileDiff))
	}
}

func cpFile(dst, src string) error {
	s, err := os.Open(src)
	if err != nil {
		return err
	}
	defer s.Close()
	d, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(d, s); err != nil {
		d.Close()
		return err
	}
	return d.Close()
}
