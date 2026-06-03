//go:build !gui
// +build !gui

// CLI-only build (default) - used by Homebrew
package main

import (
	"fmt"
	"os"

	"github.com/PopolQue/dupclean/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
