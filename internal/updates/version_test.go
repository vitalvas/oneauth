package updates

import "testing"

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
