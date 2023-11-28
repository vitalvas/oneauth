package netutil

import (
	"fmt"
	"os"
)

func CheckCreds(creds *UnixCreds) error {
	uid := os.Getuid()

	if creds.UID > 0 && creds.UID != uid {
		return fmt.Errorf("connection from another user (except root) is prohibited: %d (connection) != %d (expected)", creds.UID, uid)
	}

	return nil
}
