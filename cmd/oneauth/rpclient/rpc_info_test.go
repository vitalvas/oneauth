package rpclient

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vitalvas/oneauth/cmd/oneauth/rpcserver"
)

func TestGetInfo(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse string
		serverStatus   int
		expectError    bool
		expectedInfo   *rpcserver.Info
	}{
		{
			name:           "successful response",
			serverResponse: `{"pid": 123}`,
			serverStatus:   http.StatusOK,
			expectError:    false,
			expectedInfo:   &rpcserver.Info{Pid: 123},
		},
		{
			name:           "server error",
			serverResponse: `Internal Server Error`,
			serverStatus:   http.StatusInternalServerError,
			expectError:    true,
			expectedInfo:   nil,
		},
		{
			name:           "invalid JSON response",
			serverResponse: `invalid json`,
			serverStatus:   http.StatusOK,
			expectError:    true,
			expectedInfo:   nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(test.serverStatus)
				w.Write([]byte(test.serverResponse))
			}))
			defer server.Close()

			baseURLParsed, err := url.Parse(server.URL)
			if err != nil {
				assert.Error(t, err)
				return
			}

			client := &Client{
				client:  server.Client(),
				baseURL: baseURLParsed,
			}

			info, err := client.GetInfo()

			if test.expectError && err == nil {
				t.Errorf("Expected error, but got none")
			}
			if !test.expectError && err != nil {
				assert.Nil(t, err)
			}

			if !test.expectError {
				assert.Equal(t, test.expectedInfo, info)
			}
		})
	}
}
