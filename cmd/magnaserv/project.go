package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/omniscale/magnacarto/mml"
)

type project struct {
	Name         string   `json:"name"`
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

		log.Println(path, mmlFile, projDir, filepath.Join(projDir, "*.mss"))

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

		// make mml/mss relative to path
		mmlFile, err = filepath.Rel(path, mmlFile)
		if err != nil {
			return nil, err
		}
		for i := range mssFiles {
			mssFiles[i], err = filepath.Rel(path, mssFiles[i])
			if err != nil {
				return nil, err
			}
		}

		selectdMSS := make([]string, len(parsedMML.Stylesheets))
		for i, mssFile := range parsedMML.Stylesheets {
			selectdMSS[i] = filepath.Join(projName, mssFile)
		}
		name := parsedMML.Name

		projects = append(projects,
			project{
				Name:         name,
				MML:          mmlFile,
				AvailableMSS: mssFiles,
				SelectedMSS:  selectdMSS,
			})
	}

	return projects, nil
}
