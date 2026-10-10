//go:build windows

package share

import (
	"errors"
	"os/exec"
	"os/user"
	"syscall"
)

func protectState(dir string) error {
	owner, err := user.Current()
	if err != nil {
		return errors.New("cannot identify the Windows account for private settings")
	}
	// Arguments are passed directly, never through a shell. No shared folder ACLs are changed.
	cmd := exec.Command("icacls.exe", dir, "/inheritance:r", "/grant:r", "*"+owner.Uid+":(OI)(CI)F", "*S-1-5-18:(OI)(CI)F")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if err := cmd.Run(); err != nil {
		return errors.New("cannot restrict the private settings ACL")
	}
	return nil
}
