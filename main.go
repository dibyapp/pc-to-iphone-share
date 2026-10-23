package main

import (
	"fmt"
	"github.com/dibyapp/pc-to-iphone-share/internal/share"
	"os"
)

var version = "0.2.0-alpha.1"

func main() {
	if err := share.Run(os.Args[1:], version, os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "PC to iPhone Share:", err)
		os.Exit(1)
	}
}
