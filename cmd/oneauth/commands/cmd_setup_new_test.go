package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vitalvas/oneauth/cmd/oneauth/config"
)

func TestSetupNewCmd(t *testing.T) {
	t.Run("CommandMetadata", func(t *testing.T) {
		assert.Equal(t, "new", setupNewCmd.Name)
		assert.Equal(t, "Setup a new YubiKey", setupNewCmd.Usage)
		assert.NotNil(t, setupNewCmd.Action)
		assert.NotNil(t, setupNewCmd.Before)
	})

	t.Run("Flags", func(t *testing.T) {
		assert.NotEmpty(t, setupNewCmd.Flags)

		flagNames := make(map[string]bool)
		for _, f := range setupNewCmd.Flags {
			for _, name := range f.Names() {
				flagNames[name] = true
			}
		}

		expectedFlags := []string{"confirm", "wait", "serial", "username", "valid-days", "rsa-bits", "ecc-bits", "touch-policy", "pin-policy"}
		for _, name := range expectedFlags {
			assert.True(t, flagNames[name], "expected flag %q to exist", name)
		}
	})
}

func TestWriteConfigFile(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "test-config-*")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		configPath := filepath.Join(tmpDir, "subdir", "config.yaml")
		conf := &config.Config{
			Keyring: config.Keyring{
				Yubikey: config.KeyringYubikey{
					Serial: 12345678,
				},
			},
		}

		err = writeConfigFile(conf, configPath)
		require.NoError(t, err)

		info, err := os.Stat(configPath)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0600), info.Mode().Perm())

		data, err := os.ReadFile(configPath)
		require.NoError(t, err)
		assert.Contains(t, string(data), "serial: 12345678")
	})

	t.Run("CreatesParentDirectories", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "test-config-*")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		configPath := filepath.Join(tmpDir, "a", "b", "c", "config.yaml")
		conf := &config.Config{}

		err = writeConfigFile(conf, configPath)
		require.NoError(t, err)

		_, err = os.Stat(configPath)
		assert.NoError(t, err)
	})

	t.Run("InvalidPath", func(t *testing.T) {
		conf := &config.Config{}
		err := writeConfigFile(conf, "/dev/null/invalid/config.yaml")
		assert.Error(t, err)
	})
}
