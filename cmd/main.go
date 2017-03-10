package main

import (
	"os"

	"github.com/cyberbliss/smokeshift/pkg/util"
)

// Set via linker flag
var version string

func main() {
	cmd := NewSmokeshiftCommand(version, os.Stdin, os.Stdout)

	if err := cmd.Execute(); err != nil {
		util.PrintColor(os.Stderr, util.Red, "Error running command: %v\n", err)
		os.Exit(1)
	}
}
