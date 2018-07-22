package gametoy

import (
	"testing"
)

const (
	mockMemorySize = 1024 // Use smaller memory size for speed
)

func TestOpCodeName(t *testing.T) {
	name := "Test op code"
	testOpCode := BaseOpCode{
		name: name,
	}
	if name != testOpCode.Name() {
		t.Errorf("Name mismatch, want: %s, got: %s", name,
			testOpCode.Name())
	}
}

func setupCpu() (*Cpu, *Memory) {
	memory := &Memory{
		memory: make([]byte, mockMemorySize),
	}
	return newCpu(memory), memory
}

func TestLdOpCode_registerToRegister(t *testing.T) {
	cpu, memory := setupCpu()
	*cpu.registers["B"].Value[0] = byte('b')
	*cpu.registers["A"].Value[0] = 0 // to be explicit
	opcode := LdRegisterOpCode{
		r1: "A",
		r2: "B",
	}
	result, err := opcode.Run(cpu)
	if err != nil {
		t.Fatalf("Error running opcode: %v", err)
	}
	if result != 4 {
		t.Errorf("Incorrect number of cycles, want: 4, got: %d", result)
	}
	if *cpu.registers["A"].Value[0] != byte('b') {
		t.Errorf("R1 not set correctly, want: 0x12, got: %x",
			cpu.registers["A"].Value[0])
	}
	if *cpu.registers["B"].Value[0] != byte('b') {
		t.Errorf("R2 changed, want: 0x12, got: %x",
			cpu.registers["B"].Value[0])
	}
	// Verify memory was not updated
	for _, memoryByte := range memory.memory {
		if memoryByte != 0 {
			t.Errorf("Memory updated inappropriately, want: 0x00, got: %x",
				memoryByte)
			break // Don't need to print a million errors
		}
	}
	// Register to memory
	// Memory to register
}

