package regression

import (
	"path/filepath"
	"testing"
)

func TestCases(t *testing.T) {
	mssFiles, err := filepath.Glob("cases/*/test.mss")
	if err != nil {
		t.Fatal(err)
	}

	for _, mss := range mssFiles {
		caseDir, _ := filepath.Split(mss)
		caseName := filepath.Base(caseDir)
		t.Run(caseName, func(t *testing.T) {
			testIt(t, testCase{Name: caseName})
		})
	}
}
