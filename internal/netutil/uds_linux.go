package netutil

import (
	"fmt"
	"net"
)

func UnixSocketCreds(conn net.Conn) (UnixCreds, error) {
	unixConn, ok := conn.(*net.UnixConn)
	if !ok {
		return UnixCreds{
			UID: -1,
			PID: -1,
		}, nil
	}

	raw, err := unixConn.SyscallConn()
	if err != nil {
		return UnixCreds{}, fmt.Errorf("failed to get syscall conn: %w", err)
	}

	var cred *unix.Ucred
	cerr := raw.Control(func(fd uintptr) {
		cred, err = unix.GetsockoptUcred(int(fd), unix.SOL_SOCKET, unix.SO_PEERCRED)
	})
	if cerr != nil {
		return UnixCreds{}, fmt.Errorf("failed to get peer credentials: %w", cerr)
	}

	if err != nil {
		return UnixCreds{}, fmt.Errorf("failed to get peer credentials: %w", err)
	}

	return UnixCreds{
		UID: int(cred.Uid),
		PID: int(cred.Pid),
	}, nil
}
