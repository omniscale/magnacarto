package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/omniscale/magnacarto/mml"
)

type project struct {
	Name         string   `json:"name"`
	Base         string   `json:"base"`
	MML          string   `json:"mml"`
	AvailableMSS []string `json:"available_mss"`
	SelectedMSS  []string `json:"selected_mss"`
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
				MML:          mmlFile,
				AvailableMSS: mssFiles,
				SelectedMSS:  parsedMML.Stylesheets,
			})
	}

	return projects, nil
}
