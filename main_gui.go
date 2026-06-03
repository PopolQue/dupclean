//go:build gui
// +build gui

// Full build with GUI - used for GitHub Releases
package main

import (
	"fmt"

	"github.com/PopolQue/dupclean/cmd"
	"github.com/PopolQue/dupclean/gui"
)

func main() {
	if err := gui.SetupLogging(); err != nil {
		fmt.Printf("Failed to setup logging: %v\n", err)
	}
	cmd.LaunchGUI = gui.RunGUI
	cmd.Execute()
}
