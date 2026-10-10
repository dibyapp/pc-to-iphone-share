//go:build !windows

package share

import (
	"errors"
	"os"
)

func protectState(dir string) error {
	if err := os.Chmod(dir, 0700); err != nil {
		return errors.New("cannot restrict private settings directory permissions")
	}
	return nil
}
