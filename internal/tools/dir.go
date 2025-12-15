package tools

import "os"

func MkDir(dir string, perm os.FileMode) error {
	stat, err := os.Stat(dir)

	if os.IsNotExist(err) {
		if err = os.MkdirAll(dir, perm); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	if stat != nil && stat.Mode().Perm() != perm {
		if err = os.Chmod(dir, perm); err != nil {
			return err
		}
	}

	return nil
}
