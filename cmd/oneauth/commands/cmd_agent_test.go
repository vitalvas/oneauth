package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
	"github.com/vitalvas/oneauth/internal/buildinfo"
)

func TestAgentCmd(t *testing.T) {
	t.Run("CommandMetadata", func(t *testing.T) {
		assert.Equal(t, "agent", agentCmd.Name)
		assert.Equal(t, "SSH Agent", agentCmd.Usage)
		assert.NotEmpty(t, agentCmd.Description)
		assert.NotNil(t, agentCmd.Before)
		assert.NotNil(t, agentCmd.Action)
	})

	t.Run("BeforeHook", func(t *testing.T) {
		tests := []struct {
			name    string
			version string
			commit  string
		}{
			{name: "EmptyCommit", version: "1.0.0", commit: ""},
			{name: "ShortCommit", version: "1.0.0", commit: "abc"},
			{name: "LongCommit", version: "1.0.0", commit: "abcdef1234567890"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				origVersion := buildinfo.Version
				origCommit := buildinfo.Commit
				defer func() {
					buildinfo.Version = origVersion
					buildinfo.Commit = origCommit
				}()

				buildinfo.Version = tt.version
				buildinfo.Commit = tt.commit

				// Run with invalid config so Before executes but Action fails early
				app := &cli.App{
					Flags: []cli.Flag{
						&cli.PathFlag{
							Name:  "config",
							Value: "/nonexistent/config.yaml",
						},
					},
					Commands: []*cli.Command{agentCmd},
				}

				err := app.Run([]string{"app", "agent"})
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "failed to load config")
			})
		}
	})

	t.Run("ActionInvalidConfig", func(t *testing.T) {
		app := &cli.App{
			Flags: []cli.Flag{
				&cli.PathFlag{
					Name:  "config",
					Value: "/nonexistent/config.yaml",
				},
			},
			Commands: []*cli.Command{agentCmd},
		}

		err := app.Run([]string{"app", "agent"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to load config")
	})

	t.Run("ActionWithConfig", func(t *testing.T) {
		tests := []struct {
			name          string
			config        string
			expectedError string
		}{
			{
				name: "MissingYubikeySerial",
				config: `
socket:
  type: "dummy"
`,
				expectedError: "yubikey serial is required",
			},
			{
				name: "UnsupportedSocketType",
				config: `
socket:
  type: "unsupported"
keyring:
  yubikey:
    serial: 12345
`,
				expectedError: "socket type unsupported is not supported",
			},
			{
				name: "DummySocketZeroSerial",
				config: `
socket:
  type: "dummy"
keyring:
  yubikey:
    serial: 0
`,
				expectedError: "yubikey serial is required",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				tmpHome, err := os.MkdirTemp("", "test-home-*")
				require.NoError(t, err)
				defer os.RemoveAll(tmpHome)

				oneauthDir := filepath.Join(tmpHome, ".oneauth")
				err = os.MkdirAll(oneauthDir, 0755)
				require.NoError(t, err)

				originalHome := os.Getenv("HOME")
				os.Setenv("HOME", tmpHome)
				defer os.Setenv("HOME", originalHome)

				tmpFile, err := os.CreateTemp("", "config-*.yaml")
				require.NoError(t, err)
				defer os.Remove(tmpFile.Name())
				defer tmpFile.Close()

				_, err = tmpFile.Write([]byte(tt.config))
				require.NoError(t, err)

				app := &cli.App{
					Flags: []cli.Flag{
						&cli.PathFlag{
							Name:  "config",
							Value: tmpFile.Name(),
						},
					},
					Commands: []*cli.Command{agentCmd},
				}

				err = app.Run([]string{"app", "agent"})
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			})
		}
	})
}
