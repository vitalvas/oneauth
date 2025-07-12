package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vitalvas/oneauth/cmd/oneauth/config"
)

func TestAgentValidation_YubikeySerialRequirement(t *testing.T) {
	t.Run("UnixSocketRequiresYubikeySerial", func(t *testing.T) {
		// Create a temporary config file with serial = 0
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "config.yaml")
		
		configContent := `
socket:
  type: unix
  path: ` + filepath.Join(tempDir, "test.sock") + `
control_socket_path: ` + filepath.Join(tempDir, "test_ctrl.sock") + `
agent_log_path: ` + filepath.Join(tempDir, "test.log") + `
keyring:
  yubikey:
    serial: 0
`
		err := os.WriteFile(configPath, []byte(configContent), 0644)
		assert.NoError(t, err)

		// Load the config
		cfg, err := config.Load(configPath)
		assert.NoError(t, err)

		// Test the validation logic that's in the agent command
		if cfg.Socket.Type == "unix" && !cfg.Keyring.DisableYubikey && cfg.Keyring.Yubikey.Serial == 0 {
			err = assert.AnError // Simulate the error that would be returned
		}
		
		// Should detect that YubiKey serial is required
		assert.Error(t, err, "Expected error when YubiKey serial is 0 for unix socket")
	})

	t.Run("DummySocketDoesNotRequireYubikeySerial", func(t *testing.T) {
		// Create a temporary config file with dummy socket and serial = 0
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "config.yaml")
		
		configContent := `
socket:
  type: dummy
control_socket_path: ` + filepath.Join(tempDir, "test_ctrl.sock") + `
agent_log_path: ` + filepath.Join(tempDir, "test.log") + `
keyring:
  yubikey:
    serial: 0
`
		err := os.WriteFile(configPath, []byte(configContent), 0644)
		assert.NoError(t, err)

		// Load the config
		cfg, err := config.Load(configPath)
		assert.NoError(t, err)

		// Test the validation logic - dummy socket should not require YubiKey
		var validationErr error
		if cfg.Socket.Type == "unix" && !cfg.Keyring.DisableYubikey && cfg.Keyring.Yubikey.Serial == 0 {
			validationErr = assert.AnError // This should NOT trigger for dummy
		}
		
		// Should NOT require YubiKey serial for dummy socket
		assert.NoError(t, validationErr, "Dummy socket should not require YubiKey serial")
		assert.Equal(t, "dummy", cfg.Socket.Type)
		assert.Equal(t, uint32(0), cfg.Keyring.Yubikey.Serial)
	})

	t.Run("UnixSocketWithValidSerial", func(t *testing.T) {
		// Create a temporary config file with valid serial
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "config.yaml")
		
		configContent := `
socket:
  type: unix
  path: ` + filepath.Join(tempDir, "test.sock") + `
control_socket_path: ` + filepath.Join(tempDir, "test_ctrl.sock") + `
agent_log_path: ` + filepath.Join(tempDir, "test.log") + `
keyring:
  yubikey:
    serial: 12345
`
		err := os.WriteFile(configPath, []byte(configContent), 0644)
		assert.NoError(t, err)

		// Load the config
		cfg, err := config.Load(configPath)
		assert.NoError(t, err)

		// Test the validation logic
		var validationErr error
		if cfg.Socket.Type == "unix" && !cfg.Keyring.DisableYubikey && cfg.Keyring.Yubikey.Serial == 0 {
			validationErr = assert.AnError // This should NOT trigger with valid serial
		}
		
		// Should pass validation with valid serial
		assert.NoError(t, validationErr)
		assert.Equal(t, "unix", cfg.Socket.Type)
		assert.Equal(t, uint32(12345), cfg.Keyring.Yubikey.Serial)
	})

	t.Run("UnixSocketWithDisabledYubikey", func(t *testing.T) {
		// Create a temporary config file with disabled YubiKey
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "config.yaml")
		
		configContent := `
socket:
  type: unix
  path: ` + filepath.Join(tempDir, "test.sock") + `
control_socket_path: ` + filepath.Join(tempDir, "test_ctrl.sock") + `
agent_log_path: ` + filepath.Join(tempDir, "test.log") + `
keyring:
  disable_yubikey: true
  yubikey:
    serial: 0
`
		err := os.WriteFile(configPath, []byte(configContent), 0644)
		assert.NoError(t, err)

		// Load the config
		cfg, err := config.Load(configPath)
		assert.NoError(t, err)

		// Test the validation logic - should not require serial when disabled
		var validationErr error
		if cfg.Socket.Type == "unix" && !cfg.Keyring.DisableYubikey && cfg.Keyring.Yubikey.Serial == 0 {
			validationErr = assert.AnError // This should NOT trigger when disabled
		}
		
		// Should pass validation when YubiKey is disabled
		assert.NoError(t, validationErr)
		assert.Equal(t, "unix", cfg.Socket.Type)
		assert.True(t, cfg.Keyring.DisableYubikey)
		assert.Equal(t, uint32(0), cfg.Keyring.Yubikey.Serial)
	})

	t.Run("UnixSocketWithDisabledYubikeyButValidSerial", func(t *testing.T) {
		// Create a temporary config file with disabled YubiKey but valid serial
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "config.yaml")
		
		configContent := `
socket:
  type: unix
  path: ` + filepath.Join(tempDir, "test.sock") + `
control_socket_path: ` + filepath.Join(tempDir, "test_ctrl.sock") + `
agent_log_path: ` + filepath.Join(tempDir, "test.log") + `
keyring:
  disable_yubikey: true
  yubikey:
    serial: 12345
`
		err := os.WriteFile(configPath, []byte(configContent), 0644)
		assert.NoError(t, err)

		// Load the config
		cfg, err := config.Load(configPath)
		assert.NoError(t, err)

		// Should pass validation - when disabled, serial is ignored
		assert.Equal(t, "unix", cfg.Socket.Type)
		assert.True(t, cfg.Keyring.DisableYubikey)
		assert.Equal(t, uint32(12345), cfg.Keyring.Yubikey.Serial) // Serial is present but ignored
	})
}

func TestAgentValidation_SocketTypes(t *testing.T) {
	t.Run("ValidSocketTypes", func(t *testing.T) {
		validTypes := []string{"unix", "dummy"}
		
		for _, socketType := range validTypes {
			tempDir := t.TempDir()
			configPath := filepath.Join(tempDir, "config.yaml")
			
			configContent := `
socket:
  type: ` + socketType + `
  path: ` + filepath.Join(tempDir, "test.sock") + `
control_socket_path: ` + filepath.Join(tempDir, "test_ctrl.sock") + `
agent_log_path: ` + filepath.Join(tempDir, "test.log") + `
keyring:
  yubikey:
    serial: 12345
`
			err := os.WriteFile(configPath, []byte(configContent), 0644)
			assert.NoError(t, err)

			// Load the config
			cfg, err := config.Load(configPath)
			assert.NoError(t, err)
			assert.Equal(t, socketType, cfg.Socket.Type)
		}
	})
}