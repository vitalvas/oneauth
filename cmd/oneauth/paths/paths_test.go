package paths

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAgentSocket(t *testing.T) {
	home, err := os.UserHomeDir()
	assert.Nil(t, err, "Error getting user home directory: %v", err)

	actual, err := AgentSocket()
	assert.Nil(t, err, "Error getting agent socket path: %v", err)

	expected := filepath.Join(home, oneauthDir, "ssh-agent.sock")
	assert.Equal(t, expected, actual, "Expected result: %s, got result: %s", expected, actual)
}

func TestControlSocket(t *testing.T) {
	home, err := os.UserHomeDir()
	assert.Nil(t, err, "Error getting user home directory: %v", err)

	actual, err := ControlSocket()
	assert.Nil(t, err, "Error getting control socket path: %v", err)

	expected := filepath.Join(home, oneauthDir, "control.sock")
	assert.Equal(t, expected, actual, "Expected result: %s, got result: %s", expected, actual)
}

func TestConfig(t *testing.T) {
	home, err := os.UserHomeDir()
	assert.Nil(t, err, "Error getting user home directory: %v", err)

	expected := filepath.Join(home, oneauthDir, "config.yaml")

	actual, err := Config()
	assert.Nil(t, err, "Error getting config path: %v", err)

	assert.Equal(t, expected, actual, "Expected result: %s, got result: %s", expected, actual)
}

func TestBinDir(t *testing.T) {
	home, err := os.UserHomeDir()
	assert.Nil(t, err, "Error getting user home directory: %v", err)

	actual, err := BinDir()
	assert.Nil(t, err, "Error getting bin directory: %v", err)

	expected := filepath.Join(home, oneauthDir, "bin")

	assert.Equal(t, expected, actual, "Expected result: %s, got result: %s", expected, actual)
}

func TestServiceFile(t *testing.T) {

	var (
		correctDir string
		name       = "oneauth"
	)

	switch runtime.GOOS {
	case "darwin":
		correctDir = "Library/LaunchAgents/oneauth.plist"
	case "linux":
		correctDir = ".config/systemd/user/oneauth.service"
	}

	path, err := ServiceFile(name)
	assert.Nil(t, err, "Error getting service file path: %v", err)

	if !strings.HasSuffix(path, correctDir) {
		t.Errorf("Expected result to be in %s, got result: %s", correctDir, path)
	}
}
