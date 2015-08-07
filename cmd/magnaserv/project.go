package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/omniscale/magnacarto/mml"
)

type byLastChange []project

func (p byLastChange) Len() int           { return len(p) }
func (p byLastChange) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p byLastChange) Less(i, j int) bool { return p[i].LastChange.Before(p[j].LastChange) }

type project struct {
	Name         string    `json:"name"`
	Base         string    `json:"base"`
	MML          string    `json:"mml"`
	MCP          string    `json:"mcp"`
	LastChange   time.Time `json:"last_change"`
	AvailableMSS []string  `json:"available_mss"`
	SelectedMSS  []string  `json:"selected_mss"`
}

func findProjects(path string) ([]project, error) {
	projects := []project{}
	mmls, err := filepath.Glob(filepath.Join(path, "*", "*.mml"))
	if err != nil {
		return nil, err
	}

	for _, mmlFile := range mmls {
		projDir := filepath.Dir(mmlFile)
		projName := filepath.Base(projDir)

		mssFiles, err := filepath.Glob(filepath.Join(projDir, "*.mss"))
		if err != nil {
			return nil, err
		}

		r, err := os.Open(mmlFile)
		if err != nil {
			return nil, err
		}
		parsedMML, err := mml.Parse(r)
		r.Close()
		if err != nil {
			return nil, fmt.Errorf("error parsing %s: %v", mmlFile, err)
		}

		lastChange := lastModTime(append([]string{mmlFile}, mssFiles...)...)

		// remove base dir from mml/mss
		mmlFile = filepath.Base(mmlFile)
		for i := range mssFiles {
			mssFiles[i] = filepath.Base(mssFiles[i])
		}

		name := parsedMML.Name

		projects = append(projects,
			project{
				Name:         name,
				Base:         projName,
				LastChange:   lastChange,
				MML:          mmlFile,
				MCP:          strings.TrimSuffix(mmlFile, filepath.Ext(mmlFile)) + ".mcp",
				AvailableMSS: mssFiles,
				SelectedMSS:  parsedMML.Stylesheets,
			})
	}

	return projects, nil
}

func lastModTime(files ...string) time.Time {
	mod := time.Time{}
	for _, f := range files {
		fi, err := os.Stat(f)
		if err != nil {
			continue
		}
		if fi.ModTime().After(mod) {
			mod = fi.ModTime()
		}
	}
	return mod
}
