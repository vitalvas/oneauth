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

func TestGetJSON(t *testing.T) {
	expectedUA := fmt.Sprintf(
		"Mozilla/5.0 (compatible; %s/%s; os/%s; arch/%s)",
		"TestApp", buildinfo.Version, buildinfo.OS, buildinfo.ARCH,
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate a successful response with JSON data.
		if r.Header.Get("User-Agent") != expectedUA {
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
			remote:      fmt.Sprintf("%s/success", server.URL),
			expectedErr: nil,
		},
		{
			description: "Not Found response",
			remote:      fmt.Sprintf("%s/notfound", server.URL),
			expectedErr: ErrUpdateNotFound,
		},
		{
			description: "Forbidden response",
			remote:      fmt.Sprintf("%s/forbidden", server.URL),
			expectedErr: ErrUpdateForbidden,
		},
		{
			description: "Unexpected status code",
			remote:      fmt.Sprintf("%s/unknown", server.URL),
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
