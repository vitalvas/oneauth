package tools

import "os/user"

func IsRoot() bool {
	u, err := user.Current()
	if err != nil {
		return false
	}

	return u.Username == "root"
}
