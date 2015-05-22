package builder

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestIsStale(t *testing.T) {
	dir, err := ioutil.TempDir("", "magnacarto_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	s := style{
		mml: filepath.Join(dir, "foo.mml"),
		mss: []string{filepath.Join(dir, "foo.mss")},
	}

	f, err := os.Create(s.mml)
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	f, err = os.Create(s.mss[0])
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	// stale without s.file
	if stale, err := s.isStale(); !stale || err != nil {
		t.Fatal(stale, err)
	}

	// create fresh style.xml
	s.file = filepath.Join(dir, "style.xml")
	f, err = os.Create(s.file)
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	if stale, err := s.isStale(); stale || err != nil {
		t.Fatal(stale, err)
	}

	// touch mml file
	future := time.Now().Add(time.Minute)
	if err := os.Chtimes(s.mml, future, future); err != nil {
		t.Fatal(err)
	}
	if stale, err := s.isStale(); !stale || err != nil {
		t.Fatal(stale, err)
	}

	// not stale after update of s.file
	if err := os.Chtimes(s.file, future, future); err != nil {
		t.Fatal(err)
	}
	if stale, err := s.isStale(); stale || err != nil {
		t.Fatal(stale, err)
	}

	// touch mss file
	future = time.Now().Add(time.Minute * 2)
	if err := os.Chtimes(s.mss[0], future, future); err != nil {
		t.Fatal(err)
	}
	if stale, err := s.isStale(); !stale || err != nil {
		t.Fatal(stale, err)
	}

}
