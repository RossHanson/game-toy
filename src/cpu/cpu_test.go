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
	cpu.B.Assign(byte(0x12))
	cpu.A.Assign(0) // to be explicit
	opcode := Ld8BitRegisterOpCode{
		r1: &cpu.A,
		r2: &cpu.B,
	}
	result, err := opcode.Run(cpu)
	if err != nil {
		t.Fatalf("Error running opcode: %v", err)
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
		memoryByte, _ := memory.GetInt(address)
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
	memoryAddress := 0x33
	addressLsb, addressMsb := utils.EncodeInt(memoryAddress)
	cpu.BC.Assign(addressLsb, addressMsb)

	memory.Set(addressLsb, addressMsb, byte(0x12))

	opcode := LdMemIntoRegOpCode{
		r1: &cpu.A,
		r2: &cpu.BC,
	}
	cycles, err := opcode.Run(cpu)

	if err != nil {
		t.Fatalf("Error running opcode: %v", err)
	}

	if cycles != 8 {
		t.Errorf("Incorrect number of cycles returned, want: 8, got: %d", cycles)
	}

	r1Val := cpu.A.Retrieve()
	if r1Val != byte(0x12) {
		t.Errorf("R1 not set properly, want: 0x12, got: %x", r1Val)
	}

	r2Lsb, r2Msb := cpu.BC.Retrieve()
	if utils.CompareByteArrays([]byte{addressLsb, addressMsb}, []byte{r2Lsb, r2Msb}) != 0 {
		t.Errorf("R2 modified, want: %x, got: %x", []byte{addressLsb, addressMsb}, []byte{r2Lsb, r2Msb})
	}
}

func TestLdMemorySet(t *testing.T) {
	cpu, memory := setupCpu()
	cpu.A.Assign(byte(0x45))
	memoryAddress := 0x18
	addressLsb, addressMsb := utils.EncodeInt(memoryAddress)
	cpu.DE.Assign(addressLsb, addressMsb)
	memory.Set(addressLsb, addressMsb, byte(0x00)) // Explicitly zero out that address

	opcode := &LdRegIntoMemOpCode{
		r1: &cpu.DE,
		r2: &cpu.A,
	}

	cycles, err := opcode.Run(cpu)
	if err != nil {
		t.Fatalf("Error running opcode: %v", err)
	}

	if cycles != 8 {
		t.Errorf("Incorrect number of cycles returned, want: 8, got: %d", cycles)
	}

	memoryValue, err := memory.Get(addressLsb, addressMsb)
	if err != nil {
		t.Fatalf("Error reading memory at %x: %v", []byte{addressLsb, addressMsb}, err)
	}
	
	if memoryValue != byte(0x45) {
		t.Errorf("Incorrect value in memory, want: 0x45, got: %x", memoryValue)
	}

	r1Val := cpu.A.Retrieve()
	if r1Val != byte(0x45) {
		t.Errorf("R1 value modified, want: 0x45, got: %x", r1Val)
	}

	r2Lsb, r2Msb := cpu.DE.Retrieve()
	if utils.CompareByteArrays([]byte{addressLsb, addressMsb}, []byte{r2Lsb, r2Msb}) != 0 {
		t.Errorf("R2 modified, want: %x, got: %x", []byte{addressLsb, addressMsb}, []byte{r2Lsb, r2Msb})
	}
}

func TestLdImmediateOpCode(t *testing.T) {
	cpu, memory := setupCpu()

	pcAddress := 0x24
	addressLsb, addressMsb := utils.EncodeInt(pcAddress)
	cpu.PC.Assign(addressLsb, addressMsb)

	memory.SetInt(pcAddress + 1, 0xDE)
	memory.SetInt(pcAddress + 2, 0xAD)

	opCode := &Ld8BitImmediateOpCode{
		r1: &cpu.A,
	}
	opCode.length = 2

	cycles, err := opCode.Run(cpu)
	if err != nil {
		t.Fatalf("Error running opcode: %v", err)
	}

	if cycles != 8 {
		t.Errorf("Incorrect number of cycles, want: 8, got: %d", cycles)
	}

	r1Val := cpu.A.Retrieve()
	if r1Val != byte(0xDE) {
		t.Errorf("R1 modified, want: 0xDE, got: %x", r1Val)
	}
}
