package updates

import (
	"testing"
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
