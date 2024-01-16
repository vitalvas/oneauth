package updates

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
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

		switch r.URL.Path {
		case "/success":
			json.NewEncoder(w).Encode(map[string]interface{}{"key": "value"})

		case "/notfound":
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)

		case "/forbidden":
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)

		default:
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	}))

	defer server.Close()

	testCases := []struct {
		description string
		remote      string
		expectedErr error
	}{
		{
			description: "Successful response",
			remote:      server.URL + "/success",
			expectedErr: nil,
		},
		{
			description: "Not Found response",
			remote:      server.URL + "/notfound",
			expectedErr: ErrUpdateNotFound,
		},
		{
			description: "Forbidden response",
			remote:      server.URL + "/forbidden",
			expectedErr: ErrUpdateForbidden,
		},
		{
			description: "Unexpected status code",
			remote:      server.URL + "/unknown",
			expectedErr: errors.New("unexpected status code: 500"),
		},
	}

	for _, test := range testCases {
		t.Run(test.description, func(t *testing.T) {
			var response map[string]interface{}

			err := getJSON("TestApp", test.remote, &response)

			if err != nil && err.Error() != test.expectedErr.Error() {
				t.Errorf("Expected error: '%v', but got: '%v'", test.expectedErr, err)
			}

			if err == nil && response == nil {
				t.Errorf("Expected response to be not nil")
			}
		})
	}
}
