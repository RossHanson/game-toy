package main

import (
	"fmt"
	"os"
	"cpu"
	"memory"
)

func main() {
	fmt.Fprintf(os.Stdout, "Hello there!\n")
	memory := memory.InitializeMainMemory()
	cpu := cpu.NewCpu(memory)
	// Start the CPU next
	fmt.Fprintf(os.Stdout, "Made a cpu! %+v", *cpu)
}
