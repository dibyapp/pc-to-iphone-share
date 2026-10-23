//go:build !windows

package share

import (
	"path/filepath"
	"syscall"
	"testing"
)

func TestSpecialFilesAreNeverOpenedForTransfer(t *testing.T) {
	f := setupFixture(t)
	f.login(t)
	if err := syscall.Mkfifo(filepath.Join(f.root, "pipe.txt"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, route := range []string{"/download?path=pipe.txt", "/download.zip?path=pipe.txt", "/api/checksum?path=pipe.txt", "/api/preview?path=pipe.txt"} {
		expectStatus(t, f.request(t, "GET", route, nil, nil), 400)
	}
}
