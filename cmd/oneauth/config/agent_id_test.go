package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgentID_LoadOrCreate(t *testing.T) {
	t.Run("CreatesNewID", func(t *testing.T) {
		tmpHome, err := os.MkdirTemp("", "test-home-*")
		require.NoError(t, err)
		defer os.RemoveAll(tmpHome)

		oneauthDir := filepath.Join(tmpHome, ".oneauth")
		err = os.MkdirAll(oneauthDir, 0755)
		require.NoError(t, err)

		originalHome := os.Getenv("HOME")
		os.Setenv("HOME", tmpHome)
		defer os.Setenv("HOME", originalHome)

		id, err := LoadOrCreateAgentID()
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, id)

		// Verify file was created
		agentIDPath := filepath.Join(oneauthDir, "agent_id")
		data, err := os.ReadFile(agentIDPath)
		require.NoError(t, err)
		assert.Equal(t, id.String(), string(data))
	})

	t.Run("LoadsExistingID", func(t *testing.T) {
		tmpHome, err := os.MkdirTemp("", "test-home-*")
		require.NoError(t, err)
		defer os.RemoveAll(tmpHome)

		oneauthDir := filepath.Join(tmpHome, ".oneauth")
		err = os.MkdirAll(oneauthDir, 0755)
		require.NoError(t, err)

		originalHome := os.Getenv("HOME")
		os.Setenv("HOME", tmpHome)
		defer os.Setenv("HOME", originalHome)

		existingID := uuid.MustParse("01234567-89ab-cdef-0123-456789abcdef")
		agentIDPath := filepath.Join(oneauthDir, "agent_id")
		err = os.WriteFile(agentIDPath, []byte(existingID.String()), 0600)
		require.NoError(t, err)

		id, err := LoadOrCreateAgentID()
		require.NoError(t, err)
		assert.Equal(t, existingID, id)
	})

	t.Run("RegeneratesOnInvalidContent", func(t *testing.T) {
		tmpHome, err := os.MkdirTemp("", "test-home-*")
		require.NoError(t, err)
		defer os.RemoveAll(tmpHome)

		oneauthDir := filepath.Join(tmpHome, ".oneauth")
		err = os.MkdirAll(oneauthDir, 0755)
		require.NoError(t, err)

		originalHome := os.Getenv("HOME")
		os.Setenv("HOME", tmpHome)
		defer os.Setenv("HOME", originalHome)

		agentIDPath := filepath.Join(oneauthDir, "agent_id")
		err = os.WriteFile(agentIDPath, []byte("not-a-valid-uuid"), 0600)
		require.NoError(t, err)

		id, err := LoadOrCreateAgentID()
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, id)

		// Verify the file was overwritten with valid UUID
		data, err := os.ReadFile(agentIDPath)
		require.NoError(t, err)
		parsedID, err := uuid.Parse(string(data))
		require.NoError(t, err)
		assert.Equal(t, id, parsedID)
	})

	t.Run("ErrorOnUnwritablePath", func(t *testing.T) {
		tmpHome, err := os.MkdirTemp("", "test-home-*")
		require.NoError(t, err)
		defer os.RemoveAll(tmpHome)

		// Do NOT create .oneauth dir, and make parent read-only
		originalHome := os.Getenv("HOME")
		os.Setenv("HOME", filepath.Join(tmpHome, "nonexistent"))
		defer os.Setenv("HOME", originalHome)

		_, err = LoadOrCreateAgentID()
		assert.Error(t, err)
	})
}
