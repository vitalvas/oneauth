package updates

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/vitalvas/oneauth/internal/buildinfo"
)

func TestGetUserAgent(t *testing.T) {
	tests := []struct {
		appName  string
		expected string
	}{
		{"MyApp", fmt.Sprintf("Mozilla/5.0 (compatible; MyApp/%s; os/%s; arch/%s)", buildinfo.Version, buildinfo.OS, buildinfo.ARCH)},
		{"AnotherApp", fmt.Sprintf("Mozilla/5.0 (compatible; AnotherApp/%s; os/%s; arch/%s)", buildinfo.Version, buildinfo.OS, buildinfo.ARCH)},
	}

	for _, test := range tests {
		t.Run(test.appName, func(t *testing.T) {
			result := getUserAget(test.appName)
			if result != test.expected {
				t.Errorf("Expected: %s, Got: %s", test.expected, result)
			}
		})
	}
}

func TestGetJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate a successful response with JSON data.
		if r.Header.Get("User-Agent") != getUserAget("TestApp") {
			http.Error(w, "Invalid User-Agent", http.StatusBadRequest)
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{"key": "value"})
	}))

	defer server.Close()

	var responseData map[string]interface{}
	err := getJSON("TestApp", server.URL, &responseData)
	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}

	expectedData := map[string]interface{}{"key": "value"}

	if !reflect.DeepEqual(responseData, expectedData) {
		t.Errorf("Expected JSON data: %v, but got: %v", expectedData, responseData)
	}

	err = getJSON("InvalidApp", server.URL, &responseData)
	if err == nil {
		t.Errorf("Expected an error, but got nil")
	}
}
