package tools

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMkDir(t *testing.T) {
	t.Run("DirectoryDoesNotExist", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "newdir")
		perm := os.FileMode(0755)

		err := MkDir(dir, perm)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		stat, err := os.Stat(dir)
		if err != nil {
			t.Fatalf("expected directory to be created, got error %v", err)
		}

		if stat.Mode().Perm() != perm {
			t.Errorf("expected permissions %v, got %v", perm, stat.Mode().Perm())
		}
	})

	t.Run("DirectoryExistsWithDifferentPermissions", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "existingdir")
		initialPerm := os.FileMode(0755)
		newPerm := os.FileMode(0700)

		err := os.Mkdir(dir, initialPerm)
		if err != nil {
			t.Fatalf("setup: expected no error, got %v", err)
		}

		err = MkDir(dir, newPerm)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		stat, err := os.Stat(dir)
		if err != nil {
			t.Fatalf("expected directory to exist, got error %v", err)
		}

		if stat.Mode().Perm() != newPerm {
			t.Errorf("expected permissions %v, got %v", newPerm, stat.Mode().Perm())
		}
	})

	t.Run("DirectoryExistsWithCorrectPermissions", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "correctdir")
		perm := os.FileMode(0755)

		err := os.Mkdir(dir, perm)
		if err != nil {
			t.Fatalf("setup: expected no error, got %v", err)
		}

		err = MkDir(dir, perm)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		stat, err := os.Stat(dir)
		if err != nil {
			t.Fatalf("expected directory to exist, got error %v", err)
		}

		if stat.Mode().Perm() != perm {
			t.Errorf("expected permissions %v, got %v", perm, stat.Mode().Perm())
		}
	})
}
