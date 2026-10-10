//go:build windows

package share

import (
	"os/exec"
	"syscall"
	"testing"
)

func createDirectoryLink(t *testing.T, name, target string) {
	t.Helper()
	cmd := exec.Command("cmd.exe", "/c", "mklink", "/J", name, target)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if err := cmd.Run(); err != nil {
		t.Fatal("cannot create isolated test junction:", err)
	}
}
