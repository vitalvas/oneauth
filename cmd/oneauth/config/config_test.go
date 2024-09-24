package config

import (
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
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
