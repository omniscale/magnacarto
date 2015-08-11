package builder

import (
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gopkg.in/fsnotify.v1"

	"github.com/omniscale/magnacarto/config"
	mmlparse "github.com/omniscale/magnacarto/mml"
)

// MapMaker creates new MapWriters.
type MapMaker interface {
	New(config.Locator) MapWriter
	// Type returns a unique string for this MapMaker, used for caching generated styles.
	Type() string
	FileSuffix() string
}

type locatorCreator func() config.Locator

type style struct {
	mapMaker   MapMaker
	mml        string
	mss        []string
	file       string
	lastUpdate time.Time
}

func styleHash(mapType string, mml string, mss []string) uint32 {
	f := fnv.New32()
	f.Write([]byte(mapType))
	f.Write([]byte(mml))
	for i := range mss {
		f.Write([]byte(mss[i]))
	}
	return f.Sum32()
}

func isNewer(file string, timestamp time.Time) bool {
	info, err := os.Stat(file)
	if err != nil {
		return true
	}
	return info.ModTime().After(timestamp)
}

func (s *style) isStale() (bool, error) {
	if s.file == "" {
		return true, nil
	}

	info, err := os.Stat(s.file)
	if err != nil {
		return true, err
	}
	timestamp := info.ModTime()

	if isNewer(s.mml, timestamp) {
		return true, nil
	}
	for _, mss := range s.mss {
		if isNewer(mss, timestamp) {
			return true, nil
		}
	}
	return false, nil
}

const stylePrefix = "magnacarto-style-"

// Cache builds styles and caches the results.
// It automatically detects changes to the MSS and MML files and rebuilds
// styles if requested again.
type Cache struct {
	mu         sync.Mutex
	newLocator locatorCreator
	styles     map[uint32]*style
	destDir    string
}

func NewCache(newLocator locatorCreator) *Cache {
	return &Cache{
		newLocator: newLocator,
		styles:     make(map[uint32]*style),
	}
}

func (c *Cache) SetDestination(dest string) {
	c.destDir = dest
}

// ClearAll removes all cached styles.
// Needs to be called before shutdown to prevent leaking temp files when used _without_ SetDestination.
// Will remove all cached styles from cache dir when used _with_ SetDestination.
func (c *Cache) ClearAll() {
	c.ClearTill(time.Now())
}

// ClearTill removes all cached styles that are older then till.
func (c *Cache) ClearTill(till time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.destDir != "" {
		// all in same dir. also remove files with different suffixes (e.g. foo.map.font.lst)
		files, err := filepath.Glob(filepath.Join(c.destDir, stylePrefix+"*"))
		if err != nil {
			log.Println("cleanup error: ", err)
			return
		}
		for _, f := range files {
			if fi, err := os.Stat(f); err == nil && fi.ModTime().Before(till) {
				if err := os.Remove(f); err != nil {
					log.Println("cleanup error: ", err)
				}
			}
		}
	} else {
		// with empty destDir, each style is in its own temp dir.
		for _, style := range c.styles {
			if fi, err := os.Stat(style.file); err == nil && fi.ModTime().Before(till) {
				if err := os.RemoveAll(filepath.Dir(style.file)); err != nil {
					log.Println(err)
				}
			}
		}
	}

	for hash, _ := range c.styles {
		delete(c.styles, hash)
	}
}

type Update struct {
	Err        error
	Time       time.Time
	UpdatedMML bool
}

func (c *Cache) Notify(mm MapMaker, mml string, mss []string, done <-chan struct{}) (updatec chan Update, errc chan error) {
	updatec = make(chan Update, 1)
	errc = make(chan error, 1)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		errc <- err
		return
	}

	if err := watcher.Add(mml); err != nil {
		updatec <- Update{Err: err}
		return
	}
	for _, mss := range mss {
		if err := watcher.Add(mss); err != nil {
			updatec <- Update{Err: err}
			return
		}
	}

	if len(mss) == 0 {
		// add mss files from mml to watcher, keep mss empty so we know that
		// we style all mss files
		if err := watchMSSFromMML(watcher, mml); err != nil {
			updatec <- Update{Err: err}
			return
		}
	}

	go func() {
		// dummy event to send initial change message to client
		watcher.Events <- fsnotify.Event{}
	}()
	go func() {
		for {
			select {
			case evt := <-watcher.Events:
				if evt.Op&fsnotify.Remove == fsnotify.Remove {
					// atomic save of some editors will trigger remove event,
					// which will remove the file from the watcher. add back again
					watcher.Add(evt.Name)
				}
				if evt.Name == mml && len(mss) == 0 {
					// update mms files to watch if mml changed and mss files were not set
					if err := watchMSSFromMML(watcher, mml); err != nil {
						updatec <- Update{Err: err}
					}
				}
				style, err := c.style(mm, mml, mss)

				if err != nil {
					updatec <- Update{Err: err}
				} else {
					updatec <- Update{Time: style.lastUpdate, UpdatedMML: evt.Name == mml}
				}
			case err := <-watcher.Errors:
				errc <- err
			case <-done:
				watcher.Close()
				return
			}
		}
	}()
	return updatec, errc
}

// StyleFile returns the filename of the build result. (Re)builds style if required.
func (c *Cache) StyleFile(mm MapMaker, mml string, mss []string) (string, error) {
	style, err := c.style(mm, mml, mss)
	if err != nil {
		return "", err
	}
	return style.file, nil
}

func (c *Cache) style(mm MapMaker, mml string, mss []string) (*style, error) {
	hash := styleHash(mm.Type(), mml, mss)
	c.mu.Lock()
	defer c.mu.Unlock()
	if s, ok := c.styles[hash]; ok {
		stale, err := s.isStale()
		if err != nil {
			return nil, err
		}
		if stale {
			if len(mss) == 0 {
				var err error
				// refresh mss files
				s.mss, err = mssFilesFromMML(mml)
				if err != nil {
					return nil, err
				}
			}
			if err := c.build(s); err != nil {
				return nil, err
			}
		}
		return s, nil
	} else {
		if len(mss) == 0 {
			var err error
			mss, err = mssFilesFromMML(mml)
			if err != nil {
				return nil, err
			}
		}
		s = &style{
			mapMaker: mm,
			mml:      mml,
			mss:      mss,
		}
		if err := c.build(s); err != nil {
			return nil, err
		}
		c.styles[hash] = s
		return s, nil
	}
}

type FilesMissingError struct {
	Files []string
}

func (e *FilesMissingError) Error() string {
	return fmt.Sprintf("missing files: %v", e.Files)
}

func (c *Cache) build(style *style) error {
	l := c.newLocator()
	l.SetBaseDir(filepath.Dir(style.mml))
	l.SetOutDir(c.destDir)
	l.UseRelPaths(false)

	m := style.mapMaker.New(l)
	builder := New(m)

	builder.SetMML(style.mml)
	for _, mss := range style.mss {
		builder.AddMSS(mss)
	}

	if err := builder.Build(); err != nil {
		return err
	}

	if files := l.MissingFiles(); len(files) > 0 {
		return &FilesMissingError{files}
	}

	var styleFile string
	if c.destDir != "" {
		hash := styleHash(style.mapMaker.Type(), style.mml, style.mss)
		styleFile = filepath.Join(c.destDir, fmt.Sprintf("magnacarto-style-%d%s", hash, style.mapMaker.FileSuffix()))
		if err := m.WriteFiles(styleFile); err != nil {
			return err
		}
	} else {
		tmp, err := ioutil.TempDir("", "magnacarto-style")
		if err != nil {
			return err
		}
		styleFile = filepath.Join(tmp, "style"+style.mapMaker.FileSuffix())
		if err := m.WriteFiles(styleFile); err != nil {
			os.RemoveAll(tmp)
			return err
		}
	}
	log.Printf("rebuild style %s as %s with %v\n", style.mml, styleFile, style.mss)
	style.lastUpdate = time.Now()
	style.file = styleFile
	return nil
}

func mssFilesFromMML(mmlFile string) ([]string, error) {
	r, err := os.Open(mmlFile)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	mml, err := mmlparse.Parse(r)
	if err != nil {
		return nil, err
	}
	mssFiles := []string{}
	for _, s := range mml.Stylesheets {
		mssFiles = append(mssFiles, filepath.Join(filepath.Dir(mmlFile), s))
	}
	return mssFiles, nil
}

func watchMSSFromMML(watcher *fsnotify.Watcher, mmlFile string) error {
	mssFiles, err := mssFilesFromMML(mmlFile)
	if err != nil {
		return err
	}
	for _, mssFile := range mssFiles {
		if err := watcher.Add(mssFile); err != nil {
			return err
		}
	}
	return nil
}
