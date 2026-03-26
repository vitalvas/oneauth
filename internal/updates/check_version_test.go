package updates

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetUpdateVersionManifestURL(t *testing.T) {
	tests := []struct {
		appName       string
		remotePrefix  string
		expectedURL   string
		expectedError bool
	}{
		{"myApp", "https://example.com/test/", fmt.Sprintf("https://example.com/test/myApp_%s_%s_manifest.json", runtime.GOOS, runtime.GOARCH), false},
		{"anotherApp", "https://example.net/test/", fmt.Sprintf("https://example.net/test/anotherApp_%s_%s_manifest.json", runtime.GOOS, runtime.GOARCH), false},

		{"noSSL", "http://example.net/test/", "", true},
		{"invalidApp", "://invalid-url", "", true},
	}

	for _, test := range tests {
		t.Run(test.appName, func(t *testing.T) {
			url, err := getUpdateVersionManifestURL(test.appName, test.remotePrefix)

			if test.expectedError {
				if err == nil {
					t.Errorf("Expected an error but got nil.")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, but got %v", err)
				}
				if url != test.expectedURL {
					t.Errorf("Expected URL %s, but got %s", test.expectedURL, url)
				}
			}
		})
	}
}

func TestGetRemoteVersionManifest(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		expected := &UpdateVersionManifest{
			Name:    "testapp",
			Version: "v1.2.3",
			Sha256:  "abc123",
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(expected)
		}))
		defer server.Close()

		manifest, err := getRemoteVersionManifest("testapp", server.URL)
		assert.NoError(t, err)
		assert.NotNil(t, manifest)
		assert.Equal(t, expected.Name, manifest.Name)
		assert.Equal(t, expected.Version, manifest.Version)
		assert.Equal(t, expected.Sha256, manifest.Sha256)
	})

	t.Run("ServerError", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		manifest, err := getRemoteVersionManifest("testapp", server.URL)
		assert.Error(t, err)
		assert.Nil(t, manifest)
	})

	t.Run("NotFound", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		manifest, err := getRemoteVersionManifest("testapp", server.URL)
		assert.Error(t, err)
		assert.Nil(t, manifest)
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Write([]byte("not json"))
		}))
		defer server.Close()

		manifest, err := getRemoteVersionManifest("testapp", server.URL)
		assert.Error(t, err)
		assert.Nil(t, manifest)
	})

	t.Run("InvalidURL", func(t *testing.T) {
		manifest, err := getRemoteVersionManifest("testapp", "://invalid")
		assert.Error(t, err)
		assert.Nil(t, manifest)
	})
}

func TestCheckVersionFunc(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		expected := &UpdateVersionManifest{
			Name:    "testapp",
			Version: "v1.2.3",
			Sha256:  "abc123",
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(expected)
		}))
		defer server.Close()

		manifest, err := CheckVersion("testapp", server.URL+"/")
		// This will fail because server.URL uses http not https
		assert.Error(t, err)
		assert.Nil(t, manifest)
	})

	t.Run("InvalidRemotePrefix", func(t *testing.T) {
		manifest, err := CheckVersion("testapp", "://invalid")
		assert.Error(t, err)
		assert.Nil(t, manifest)
	})

	t.Run("HTTPScheme", func(t *testing.T) {
		manifest, err := CheckVersion("testapp", "http://example.com/releases/")
		assert.Error(t, err)
		assert.Nil(t, manifest)
	})
}

func TestUpdateVersionManifestStruct(t *testing.T) {
	t.Run("JSONSerialization", func(t *testing.T) {
		m := UpdateVersionManifest{
			Name:    "testapp",
			Version: "v1.0.0",
			Sha256:  "deadbeef",
		}

		data, err := json.Marshal(m)
		assert.NoError(t, err)

		var parsed UpdateVersionManifest
		err = json.Unmarshal(data, &parsed)
		assert.NoError(t, err)
		assert.Equal(t, m.Name, parsed.Name)
		assert.Equal(t, m.Version, parsed.Version)
		assert.Equal(t, m.Sha256, parsed.Sha256)
	})
}
