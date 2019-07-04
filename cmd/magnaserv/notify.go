package main

import (
	"os"
	"path/filepath"
	"time"

	"github.com/omniscale/magnacarto/builder"
	mmlparse "github.com/omniscale/magnacarto/mml"
	"gopkg.in/fsnotify.v1"
)

type Update struct {
	Err        error
	Time       time.Time
	UpdatedMML bool
}

type buildStyleFunc func(mm builder.MapMaker, mml string, mss []string) (string, error)

func notifier(buildStyle buildStyleFunc, mm builder.MapMaker, mml string, mss []string, done <-chan struct{}) (updatec chan Update, errc chan error) {
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
				style, err := buildStyle(mm, mml, mss)
				if evt.Name == mml && len(mss) == 0 {
					// update mms files to watch if mml changed and mss files were not set
					if err := watchMSSFromMML(watcher, mml); err != nil {
						updatec <- Update{Err: err}
					}
				}
				// atomic save of some editors will trigger remove event,
				// which will remove the file from the watcher. add back again
				if evt.Name != "" {
					watcher.Add(evt.Name)
				}
				if err != nil {
					updatec <- Update{Err: err}
				} else {
					fi, _ := os.Stat(style)
					updatec <- Update{Time: fi.ModTime(), UpdatedMML: evt.Name == mml}
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
