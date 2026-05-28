//go:build !gui
// +build !gui

// CLI-only build (default) - used by Homebrew
package main

import (
	"dupclean/cmd"
)

func main() {
	cmd.Execute()
}
