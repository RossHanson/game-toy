// Package main begins the emulator
package main

import (
	"os"
	"fmt"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Incorrect usage, expected %s <path-to-cartridge>\n", os.Args[0])
		os.Exit(2) // TODO - make sure this is the right error code for bad usage
	}

	cartridgePath := os.Args[1]
	if _, err := os.Stat(cartridgePath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "No cartridge found at path %s\n", cartridgePath)
		os.Exit(2)
	} else if err != nil {
		fmt.Fprintf(os.Stderr, "Error statting cartridge file at path %s - %v", cartridgePath, err)
		os.Exit(2)
	}
}
