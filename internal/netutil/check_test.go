package netutil

import (
	"os"
	"testing"
)

func TestCheckCreds(t *testing.T) {
	tests := []struct {
		name      string
		UID       int
		exceptErr bool
	}{
		{"root", 0, false},
		{"current-user", os.Getuid(), false},
		{"non-valid", 667, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := CheckCreds(&UnixCreds{
				UID: test.UID,
			})

			if test.exceptErr && err == nil {
				t.Errorf("Expected error, but got no error")
			}

			if !test.exceptErr && err != nil {
				t.Errorf("Expected no error, but got: %v", err)
			}
		})
	}
}
