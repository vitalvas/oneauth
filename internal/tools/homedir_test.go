package tools

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetHomeDir(t *testing.T) {
	home, err := GetHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	if home == "" {
		t.Fatal("home directory is empty")
	}

	if home != os.Getenv("HOME") {
		t.Fatal("home directory is not equal to $HOME")
	}
}

func TestInHomeDir(t *testing.T) {
	home, err := GetHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	path, err := InHomeDir("test")
	if err != nil {
		t.Fatal(err)
	}

	if path != filepath.Join(home, "test") {
		t.Fatal("path is not equal to home directory + /test")
	}
}
