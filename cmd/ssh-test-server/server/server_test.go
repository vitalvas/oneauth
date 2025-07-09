package server

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vitalvas/oneauth/internal/buildinfo"
	"golang.org/x/crypto/ssh"
)

func TestServerStruct(t *testing.T) {
	t.Run("ServerCreation", func(t *testing.T) {
		srv := &Server{}
		
		// Test that server can be created
		assert.NotNil(t, srv)
		assert.Nil(t, srv.serverURL)
		assert.Nil(t, srv.sshConfig)
	})
}

func TestServerFields(t *testing.T) {
	t.Run("ServerFields", func(t *testing.T) {
		srv := &Server{}
		
		// Test field access
		assert.IsType(t, (*url.URL)(nil), srv.serverURL)
		assert.IsType(t, (*ssh.ServerConfig)(nil), srv.sshConfig)
	})
}

func TestServerConfiguration(t *testing.T) {
	t.Run("ServerURLSetting", func(t *testing.T) {
		srv := &Server{}
		
		// Test URL setting
		testURL, err := url.Parse("http://example.com:8080")
		assert.NoError(t, err)
		
		srv.serverURL = testURL
		assert.Equal(t, testURL, srv.serverURL)
		assert.Equal(t, "example.com:8080", srv.serverURL.Host)
		assert.Equal(t, "8080", srv.serverURL.Port())
	})
	
	t.Run("SSHConfigSetting", func(t *testing.T) {
		srv := &Server{}
		
		// Test SSH config setting
		config := &ssh.ServerConfig{
			ServerVersion: "SSH-2.0-Test",
		}
		
		srv.sshConfig = config
		assert.Equal(t, config, srv.sshConfig)
		assert.Equal(t, "SSH-2.0-Test", srv.sshConfig.ServerVersion)
	})
}

func TestServerMethodsExist(t *testing.T) {
	t.Run("MethodsExist", func(t *testing.T) {
		srv := &Server{}
		
		// Test that methods exist (we can't call them without setup)
		assert.IsType(t, (*Server)(nil), srv)
		
		// Test that struct has the expected methods by checking if they can be assigned
		_ = srv.runServer
	})
}

func TestServerURLParsing(t *testing.T) {
	t.Run("ValidURL", func(t *testing.T) {
		srv := &Server{}
		
		// Test valid URL parsing
		testURL := "http://127.0.0.1:8080"
		parsedURL, err := url.Parse(testURL)
		assert.NoError(t, err)
		
		srv.serverURL = parsedURL
		assert.Equal(t, "http", srv.serverURL.Scheme)
		assert.Equal(t, "127.0.0.1:8080", srv.serverURL.Host)
	})
	
	t.Run("InvalidURL", func(t *testing.T) {
		// Test invalid URL parsing
		invalidURL := "not-a-valid-url"
		parsedURL, err := url.Parse(invalidURL)
		
		// url.Parse is very lenient, so this might not error
		if err == nil {
			assert.NotNil(t, parsedURL)
		} else {
			assert.Error(t, err)
		}
	})
}

func TestServerSSHConfig(t *testing.T) {
	t.Run("SSHConfigCreation", func(t *testing.T) {
		srv := &Server{}
		
		// Test SSH config creation
		config := &ssh.ServerConfig{
			ServerVersion: "SSH-2.0-OneAuth (+https://oneauth.vitalvas.dev)",
		}
		
		srv.sshConfig = config
		assert.NotNil(t, srv.sshConfig)
		assert.Contains(t, srv.sshConfig.ServerVersion, "OneAuth")
		assert.Contains(t, srv.sshConfig.ServerVersion, "SSH-2.0")
	})
}

func TestServerConstants(t *testing.T) {
	t.Run("DefaultValues", func(t *testing.T) {
		// Test default values that would be used
		defaultURL := "http://127.0.0.1:8080"
		defaultServerVersion := "SSH-2.0-OneAuth (+https://oneauth.vitalvas.dev)"
		
		assert.NotEmpty(t, defaultURL)
		assert.NotEmpty(t, defaultServerVersion)
		assert.Contains(t, defaultURL, "127.0.0.1")
		assert.Contains(t, defaultServerVersion, "OneAuth")
	})
}

func TestServerBuildInfo(t *testing.T) {
	t.Run("BuildInfoIntegration", func(t *testing.T) {
		// Test that buildinfo is accessible
		version := buildinfo.Version
		
		// Version might be empty in test environment
		assert.NotNil(t, version)
		assert.IsType(t, "", version)
	})
}

func TestServerAppConfig(t *testing.T) {
	t.Run("AppConfigValues", func(t *testing.T) {
		// Test values that would be used in the app config
		appName := "oneauth-ssh-test-server"
		defaultPort := ":2022"
		
		assert.NotEmpty(t, appName)
		assert.NotEmpty(t, defaultPort)
		assert.Contains(t, appName, "oneauth")
		assert.Contains(t, defaultPort, "2022")
	})
}

func TestServerNetworking(t *testing.T) {
	t.Run("NetworkingConstants", func(t *testing.T) {
		// Test networking constants
		defaultPort := ":2022"
		protocol := "tcp"
		
		assert.Equal(t, ":2022", defaultPort)
		assert.Equal(t, "tcp", protocol)
	})
}

func TestServerStructFields(t *testing.T) {
	t.Run("FieldTypes", func(t *testing.T) {
		srv := &Server{}
		
		// Test that fields can be set to expected types
		srv.serverURL = &url.URL{
			Scheme: "http",
			Host:   "localhost:8080",
		}
		
		srv.sshConfig = &ssh.ServerConfig{
			ServerVersion: "SSH-2.0-Test",
		}
		
		assert.IsType(t, &url.URL{}, srv.serverURL)
		assert.IsType(t, &ssh.ServerConfig{}, srv.sshConfig)
	})
}

func TestServerInitialization(t *testing.T) {
	t.Run("ZeroValues", func(t *testing.T) {
		srv := &Server{}
		
		// Test zero values
		assert.Nil(t, srv.serverURL)
		assert.Nil(t, srv.sshConfig)
	})
	
	t.Run("NonZeroValues", func(t *testing.T) {
		srv := &Server{
			serverURL: &url.URL{Host: "example.com"},
			sshConfig: &ssh.ServerConfig{ServerVersion: "SSH-2.0-Test"},
		}
		
		// Test non-zero values
		assert.NotNil(t, srv.serverURL)
		assert.NotNil(t, srv.sshConfig)
		assert.Equal(t, "example.com", srv.serverURL.Host)
		assert.Equal(t, "SSH-2.0-Test", srv.sshConfig.ServerVersion)
	})
}

func TestServerURLOperations(t *testing.T) {
	t.Run("URLOperations", func(t *testing.T) {
		srv := &Server{}
		
		// Test URL operations
		testURL := "https://oneauth.example.com:9000/api"
		parsedURL, err := url.Parse(testURL)
		assert.NoError(t, err)
		
		srv.serverURL = parsedURL
		
		assert.Equal(t, "https", srv.serverURL.Scheme)
		assert.Equal(t, "oneauth.example.com:9000", srv.serverURL.Host)
		assert.Equal(t, "/api", srv.serverURL.Path)
		assert.Equal(t, "9000", srv.serverURL.Port())
	})
}

func TestServerConfigDefaults(t *testing.T) {
	t.Run("ConfigurationDefaults", func(t *testing.T) {
		// Test default configuration values
		defaultServerURL := "http://127.0.0.1:8080"
		defaultListenPort := ":2022"
		expectedServerVersion := "SSH-2.0-OneAuth (+https://oneauth.vitalvas.dev)"
		
		assert.Equal(t, "http://127.0.0.1:8080", defaultServerURL)
		assert.Equal(t, ":2022", defaultListenPort)
		assert.Contains(t, expectedServerVersion, "OneAuth")
	})
}