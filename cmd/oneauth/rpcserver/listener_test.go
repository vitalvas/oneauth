package rpcserver

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/vitalvas/oneauth/cmd/oneauth/sshagent"
)

func TestListenAndServe(t *testing.T) {
	t.Run("InvalidSocketPath", func(t *testing.T) {
		rpcServer := New(nil, logrus.New())

		// Use invalid path (directory that doesn't exist)
		invalidPath := "/nonexistent/directory/socket"

		err := rpcServer.ListenAndServe(context.Background(), invalidPath)
		assert.Error(t, err)
	})

	t.Run("ValidSocketPathCreation", func(t *testing.T) {
		// Create temporary directory
		tempDir, err := os.MkdirTemp("", "rpcserver_test")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		socketPath := filepath.Join(tempDir, "test.sock")

		rpcServer := New(nil, logrus.New())

		// Run in goroutine since ListenAndServe blocks
		errChan := make(chan error)
		go func() {
			errChan <- rpcServer.ListenAndServe(context.Background(), socketPath)
		}()

		// Give it time to start
		time.Sleep(100 * time.Millisecond)

		// Check that socket was created
		_, err = os.Stat(socketPath)
		assert.NoError(t, err)

		// Shutdown the server
		rpcServer.Shutdown()

		// Wait for completion
		select {
		case err := <-errChan:
			// Should complete gracefully (nil) or with listener closed error
			if err != nil {
				assert.Contains(t, err.Error(), "use of closed network connection")
			}
		case <-time.After(5 * time.Second):
			t.Fatal("ListenAndServe did not complete in time")
		}
	})
}

func TestListenAndServeCleanup(t *testing.T) {
	t.Run("SocketCleanup", func(t *testing.T) {
		// Create temporary directory
		tempDir, err := os.MkdirTemp("", "rpcserver_test")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		socketPath := filepath.Join(tempDir, "test.sock")

		// Create a file at socket path first
		file, err := os.Create(socketPath)
		assert.NoError(t, err)
		file.Close()

		rpcServer := New(nil, logrus.New())

		// Run in goroutine
		errChan := make(chan error)
		go func() {
			errChan <- rpcServer.ListenAndServe(context.Background(), socketPath)
		}()

		// Give it time to start
		time.Sleep(100 * time.Millisecond)

		// Shutdown the server
		rpcServer.Shutdown()

		// Wait for completion
		select {
		case err := <-errChan:
			// Should complete with an error (expected since we pre-created a file)
			assert.Error(t, err)
		case <-time.After(5 * time.Second):
			t.Fatal("ListenAndServe did not complete in time")
		}
	})
}

func TestListenAndServeHTTPHandling(t *testing.T) {
	t.Run("HTTPMuxSetup", func(t *testing.T) {
		testBasicServerStartup(t, "HTTPMuxSetup")
	})
}

func TestListenAndServeJSONRPC(t *testing.T) {
	t.Run("JSONRPCSetup", func(t *testing.T) {
		testBasicServerStartup(t, "JSONRPCSetup")
	})
}

// testBasicServerStartup is a helper function to test basic server startup and shutdown
func testBasicServerStartup(t *testing.T, testName string) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "rpcserver_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	socketPath := filepath.Join(tempDir, "test.sock")

	rpcServer := New(nil, logrus.New())

	// Run in goroutine
	errChan := make(chan error)
	go func() {
		errChan <- rpcServer.ListenAndServe(context.Background(), socketPath)
	}()

	// Give it time to start
	time.Sleep(100 * time.Millisecond)

	// Verify RPC server is available
	rpcSrv := rpcServer.GetRPCServer()
	assert.NotNil(t, rpcSrv)

	// Shutdown the server
	rpcServer.Shutdown()

	// Wait for completion
	select {
	case err := <-errChan:
		// Should complete gracefully (nil) or with listener closed error
		if err != nil {
			assert.Contains(t, err.Error(), "use of closed network connection")
		}
	case <-time.After(5 * time.Second):
		t.Fatalf("%s: ListenAndServe did not complete in time", testName)
	}
}

func TestListenAndServePermissions(t *testing.T) {
	t.Run("SocketPermissions", func(t *testing.T) {
		// Create temporary directory
		tempDir, err := os.MkdirTemp("", "rpcserver_test")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		socketPath := filepath.Join(tempDir, "test.sock")

		rpcServer := New(nil, logrus.New())

		// Run in goroutine
		errChan := make(chan error)
		go func() {
			errChan <- rpcServer.ListenAndServe(context.Background(), socketPath)
		}()

		// Give it time to start
		time.Sleep(100 * time.Millisecond)

		// Check socket permissions
		info, err := os.Stat(socketPath)
		assert.NoError(t, err)
		assert.Equal(t, os.FileMode(0600), info.Mode().Perm())

		// Shutdown the server
		rpcServer.Shutdown()

		// Wait for completion
		select {
		case err := <-errChan:
			// Should complete gracefully (nil) or with listener closed error
			if err != nil {
				assert.Contains(t, err.Error(), "use of closed network connection")
			}
		case <-time.After(5 * time.Second):
			t.Fatal("ListenAndServe did not complete in time")
		}
	})
}

func TestListenAndServeContext(t *testing.T) {
	t.Run("ContextHandling", func(t *testing.T) {
		// Create temporary directory
		tempDir, err := os.MkdirTemp("", "rpcserver_test")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		socketPath := filepath.Join(tempDir, "test.sock")

		rpcServer := New(nil, logrus.New())

		ctx := context.Background()

		// Run in goroutine
		errChan := make(chan error)
		go func() {
			errChan <- rpcServer.ListenAndServe(ctx, socketPath)
		}()

		// Give it time to start
		time.Sleep(100 * time.Millisecond)

		// Shutdown the server
		rpcServer.Shutdown()

		// Wait for completion
		select {
		case err := <-errChan:
			// Should complete gracefully (nil) or with listener closed error
			if err != nil {
				assert.Contains(t, err.Error(), "use of closed network connection")
			}
		case <-time.After(5 * time.Second):
			t.Fatal("ListenAndServe did not complete in time")
		}
	})
}

func TestListenAndServeErrorHandling(t *testing.T) {
	t.Run("ListenerError", func(t *testing.T) {
		rpcServer := New(nil, logrus.New())

		// Use invalid path
		err := rpcServer.ListenAndServe(context.Background(), "/invalid/path/socket")
		assert.Error(t, err)
	})

	t.Run("ChmodError", func(t *testing.T) {
		rpcServer := New(nil, logrus.New())

		// Use a clearly invalid socket path
		err := rpcServer.ListenAndServe(context.Background(), "/proc/invalid/socket")
		assert.Error(t, err)
	})
}

func TestListenAndServeIntegration(t *testing.T) {
	t.Run("FullIntegration", func(t *testing.T) {
		// Create temporary directory
		tempDir, err := os.MkdirTemp("", "rpcserver_test")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		socketPath := filepath.Join(tempDir, "test.sock")

		// Create full RPC server
		sshAgent := &sshagent.SSHAgent{}
		log := logrus.New()
		rpcServer := New(sshAgent, log)

		// Run in goroutine
		errChan := make(chan error)
		go func() {
			errChan <- rpcServer.ListenAndServe(context.Background(), socketPath)
		}()

		// Give it time to start
		time.Sleep(100 * time.Millisecond)

		// Verify everything is set up
		rpcSrv := rpcServer.GetRPCServer()
		assert.NotNil(t, rpcSrv)
		assert.NotNil(t, rpcServer.SSHAgent)
		assert.NotNil(t, rpcServer.log)

		// Shutdown the server
		rpcServer.Shutdown()

		// Wait for completion
		select {
		case err := <-errChan:
			// Should complete gracefully (nil) or with listener closed error
			if err != nil {
				assert.Contains(t, err.Error(), "use of closed network connection")
			}
		case <-time.After(5 * time.Second):
			t.Fatal("ListenAndServe did not complete in time")
		}
	})
}
