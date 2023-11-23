package sshagent

import (
	"fmt"
	"os"

	"github.com/vitalvas/oneauth/internal/netutil"
)

func checkCreds(creds *netutil.UnixCreds) error {
	uid := os.Getuid()

	if creds.UID > 0 && creds.UID != uid {
		return fmt.Errorf("connection from another user (except root) is prohibited: %d (connection) != %d (expected)", creds.UID, uid)
	}

	return nil
}
