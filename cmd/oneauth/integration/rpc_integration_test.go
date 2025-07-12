package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/vitalvas/oneauth/cmd/oneauth/rpclient"
	"github.com/vitalvas/oneauth/cmd/oneauth/rpcserver"
	"github.com/vitalvas/oneauth/cmd/oneauth/sshagent"
)

func TestRPCIntegration(t *testing.T) {
	t.Run("FullServerClientIntegration", func(t *testing.T) {
		// Create temporary directory
		tempDir, err := os.MkdirTemp("", "integration_test")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		socketPath := filepath.Join(tempDir, "test.sock")

		// Create RPC server
		sshAgent := &sshagent.SSHAgent{}
		log := logrus.New()
		log.SetLevel(logrus.FatalLevel) // Reduce log noise in tests
		rpcServer := rpcserver.New(sshAgent, log)

		// Start server in goroutine
		serverErrChan := make(chan error)
		go func() {
			serverErrChan <- rpcServer.ListenAndServe(context.Background(), socketPath)
		}()

		// Give server time to start
		time.Sleep(100 * time.Millisecond)

		// Verify socket was created
		_, err = os.Stat(socketPath)
		assert.NoError(t, err)

		// Create client and test connection
		client, err := rpclient.New(socketPath)
		assert.NoError(t, err)
		assert.NotNil(t, client)
		defer client.Close()

		// Test RPC call
		info, err := client.GetInfo()
		assert.NoError(t, err)
		assert.NotNil(t, info)
		assert.Equal(t, os.Getpid(), info.Pid)

		// Test multiple concurrent calls
		doneChan := make(chan bool, 5)
		for i := 0; i < 5; i++ {
			go func() {
				info, err := client.GetInfo()
				assert.NoError(t, err)
				assert.NotNil(t, info)
				assert.Equal(t, os.Getpid(), info.Pid)
				doneChan <- true
			}()
		}

		// Wait for all goroutines to complete
		for i := 0; i < 5; i++ {
			<-doneChan
		}

		// Close client
		err = client.Close()
		assert.NoError(t, err)

		// Shutdown server
		rpcServer.Shutdown()

		// Wait for server to shutdown
		select {
		case err := <-serverErrChan:
			// Should complete with listener closed error (expected)
			if err != nil {
				assert.Contains(t, err.Error(), "use of closed network connection")
			}
		case <-time.After(5 * time.Second):
			t.Fatal("Server did not shutdown in time")
		}
	})

	t.Run("ClientConnectToNonExistentServer", func(t *testing.T) {
		nonExistentSocketPath := "/tmp/non_existent_socket.sock"

		client, err := rpclient.New(nonExistentSocketPath)
		assert.Error(t, err)
		assert.Nil(t, client)
	})

	t.Run("ClientCallAfterServerShutdown", func(t *testing.T) {
		// Create temporary directory
		tempDir, err := os.MkdirTemp("", "integration_test")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		socketPath := filepath.Join(tempDir, "test.sock")

		// Create and start server
		log := logrus.New()
		log.SetLevel(logrus.FatalLevel)
		server := rpcserver.New(nil, log)

		go func() {
			server.ListenAndServe(context.Background(), socketPath)
		}()

		// Give server time to start
		time.Sleep(100 * time.Millisecond)

		// Create client and verify it works
		client, err := rpclient.New(socketPath)
		assert.NoError(t, err)

		info, err := client.GetInfo()
		assert.NoError(t, err)
		assert.NotNil(t, info)

		client.Close()

		// Shutdown server
		server.Shutdown()
		time.Sleep(100 * time.Millisecond)

		// Try to create new client after shutdown - should fail
		client2, err := rpclient.New(socketPath)
		assert.Error(t, err)
		assert.Nil(t, client2)
	})
}
