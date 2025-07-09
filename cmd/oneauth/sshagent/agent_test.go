package sshagent

import (
	"net"
	"os"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vitalvas/oneauth/cmd/oneauth/config"
	"github.com/vitalvas/oneauth/internal/mock"
	"github.com/vitalvas/oneauth/internal/yubikey"
)

func TestNew(t *testing.T) {
	t.Run("WithYubikey", func(t *testing.T) {
		cards, err := yubikey.Cards()
		if err != nil || len(cards) == 0 {
			t.Skip("No YubiKey available")
		}

		log := logrus.New()
		cfg := &config.Config{
			Keyring: config.Keyring{
				BeforeSignHook: "echo test",
				KeepKeySeconds: 300,
			},
		}

		agent, err := New(cards[0].Serial, log, cfg)
		require.NoError(t, err)
		assert.NotNil(t, agent)
		assert.Equal(t, "echo test", agent.actions.BeforeSignHook)

		agent.Close()
	})

	t.Run("WithInvalidSerial", func(t *testing.T) {
		log := logrus.New()
		cfg := &config.Config{}

		agent, err := New(999999, log, cfg)
		assert.Error(t, err)
		assert.Nil(t, agent)
	})
}

func TestSSHAgent_BasicOperations(t *testing.T) {
	agent := &SSHAgent{
		softKeys: mock.NewKeystore(),
	}

	t.Run("Close", func(t *testing.T) {
		err := agent.Close()
		assert.NoError(t, err)
	})

	t.Run("Shutdown", func(t *testing.T) {
		err := agent.Shutdown()
		assert.NoError(t, err)
	})

	t.Run("ShutdownWithListener", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "sshagent_test")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		socketPath := tmpDir + "/test.sock"
		listener, err := net.Listen("unix", socketPath)
		require.NoError(t, err)

		agent.setListener(listener)
		err = agent.Shutdown()
		assert.NoError(t, err)
	})
}

func TestSSHAgent_HandleConn(t *testing.T) {
	agent := &SSHAgent{
		log:      logrus.New().WithField("test", "handleConn"),
		softKeys: mock.NewKeystore(),
	}

	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	done := make(chan bool)
	go func() {
		agent.handleConn(server)
		done <- true
	}()

	client.Close()

	select {
	case <-done:
		// Success
	case <-time.After(100 * time.Millisecond):
		t.Fatal("handleConn did not complete")
	}
}
