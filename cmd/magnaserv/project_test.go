package main

import "testing"

func TestFindProjects(t *testing.T) {
	projects, err := findProjects("../../regression/cases")
	if err != nil {
		t.Fatal(err)
	}
	if len(projects) == 0 {
		t.Error(projects)
	}

	project := projects[0]
	if project.Name != "Magnacarto Test" ||
		project.MML != "010-linestrings-default/test.mml" ||
		len(project.SelectedMSS) != 1 ||
		len(project.AvailableMSS) != 1 ||
		project.SelectedMSS[0] != "010-linestrings-default/test.mss" ||
		project.AvailableMSS[0] != "010-linestrings-default/test.mss" {
		t.Error("unexpected project", project)
	}
}
