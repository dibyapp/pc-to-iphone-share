//go:build !windows

package share

import (
	"os"
	"testing"
)

func createDirectoryLink(t *testing.T, name, target string) {
	t.Helper()
	if err := os.Symlink(target, name); err != nil {
		t.Fatal(err)
	}
}
