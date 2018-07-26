package cpu

import (
	"testing"
	"memory"
	"utils"
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
	cpu.registers["B"].Assign(byte(0x12))
	cpu.registers["A"].Assign(0) // to be explicit
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
	r1Value := cpu.registers["A"].Retrieve()[0]
	if r1Value != byte(0x12) {
		t.Errorf("R1 not set correctly, want: b, got: %x", r1Value)
	}
	r2Value := cpu.registers["B"].Retrieve()[0]
	if r2Value  != byte(0x12) {
		t.Errorf("R2 changed, want: b, got: %x", r2Value)
	}
	// Verify memory was not updated
	for address := 0; address < mockMemorySize; address++ {
		memoryByte, _ := memory.GetInt(address)
		if memoryByte != 0 {
			t.Errorf("Memory updated inappropriately, want: 0x00, got: %x",
				memoryByte)
			break // Don't need to print a million errors
		}
	}
	// Register to memory
	// Memory to register
}

func TestLdMemoryRead(t *testing.T) {
	cpu, memory := setupCpu()
	cpu.registers["A"].Assign(0)
	memoryAddress := 0x33
	addressBytes := utils.EncodeInt(memoryAddress)
	cpu.registers["BC"].Assign(addressBytes...)

	memory.Set(addressBytes, byte(0x12))

	opcode := LdRegisterOpCode{
		r1: "A",
		r2: "BC",
	}
	cycles, err := opcode.Run(cpu)

	if err != nil {
		t.Fatalf("Error running opcode: %v", err)
	}

	if cycles != 8 {
		t.Errorf("Incorrect number of cycles returned, want: 8, got: %d", cycles)
	}

	r1Val := cpu.registers["A"].Retrieve()[0]
	if r1Val != byte(0x12) {
		t.Errorf("R1 not set properly, want: 0x12, got: %x", r1Val)
	}

	r2Val := cpu.registers["BC"].Retrieve()
	if utils.CompareByteArrays(addressBytes, r2Val) != 0 {
		t.Errorf("R2 modified, want: %x, got: %x", addressBytes, r2Val)
	}
}

func TestLdMemorySet(t *testing.T) {
	cpu, memory := setupCpu()
	cpu.registers["A"].Assign(byte(0x45))
	memoryAddress := 0x18
	addressBytes := utils.EncodeInt(memoryAddress)
	cpu.registers["DE"].Assign(addressBytes...)
	memory.Set(addressBytes, byte(0x00)) // Explicitly zero out that address

	opcode := &LdRegisterOpCode{
		r1: "DE",
		r2: "A",
	}

	cycles, err := opcode.Run(cpu)
	if err != nil {
		t.Fatalf("Error running opcode: %v", err)
	}

	if cycles != 8 {
		t.Errorf("Incorrect number of cycles returned, want: 8, got: %d", cycles)
	}

	memoryValue, err := memory.Get(addressBytes)
	if err != nil {
		t.Fatalf("Error reading memory at %x: %v", addressBytes, err)
	}
	
	if memoryValue != byte(0x45) {
		t.Errorf("Incorrect value in memory, want: 0x45, got: %x", memoryValue)
	}

	r1Val := cpu.registers["A"].Retrieve()[0]
	if r1Val != byte(0x45) {
		t.Errorf("R1 value modified, want: 0x45, got: %x", r1Val)
	}

	r2Val := cpu.registers["DE"].Retrieve()
	if utils.CompareByteArrays(addressBytes, r2Val) != 0 {
		t.Errorf("R2 modified, want: %x, got: %x", addressBytes, r2Val)
	}
}

func TestLdImmediateOpCode(t *testing.T) {
	cpu, memory := setupCpu()

	pcAddress := 0x24
	pcAddressBytes := utils.EncodeInt(pcAddress)
	cpu.programCounter.Assign(pcAddressBytes...)

	memory.SetInt(pcAddress + 1, 0xDE)
	memory.SetInt(pcAddress + 2, 0xAD)
	
	testCases := []struct{
		name string
		r1 RegisterName
		length int
		expectedCycles int
		expectedRegisterValue []byte
	} {
		{
			name: "8 bit register",
			r1: "A",
			length: 2,
			expectedCycles: 8,
			expectedRegisterValue: []byte{byte(0xDE)},
		},
		{
			name: "16 bit register",
			r1: "BC",
			length: 3,
			expectedCycles: 12,
			expectedRegisterValue: []byte{byte(0xDE), byte(0xAD)},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			opCode := &LdImmediateOpCode{
				r1: tc.r1,
			}
			opCode.length = tc.length

			cycles, err := opCode.Run(cpu)
			if err != nil {
				t.Fatalf("Error running opcode: %v", err)
			}

			if cycles != tc.expectedCycles {
				t.Errorf("Incorrect number of cycles, want: %d, got: %d", tc.expectedCycles, cycles)
			}

			r1Val := cpu.registers[tc.r1].Retrieve()
			if utils.CompareByteArrays(r1Val, tc.expectedRegisterValue) != 0 {
				t.Errorf("R1 modified, want: %x, got: %x", r1Val, tc.expectedRegisterValue)
			}

			pcValue := cpu.programCounter.Retrieve()
			if utils.CompareByteArrays(pcValue, pcAddressBytes) != 0 {
				t.Errorf("PC was incorrectly updated, want: %x, got: %x", pcAddressBytes, pcValue)
			}
		})
	}
}
