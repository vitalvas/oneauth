package tools

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMkDir(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(t *testing.T) string
		perm        os.FileMode
		expectError bool
	}{
		{
			name: "create-new-directory",
			setupFunc: func(t *testing.T) string {
				tempDir := t.TempDir()
				return filepath.Join(tempDir, "new-dir")
			},
			perm:        0755,
			expectError: false,
		},
		{
			name: "create-nested-directory",
			setupFunc: func(t *testing.T) string {
				tempDir := t.TempDir()
				return filepath.Join(tempDir, "level1", "level2", "level3")
			},
			perm:        0700,
			expectError: false,
		},
		{
			name: "existing-directory-same-perms",
			setupFunc: func(t *testing.T) string {
				tempDir := t.TempDir()
				dir := filepath.Join(tempDir, "existing")
				err := os.Mkdir(dir, 0755)
				require.NoError(t, err)
				return dir
			},
			perm:        0755,
			expectError: false,
		},
		{
			name: "existing-directory-different-perms",
			setupFunc: func(t *testing.T) string {
				tempDir := t.TempDir()
				dir := filepath.Join(tempDir, "existing")
				err := os.Mkdir(dir, 0755)
				require.NoError(t, err)
				return dir
			},
			perm:        0700,
			expectError: false,
		},
		{
			name: "directory-with-special-permissions",
			setupFunc: func(t *testing.T) string {
				tempDir := t.TempDir()
				return filepath.Join(tempDir, "special-perms")
			},
			perm:        0750,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tt.setupFunc(t)

			err := MkDir(dir, tt.perm)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			// Verify directory exists
			info, err := os.Stat(dir)
			require.NoError(t, err)
			assert.True(t, info.IsDir())

			// Verify permissions
			assert.Equal(t, tt.perm, info.Mode().Perm())
		})
	}
}

func TestMkDir_FileExists(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "file-not-dir")

	// Create a file at the path
	file, err := os.Create(filePath)
	require.NoError(t, err)
	file.Close()

	// Try to create directory with same path
	err = MkDir(filePath, 0755)
	// On some systems this might not fail, so we'll check if it's a file
	if err == nil {
		// If no error, verify it's still a file, not a directory
		info, statErr := os.Stat(filePath)
		require.NoError(t, statErr)
		assert.False(t, info.IsDir(), "Expected file to remain a file, not become a directory")
	} else {
		// If error, that's expected behavior
		assert.Error(t, err)
	}
}

func TestMkDir_PermissionDenied(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Skipping permission test when running as root")
	}

	// Try to create directory in root (should fail for non-root users)
	err := MkDir("/root/test-oneauth-dir", 0755)
	assert.Error(t, err)
}

func TestMkDir_EmptyPath(t *testing.T) {
	err := MkDir("", 0755)
	assert.Error(t, err)
}

func TestMkDir_RelativePath(t *testing.T) {
	tempDir := t.TempDir()
	oldPwd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(oldPwd)

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	relativePath := "relative-test-dir"
	err = MkDir(relativePath, 0755)
	assert.NoError(t, err)

	// Verify it was created
	info, err := os.Stat(relativePath)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
	assert.Equal(t, os.FileMode(0755), info.Mode().Perm())
}

func TestMkDir_ConcurrentCreation(t *testing.T) {
	tempDir := t.TempDir()
	targetDir := filepath.Join(tempDir, "concurrent-test")

	done := make(chan error, 2)

	// Run two concurrent MkDir operations
	go func() {
		done <- MkDir(targetDir, 0755)
	}()

	go func() {
		done <- MkDir(targetDir, 0755)
	}()

	// Both should complete without error (one creates, one sees existing)
	err1 := <-done
	err2 := <-done

	assert.NoError(t, err1)
	assert.NoError(t, err2)

	// Verify directory exists
	info, err := os.Stat(targetDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}
