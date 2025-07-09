package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadYamlFile(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		expectError bool
		expect      *Config
	}{
		{
			name: "ValidFile",
			content: `
control_socket_path: "/tmp/control.sock"
agent_log_path: "/var/log/agent.log"
socket:
  type: "unix"
  path: "/tmp/agent.sock"
`,
			expectError: false,
			expect: &Config{
				ControlSocketPath: "/tmp/control.sock",
				AgentLogPath:      "/var/log/agent.log",
				Socket: Socket{
					Type: "unix",
					Path: "/tmp/agent.sock",
				},
			},
		},
		{
			name: "InvalidFile",
			content: `
control_socket_path: "/tmp/control.sock"
agent_log_path: "/var/log/agent.log"
socket:
  type: "unix"
  path: "/tmp/agent.sock
`,
			expectError: true,
		},
		{
			name: "UnknownFields",
			content: `
control_socket_path: "/tmp/control.sock"
agent_log_path: "/var/log/agent.log"
socket:
  type: "unix"
  path: "/tmp/agent.sock"
unknown_field: "unknown"
`,
			expectError: true,
		},
		{
			name:        "NonExistentFile",
			content:     "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tmpFile *os.File
			var err error

			if tt.name != "NonExistentFile" {
				tmpFile, err = os.CreateTemp("", "config-*.yaml")
				if err != nil {
					assert.Error(t, err)
					return
				}

				defer os.Remove(tmpFile.Name())
				defer tmpFile.Close()

				if _, err := tmpFile.Write([]byte(tt.content)); err != nil {
					assert.Error(t, err)
					return
				}
			}

			conf := &Config{}

			filePath := "nonexistent.yaml"
			if tmpFile != nil {
				filePath = tmpFile.Name()
			}

			err = loadYamlFile(filePath, conf)
			if tt.expectError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)

				assert.Equal(t, tt.expect, conf)
			}
		})
	}

	t.Run("Pre-Filled-Config", func(t *testing.T) {
		agentID := uuid.New()

		conf := &Config{
			AgentID: agentID,
			Socket: Socket{
				Type: "unix",
				Path: "/tmp/agent.sock",
			},
		}

		config := []byte(`
control_socket_path: "/tmp/control.sock"
agent_log_path: "/var/log/agent.log"
`)

		expected := &Config{
			AgentID:           agentID,
			ControlSocketPath: "/tmp/control.sock",
			AgentLogPath:      "/var/log/agent.log",
			Socket: Socket{
				Type: "unix",
				Path: "/tmp/agent.sock",
			},
		}

		tmpFile, err := os.CreateTemp("", "config-*.yaml")
		if err != nil {
			assert.Error(t, err)
			return
		}

		defer os.Remove(tmpFile.Name())
		defer tmpFile.Close()

		if _, err := tmpFile.Write(config); err != nil {
			assert.Error(t, err)
			return
		}

		err = loadYamlFile(tmpFile.Name(), conf)
		assert.Nil(t, err)

		assert.Equal(t, expected, conf)
	})
}

func TestLoad(t *testing.T) {
	t.Run("ValidConfigFile", func(t *testing.T) {
		// Create a temporary config file
		tmpFile, err := os.CreateTemp("", "config-*.yaml")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())
		defer tmpFile.Close()

		configContent := `
control_socket_path: "/tmp/test-control.sock"
agent_log_path: "/tmp/test-agent.log"
socket:
  type: "unix"
  path: "/tmp/test-agent.sock"
keyring:
  yubikey:
    serial: 12345
  before_sign_hook: "echo test"
  keep_key_seconds: 300
`
		_, err = tmpFile.Write([]byte(configContent))
		require.NoError(t, err)

		config, err := Load(tmpFile.Name())
		require.NoError(t, err)
		
		assert.NotEqual(t, uuid.Nil, config.AgentID)
		assert.Equal(t, "/tmp/test-control.sock", config.ControlSocketPath)
		assert.Equal(t, "/tmp/test-agent.log", config.AgentLogPath)
		assert.Equal(t, "unix", config.Socket.Type)
		assert.Equal(t, "/tmp/test-agent.sock", config.Socket.Path)
		assert.Equal(t, uint32(12345), config.Keyring.Yubikey.Serial)
		assert.Equal(t, "echo test", config.Keyring.BeforeSignHook)
		assert.Equal(t, int64(300), config.Keyring.KeepKeySeconds)
	})

	t.Run("EmptyConfigFile", func(t *testing.T) {
		// Create an empty config file
		tmpFile, err := os.CreateTemp("", "config-*.yaml")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())
		defer tmpFile.Close()

		// Write a minimal YAML content to avoid EOF error
		_, err = tmpFile.Write([]byte("{}\n"))
		require.NoError(t, err)

		config, err := Load(tmpFile.Name())
		require.NoError(t, err)
		
		// Should have default values and generated agent ID
		assert.NotEqual(t, uuid.Nil, config.AgentID)
		assert.NotEmpty(t, config.ControlSocketPath)
		assert.NotEmpty(t, config.AgentLogPath)
		assert.Equal(t, "unix", config.Socket.Type)
		assert.NotEmpty(t, config.Socket.Path)
	})

	t.Run("NonExistentConfigFile", func(t *testing.T) {
		config, err := Load("/nonexistent/config.yaml")
		assert.Error(t, err)
		assert.Nil(t, config)
	})
}

func TestLoadOrCreateAgentID(t *testing.T) {
	t.Run("CreateNewAgentID", func(t *testing.T) {
		// Create a temporary directory for testing
		tmpDir, err := os.MkdirTemp("", "oneauth-test-")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		// Create the .oneauth directory
		oneauthDir := filepath.Join(tmpDir, ".oneauth")
		err = os.MkdirAll(oneauthDir, 0755)
		require.NoError(t, err)

		// Mock the agent ID path
		originalEnv := os.Getenv("HOME")
		os.Setenv("HOME", tmpDir)
		defer os.Setenv("HOME", originalEnv)

		agentID, err := LoadOrCreateAgentID()
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, agentID)

		// Verify the ID was written to file
		agentIDPath := filepath.Join(tmpDir, ".oneauth", "agent_id")
		_, err = os.Stat(agentIDPath)
		assert.NoError(t, err)
	})

	t.Run("LoadExistingAgentID", func(t *testing.T) {
		// Create a temporary directory for testing
		tmpDir, err := os.MkdirTemp("", "oneauth-test-")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		// Create the config directory
		configDir := filepath.Join(tmpDir, ".oneauth")
		err = os.MkdirAll(configDir, 0755)
		require.NoError(t, err)

		// Create an existing agent ID file
		existingID := uuid.New()
		agentIDPath := filepath.Join(configDir, "agent_id")
		err = os.WriteFile(agentIDPath, []byte(existingID.String()), 0600)
		require.NoError(t, err)

		// Mock the agent ID path
		originalEnv := os.Getenv("HOME")
		os.Setenv("HOME", tmpDir)
		defer os.Setenv("HOME", originalEnv)

		agentID, err := LoadOrCreateAgentID()
		require.NoError(t, err)
		assert.Equal(t, existingID, agentID)
	})

	t.Run("InvalidExistingAgentID", func(t *testing.T) {
		// Create a temporary directory for testing
		tmpDir, err := os.MkdirTemp("", "oneauth-test-")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		// Create the config directory
		configDir := filepath.Join(tmpDir, ".oneauth")
		err = os.MkdirAll(configDir, 0755)
		require.NoError(t, err)

		// Create an invalid agent ID file
		agentIDPath := filepath.Join(configDir, "agent_id")
		err = os.WriteFile(agentIDPath, []byte("invalid-uuid"), 0600)
		require.NoError(t, err)

		// Mock the agent ID path
		originalEnv := os.Getenv("HOME")
		os.Setenv("HOME", tmpDir)
		defer os.Setenv("HOME", originalEnv)

		agentID, err := LoadOrCreateAgentID()
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, agentID)
		
		// Verify a new ID was created
		newContent, err := os.ReadFile(agentIDPath)
		require.NoError(t, err)
		assert.NotEqual(t, "invalid-uuid", string(newContent))
	})
}

func TestConfigStruct(t *testing.T) {
	t.Run("DefaultValues", func(t *testing.T) {
		config := &Config{}
		assert.Equal(t, uuid.Nil, config.AgentID)
		assert.Empty(t, config.ControlSocketPath)
		assert.Empty(t, config.AgentLogPath)
		assert.Equal(t, Socket{}, config.Socket)
		assert.Equal(t, Keyring{}, config.Keyring)
	})

	t.Run("CompleteConfig", func(t *testing.T) {
		agentID := uuid.New()
		config := &Config{
			AgentID:           agentID,
			ControlSocketPath: "/tmp/control.sock",
			AgentLogPath:      "/tmp/agent.log",
			Socket: Socket{
				Type: "unix",
				Path: "/tmp/agent.sock",
			},
			Keyring: Keyring{
				Yubikey: KeyringYubikey{
					Serial: 12345,
				},
				BeforeSignHook: "echo test",
				KeepKeySeconds: 300,
			},
		}

		assert.Equal(t, agentID, config.AgentID)
		assert.Equal(t, "/tmp/control.sock", config.ControlSocketPath)
		assert.Equal(t, "/tmp/agent.log", config.AgentLogPath)
		assert.Equal(t, "unix", config.Socket.Type)
		assert.Equal(t, "/tmp/agent.sock", config.Socket.Path)
		assert.Equal(t, uint32(12345), config.Keyring.Yubikey.Serial)
		assert.Equal(t, "echo test", config.Keyring.BeforeSignHook)
		assert.Equal(t, int64(300), config.Keyring.KeepKeySeconds)
	})
}

func TestLoadYamlFile_EdgeCases(t *testing.T) {
	t.Run("EmptyFile", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "config-*.yaml")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())
		defer tmpFile.Close()

		config := &Config{}
		err = loadYamlFile(tmpFile.Name(), config)
		// Empty file should return EOF error
		assert.Error(t, err)
	})

	t.Run("OnlyWhitespace", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "config-*.yaml")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())
		defer tmpFile.Close()

		_, err = tmpFile.Write([]byte("   \n\t  \n  "))
		require.NoError(t, err)

		config := &Config{}
		err = loadYamlFile(tmpFile.Name(), config)
		// Whitespace with tabs can cause YAML parsing errors
		assert.Error(t, err)
	})

	t.Run("OnlyComments", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "config-*.yaml")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())
		defer tmpFile.Close()

		_, err = tmpFile.Write([]byte("# This is a comment\n# Another comment"))
		require.NoError(t, err)

		config := &Config{}
		err = loadYamlFile(tmpFile.Name(), config)
		// Comments-only file should return EOF error
		assert.Error(t, err)
	})

	t.Run("PartialConfig", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "config-*.yaml")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())
		defer tmpFile.Close()

		_, err = tmpFile.Write([]byte("control_socket_path: \"/tmp/control.sock\""))
		require.NoError(t, err)

		config := &Config{}
		err = loadYamlFile(tmpFile.Name(), config)
		assert.NoError(t, err)
		assert.Equal(t, "/tmp/control.sock", config.ControlSocketPath)
		assert.Empty(t, config.AgentLogPath)
	})

	t.Run("PermissionDenied", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "config-*.yaml")
		require.NoError(t, err)
		defer tmpFile.Close()

		// Remove read permissions
		err = os.Chmod(tmpFile.Name(), 0000)
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		config := &Config{}
		err = loadYamlFile(tmpFile.Name(), config)
		assert.Error(t, err)
	})
}
