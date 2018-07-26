package cpu

import (
	"testing"
	"memory"
	"types"
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

func setupCpu() (*Cpu, *memory.Memory) {
	memory := memory.SetupBlankMemory(mockMemorySize)
	return NewCpu(memory), memory
}

func TestLdOpCode_registerToRegister(t *testing.T) {
	cpu, memory := setupCpu()
	cpu.B.Assign(byte(0x12))
	cpu.A.Assign(0) // to be explicit
	opcode := Ld8BitRegisterOpCode{
		r1: &cpu.A,
		r2: &cpu.B,
	}
	result, pcModified, err := opcode.Run(cpu)
	if err != nil {
		t.Fatalf("Error running opcode: %v", err)
	}
	if pcModified {
		t.Error("Incorrectly changed PC, want: false, got: true")
	}
	if result != 4 {
		t.Errorf("Incorrect number of cycles, want: 4, got: %d", result)
	}
	r1Value := cpu.A.Retrieve()
	if r1Value != byte(0x12) {
		t.Errorf("R1 not set correctly, want: b, got: %x", r1Value)
	}
	r2Value := cpu.B.Retrieve()
	if r2Value  != byte(0x12) {
		t.Errorf("R2 changed, want: b, got: %x", r2Value)
	}
	// Verify memory was not updated
	for address := 0; address < mockMemorySize; address++ {
		memoryByte, _ := memory.Get(types.Word(address))
		if memoryByte != 0 {
			t.Errorf("Memory updated inappropriately, want: 0x00, got: %x",
				memoryByte)
			break // Don't need to print a million errors
		}
	}
}

func TestLdMemoryRead(t *testing.T) {
	cpu, memory := setupCpu()
	cpu.A.Assign(0)
	memoryAddress := types.Word(0x33)
	cpu.BC.Assign(memoryAddress)

	memory.Set(memoryAddress, byte(0x12))

	opcode := LdMemIntoRegOpCode{
		r1: &cpu.A,
		r2: &cpu.BC,
	}
	cycles, pcModified, err := opcode.Run(cpu)

	if err != nil {
		t.Fatalf("Error running opcode: %v", err)
	}
	if pcModified {
		t.Errorf("PC modified incorrectly, want: false, got: true")
	}

	if cycles != 8 {
		t.Errorf("Incorrect number of cycles returned, want: 8, got: %d", cycles)
	}

	r1Val := cpu.A.Retrieve()
	if r1Val != byte(0x12) {
		t.Errorf("R1 not set properly, want: 0x12, got: %x", r1Val)
	}

	r2Val := cpu.BC.Retrieve()
	if r2Val != memoryAddress {
		t.Errorf("R2 modified, want: %s, got: %s", memoryAddress, r2Val)
	}
}

func TestLdMemorySet(t *testing.T) {
	testCases := []struct{
		name string
		isIncrement bool
		isDecrement bool
	} {
		{
			name: "Normal load",
		},
		{
			name: "Post load increment",
			isIncrement: true,
		},
		{
			name: "Post load decrement",
			isDecrement: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cpu, memory := setupCpu()
			memoryAddress := types.Word(0x18)
			memory.Set(memoryAddress, byte(0x00)) // Explicitly zero out

			cpu.A.Assign(byte(0x45))
			cpu.DE.Assign(memoryAddress)

			opCode := &LdRegIntoMemOpCode{
				r1: &cpu.DE,
				r2: &cpu.A,
				incrementR1: tc.isIncrement,
				decrementR1: tc.isDecrement,
			}

			cycles, pcModified, err := opCode.Run(cpu)
			if err != nil {
				t.Fatalf("Error running test: %v", err)
			}
			if pcModified {
				t.Error("PC incorrectly modified, want: false, got: true")
			}
			if cycles != 8 {
				t.Errorf("Incorrect number of cycles, want: 8, got: %d", cycles)
			}

			memoryValue, err := memory.Get(memoryAddress)
			if err != nil {
				t.Fatalf("Error fetching memory at %s: %v", memoryAddress, err)
			}
			if memoryValue != byte(0x45) {
				t.Errorf("Incorrect memory value, want: 0x45, got: %x", memoryValue)
			}

			r1Val := cpu.A.Retrieve()
			if r1Val != byte(0x45) {
				t.Errorf("R1 value modified, want: 0x45, got: %x", r1Val)
			}

			r2Val := cpu.DE.Retrieve()
			expectedValue := memoryAddress
			if tc.isIncrement {
				expectedValue += types.Word(1)
			} else if tc.isDecrement {
				expectedValue -= types.Word(1)
			}
			if r2Val != expectedValue {
				t.Errorf("R2 val incorrect, want: %s, got: %s", expectedValue, r2Val)
			}
		})
	}
}

func TestLdImmediateByteOpCode(t *testing.T) {
	cpu, memory := setupCpu()

	pcAddress := types.Word(0x24)
	cpu.PC.Assign(pcAddress)

	memory.Set(pcAddress + 1, 0xDE)

	opCode := &Ld8BitImmediateOpCode{
		r1: &cpu.A,
	}

	cycles, pcModified, err := opCode.Run(cpu)
	if err != nil {
		t.Fatalf("Error running opcode: %v", err)
	}
	if pcModified {
		t.Errorf("PC modified incorrectly, want: false, got: true")
	}

	if cycles != 8 {
		t.Errorf("Incorrect number of cycles, want: 8, got: %d", cycles)
	}

	r1Val := cpu.A.Retrieve()
	if r1Val != byte(0xDE) {
		t.Errorf("R1 modified, want: 0xDE, got: %x", r1Val)
	}
}

func TestLdImmediateWordOpCode(t *testing.T) {
	cpu, memory := setupCpu()

	pcAddress := types.Word(0x32)
	cpu.PC.Assign(pcAddress)
	memory.Set(pcAddress + 1, 0xBE)
	memory.Set(pcAddress + 2, 0xEF)

	opCode := &Ld16BitImmediateOpCode{
		r1: &cpu.DE,
	}
	cycles, pcModified, err := opCode.Run(cpu)
	if err != nil {
		t.Fatalf("Error running opcode: %v", err)
	}
	if pcModified {
		t.Errorf("PC modified incorrectly, want: false, got: true")
	}
	if cycles != 12 {
		t.Errorf("Incorrect number of cycles, want: 12, got: %d", cycles)
	}

	r1Val := cpu.DE.Retrieve()
	if r1Val != types.WordFromBytes(0xBE, 0xEF) {
		t.Errorf("Incorrect r1val, want: 0xBEEF, got: %s", r1Val)
	}
}

func TestIncDecOpCodes(t *testing.T) {
	testCases := []struct{
		name string
		inputValue byte
		expectedValue byte
		expectedZeroFlag bool
		// Probably should figure out half-carry?
	} {
		{
			name: "Normal increment",
			inputValue: byte(0x09),
			expectedValue: byte(0x0A),
			expectedZeroFlag: false,
		},
		{
			name: "Roll-over",
			inputValue: byte(0xFF),
			expectedValue: byte(0x00),
			expectedZeroFlag: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cpu, _ := setupCpu()

			cpu.A.Assign(tc.inputValue)

			opCode := &IncRegOpCode{
				r1: &cpu.A,
			}

			cycles, pcModified, err := opCode.Run(cpu)
			if err != nil {
				t.Fatalf("Error running opcode: %v", err)
			}

			if pcModified {
				t.Error("PC incorrectly modified, want: false, got: true")
			}

			if cycles != 4 {
				t.Errorf("Incorrect number of cycles, want: 4, got: %d", cycles)
			}

			r1Val := cpu.A.Retrieve()
			if r1Val != tc.expectedValue {
				t.Errorf("Incorrect R1 val, want: %x, got: %x", tc.expectedValue, r1Val)
			}

			if cpu.GetFlag(Z) != tc.expectedZeroFlag {
				t.Errorf("Incorrect zero flag, want: %t, got: %t", tc.expectedZeroFlag, cpu.GetFlag(Z))
			}

			if cpu.GetFlag(N) {
				t.Errorf("Incorrect subtract flag, want: false, got: true")
			}
		})
	}
}
