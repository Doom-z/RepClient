package fileutil

import "os"

func EnsureDir(path string) error {
	// currentdir
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.MkdirAll(path, 0755)
	}
	return nil
}
