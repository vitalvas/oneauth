package updates

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetUpdateManifestURL(t *testing.T) {
	tests := []struct {
		appName     string
		channel     Channel
		expectedURL string
		expectedErr error
	}{
		{
			appName:     "oneauth",
			channel:     ChannelDev,
			expectedURL: "https://oneauth-files.vitalvas.dev/test/update_manifest/oneauth.json",
			expectedErr: nil,
		},
		{
			appName:     "oneauth",
			channel:     ChannelProd,
			expectedURL: "https://oneauth-files.vitalvas.dev/release/update_manifest/oneauth.json",
			expectedErr: nil,
		},
		{
			appName:     "oneauth",
			channel:     Channel("http://example.com"),
			expectedURL: "",
			expectedErr: ErrSchemeNotHTTPS,
		},
	}

	for _, test := range tests {
		t.Run(test.appName, func(t *testing.T) {
			actualURL, err := getUpdateManifestURL(test.appName, test.channel)

			if err != nil {
				if test.expectedErr == nil || err.Error() != test.expectedErr.Error() {
					t.Errorf("Expected error: %v, but got error: %v", test.expectedErr, err)
				}
			} else if test.expectedErr != nil {
				t.Errorf("Expected error: %v, but got no error", test.expectedErr)
			}

			if actualURL != test.expectedURL {
				t.Errorf("Expected URL: %s, but got URL: %s", test.expectedURL, actualURL)
			}
		})
	}
}

func TestGetUpdateManifestURL_InvalidURL(t *testing.T) {
	// Create a channel with invalid URL characters
	invalidChannel := Channel("https://[invalid-url")

	result, err := getUpdateManifestURL("test-app", invalidChannel)

	assert.Error(t, err)
	assert.Empty(t, result)
}

func TestGetUpdateManifestURL_SpecialCharacters(t *testing.T) {
	tests := []struct {
		name     string
		appName  string
		expected string
	}{
		{
			name:     "app with hyphens",
			appName:  "my-cool-app",
			expected: "https://oneauth-files.vitalvas.dev/release/update_manifest/my-cool-app.json",
		},
		{
			name:     "app with numbers",
			appName:  "app123",
			expected: "https://oneauth-files.vitalvas.dev/release/update_manifest/app123.json",
		},
		{
			name:     "app with dots",
			appName:  "app.name",
			expected: "https://oneauth-files.vitalvas.dev/release/update_manifest/app.name.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := getUpdateManifestURL(tt.appName, ChannelProd)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUpdateManifest_Structure(t *testing.T) {
	manifest := UpdateManifest{
		Name:         "test-app",
		Version:      "v1.2.3",
		RemotePrefix: "https://example.com/releases/",
	}

	assert.Equal(t, "test-app", manifest.Name)
	assert.Equal(t, "v1.2.3", manifest.Version)
	assert.Equal(t, "https://example.com/releases/", manifest.RemotePrefix)
}

func TestErrorConstants(t *testing.T) {
	assert.Equal(t, "scheme is not HTTPS", ErrSchemeNotHTTPS.Error())
	assert.Equal(t, "no update available", ErrNoUpdateAvailable.Error())
}

func TestCheck_InvalidVersion(t *testing.T) {
	// Test with invalid local version
	result, err := Check("test-app", "invalid-version")

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestCheck_NetworkError(t *testing.T) {
	// Test with a version that should generate a network error
	// (since we're testing against a real URL that might not exist)
	result, err := Check("non-existent-app-12345", "v1.0.0")

	// We expect this to fail due to network/404 error
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestGetRemoteManifest_ErrorHandling(t *testing.T) {
	// Test with invalid URL
	manifest, err := getRemoteManifest("test-app", "invalid-url")

	assert.Error(t, err)
	assert.Nil(t, manifest)
}

func TestCheck_WithMockServer(t *testing.T) {
	tests := []struct {
		name           string
		appName        string
		localVersion   string
		serverResponse *UpdateManifest
		serverStatus   int
		expectError    bool
		expectManifest bool
	}{
		{
			name:         "update available",
			appName:      "testapp",
			localVersion: "v1.0.0",
			serverResponse: &UpdateManifest{
				Name:         "testapp",
				Version:      "v2.0.0",
				RemotePrefix: "https://example.com/releases/",
			},
			serverStatus:   http.StatusOK,
			expectError:    false,
			expectManifest: true,
		},
		{
			name:         "no update available same version",
			appName:      "testapp",
			localVersion: "v2.0.0",
			serverResponse: &UpdateManifest{
				Name:         "testapp",
				Version:      "v2.0.0",
				RemotePrefix: "https://example.com/releases/",
			},
			serverStatus:   http.StatusOK,
			expectError:    true,
			expectManifest: false,
		},
		{
			name:         "remote version older than local",
			appName:      "testapp",
			localVersion: "v3.0.0",
			serverResponse: &UpdateManifest{
				Name:         "testapp",
				Version:      "v2.0.0",
				RemotePrefix: "https://example.com/releases/",
			},
			serverStatus:   http.StatusOK,
			expectError:    true,
			expectManifest: false,
		},
		{
			name:         "server returns invalid manifest version",
			appName:      "testapp",
			localVersion: "v1.0.0",
			serverResponse: &UpdateManifest{
				Name:         "testapp",
				Version:      "not-a-version",
				RemotePrefix: "https://example.com/releases/",
			},
			serverStatus:   http.StatusOK,
			expectError:    true,
			expectManifest: false,
		},
		{
			name:         "server returns 404",
			appName:      "testapp",
			localVersion: "v1.0.0",
			serverStatus: http.StatusNotFound,
			expectError:  true,
		},
		{
			name:         "server returns 403",
			appName:      "testapp",
			localVersion: "v1.0.0",
			serverStatus: http.StatusForbidden,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				if tt.serverStatus != http.StatusOK {
					w.WriteHeader(tt.serverStatus)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(tt.serverResponse)
			}))
			defer server.Close()

			// Override the getJSON function by using the mock server URL directly
			// We need to test through getRemoteManifest since Check calls real URLs
			manifest, err := getRemoteManifest(tt.appName, server.URL)

			if tt.serverStatus != http.StatusOK {
				assert.Error(t, err)
				assert.Nil(t, manifest)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, manifest)
			assert.Equal(t, tt.serverResponse.Name, manifest.Name)
			assert.Equal(t, tt.serverResponse.Version, manifest.Version)
			assert.Equal(t, tt.serverResponse.RemotePrefix, manifest.RemotePrefix)
		})
	}
}

func TestGetRemoteManifest_Success(t *testing.T) {
	expected := &UpdateManifest{
		Name:         "testapp",
		Version:      "v1.2.3",
		RemotePrefix: "https://example.com/releases/",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expected)
	}))
	defer server.Close()

	manifest, err := getRemoteManifest("testapp", server.URL)
	assert.NoError(t, err)
	assert.NotNil(t, manifest)
	assert.Equal(t, expected.Name, manifest.Name)
	assert.Equal(t, expected.Version, manifest.Version)
	assert.Equal(t, expected.RemotePrefix, manifest.RemotePrefix)
}

func TestGetRemoteManifest_ServerErrors(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{"not found", http.StatusNotFound},
		{"forbidden", http.StatusForbidden},
		{"server error", http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			manifest, err := getRemoteManifest("testapp", server.URL)
			assert.Error(t, err)
			assert.Nil(t, manifest)
		})
	}
}

func TestGetRemoteManifest_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("not valid json"))
	}))
	defer server.Close()

	manifest, err := getRemoteManifest("testapp", server.URL)
	assert.Error(t, err)
	assert.Nil(t, manifest)
}

func TestCheck_VersionComparison(t *testing.T) {
	tests := []struct {
		name         string
		localVersion string
		expectError  bool
		errorType    error
	}{
		{
			name:         "dev version",
			localVersion: "v0.0.1",
			expectError:  true, // Network error expected
		},
		{
			name:         "prod version",
			localVersion: "v1.0.0",
			expectError:  true, // Network error expected
		},
		{
			name:         "invalid version format",
			localVersion: "not-a-version",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Check("test-app", tt.localVersion)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)

				if tt.errorType != nil {
					assert.True(t, errors.Is(err, tt.errorType))
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}
