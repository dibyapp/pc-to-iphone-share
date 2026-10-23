//go:build !windows

package share

import (
	"os"
	"syscall"
)

// Nonblocking open also prevents a concurrently swapped FIFO from pinning a handler.
// Ordinary files and directory reads retain normal semantics.
func openRootRead(root *os.Root, name string) (*os.File, error) {
	return root.OpenFile(name, os.O_RDONLY|syscall.O_NONBLOCK, 0)
}
