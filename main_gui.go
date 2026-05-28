//go:build gui
// +build gui

// Full build with GUI - used for GitHub Releases
package main

import (
	"dupclean/cmd"
	"dupclean/gui"
)

func main() {
	cmd.LaunchGUI = gui.RunGUI
	cmd.Execute()
}
