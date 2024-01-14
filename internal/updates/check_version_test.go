package updates

import (
	"fmt"
	"runtime"
	"testing"
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
