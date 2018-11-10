package main

import (
	"cpu"
	"memory"
)

func main() {
	mem := memory.InitializeMainMemory()
	c := cpu.NewCpu(mem)
	c.PrintKnownOpCodes()
}
