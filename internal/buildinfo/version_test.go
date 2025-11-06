package buildinfo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormattedVersion(t *testing.T) {
	// Save original values
	originalVersion := Version
	originalCommit := Commit

	// Restore original values after test
	defer func() {
		Version = originalVersion
		Commit = originalCommit
	}()

	t.Run("VersionOnly", func(t *testing.T) {
		Version = "1.0.0"
		Commit = ""
		
		result := FormattedVersion()
		assert.Equal(t, "1.0.0", result)
	})

	t.Run("VersionWithCommit", func(t *testing.T) {
		Version = "1.0.0"
		Commit = "abcdef1234567890"
		
		result := FormattedVersion()
		assert.Equal(t, "1.0.0-abcdef12", result)
	})

	t.Run("VersionWithShortCommit", func(t *testing.T) {
		Version = "1.0.0"
		Commit = "abc123"
		
		result := FormattedVersion()
		assert.Equal(t, "1.0.0", result) // Should not include short commit
	})

	t.Run("VersionWithExactly8CharCommit", func(t *testing.T) {
		Version = "1.0.0"
		Commit = "abcdef12"
		
		result := FormattedVersion()
		assert.Equal(t, "1.0.0-abcdef12", result)
	})

	t.Run("EmptyVersion", func(t *testing.T) {
		Version = ""
		Commit = "abcdef1234567890"
		
		result := FormattedVersion()
		assert.Equal(t, "-abcdef12", result)
	})

	t.Run("BothEmpty", func(t *testing.T) {
		Version = ""
		Commit = ""
		
		result := FormattedVersion()
		assert.Equal(t, "", result)
	})
}