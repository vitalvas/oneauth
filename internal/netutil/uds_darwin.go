package netutil

import (
	"errors"
	"fmt"
	"net"
	"syscall"

	"golang.org/x/sys/unix"
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

	var (
		cred   *unix.Xucred
		pid    int
		interr error
	)

	if err = raw.Control(func(fd uintptr) {
		cred, err = unix.GetsockoptXucred(int(fd), unix.SOL_LOCAL, unix.LOCAL_PEERCRED)
		if err != nil {
			interr = fmt.Errorf("failed to get peer credentials: %w", err)
			return
		}

		pid, err = unix.GetsockoptInt(int(fd), unix.SOL_LOCAL, unix.LOCAL_PEERPID)
		if err != nil {
			// ENOTCONN means that the socket is not connected, which is fine.
			if errors.Is(err, syscall.ENOTCONN) {
				err = nil
				pid = -1
			} else {
				interr = fmt.Errorf("failed to get peer pid: %w", err)
			}
		}
	}); err != nil {
		return UnixCreds{}, fmt.Errorf("failed to get unix control: %w", err)
	}

	if interr != nil {
		return UnixCreds{}, interr
	}

	return UnixCreds{
		UID: int(cred.Uid),
		PID: pid,
	}, nil
}
