package updates

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckVersion(t *testing.T) {
	tests := []struct {
		version string
		wantErr bool
	}{
		{"v0.1.0", false},
		{"v0.0.0", false},
		{"v0.0.1704954878", false},

		{"v0.1.0.0", true},
		{"v0.1", true},
		{"v0", true},
		{"v", true},
		{"0.1.0", true},
		{"test", true},
		{"", true},
	}

	for _, test := range tests {
		t.Run(test.version, func(t *testing.T) {
			if resp, err := checkVersion(test.version); (err != nil) != test.wantErr {
				t.Errorf("version = %s, error = %v, wantErr %v", test.version, err, test.wantErr)

				if err != ErrInvalidVersion {
					t.Error("error is not ErrInvalidVersion")
				}

			} else if err == nil && resp == nil {
				t.Error("version is nil")
			}
		})
	}
}

func TestCheckNewVersion(t *testing.T) {
	tests := []struct {
		name           string
		currentVersion string
		newVersion     string
		expectedResult bool
		expectError    bool
	}{
		{
			name:           "New version is greater",
			currentVersion: "v1.0.0",
			newVersion:     "v1.1.0",
			expectedResult: true,
			expectError:    false,
		},
		{
			name:           "New version is equal",
			currentVersion: "v1.0.0",
			newVersion:     "v1.0.0",
			expectedResult: false,
			expectError:    false,
		},
		{
			name:           "New version is lower",
			currentVersion: "v1.1.0",
			newVersion:     "v1.0.0",
			expectedResult: false,
			expectError:    false,
		},
		{
			name:           "Invalid current version",
			currentVersion: "invalid",
			newVersion:     "v1.1.0",
			expectedResult: false,
			expectError:    true,
		},
		{
			name:           "Invalid new version",
			currentVersion: "v1.0.0",
			newVersion:     "invalid",
			expectedResult: false,
			expectError:    true,
		},
		{
			name:           "Empty current version",
			currentVersion: "",
			newVersion:     "v1.1.0",
			expectedResult: false,
			expectError:    true,
		},
		{
			name:           "Empty new version",
			currentVersion: "v1.0.0",
			newVersion:     "",
			expectedResult: false,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CheckNewVersion(tt.currentVersion, tt.newVersion)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedResult, result)
		})
	}
}
