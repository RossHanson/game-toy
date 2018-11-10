package cpu

import (
	"testing"
	"memory"
	"fmt"
	"log"
	"math/rand"
	"types"
)

const (
	mockMemorySize = 1024 // Use smaller memory size for speed
)

func setupCpu() (*Cpu, *memory.Memory) {
	memory := memory.SetupBlankMemory(mockMemorySize)
	return NewCpu(memory), memory
}

func setupCpuWithState(state SystemState) (*Cpu, *memory.Memory) {
	memory := memory.SetupBlankMemory(mockMemorySize)
	cpu := NewCpu(memory)
	cpu.A.Assign(state.A)
	cpu.B.Assign(state.B)
	cpu.C.Assign(state.C)
	cpu.D.Assign(state.D)
	cpu.E.Assign(state.E)
	cpu.F.Assign(state.F)
	cpu.H.Assign(state.H)
	cpu.L.Assign(state.L)
	cpu.SP.Assign(state.SP)
	cpu.PC.Assign(state.PC)
	for address, val := range state.memVals {
		memory.Set(address, val)
	}
	return cpu, memory
}

// Sets CPU flags to the opposite of their expected values (or sets them for unchanged ones)
func setupCpuFlags(cpu *Cpu, flags FlagChanges) byte {
	setFlagVal := func(flagChange FlagState) bool {
		if flagChange != FlagUnchanged {
			// Set to the opposite of what the final result should be
			return flagChange == FlagFalse
		}
		// Otherwise choose true/false randomly
		return rand.Intn(2) == 0
	}
	cpu.SetFlag(Z, setFlagVal(flags.ZeroFlag))
	cpu.SetFlag(H, setFlagVal(flags.HalfCarryFlag))
	cpu.SetFlag(C, setFlagVal(flags.CarryFlag))
	cpu.SetFlag(N, setFlagVal(flags.SubtractFlag))
	return cpu.F.Retrieve()
}

// Check CPU flags after computation. Returns a list
func checkCpuFlags(t *testing.T, cpu *Cpu, flagChanges FlagChanges, originalFlags byte) {
	getOriginalFlag := func(flag int) bool {
		var index int
		switch flag {
		case Z:
			index = 7
		case N:
			index = 6
		case H:
			index = 5
		case C:
			index = 4
		default:
			panic(fmt.Sprintf("Unknown flag: %d",  flag))
		}
		return originalFlags & (1 << uint(index)) != 0x0
	}
	checkFlag := func(flag int, expectedFlagState FlagState) {
		flagVal := cpu.GetFlag(flag)
		if expectedFlagState == FlagUnchanged {
			if getOriginalFlag(flag) != flagVal {
				t.Errorf("Flag %s incorrectly modified, want: %t, got: %t",
					flag, getOriginalFlag(flag), flagVal)
			}
		} else {
			expectedFlagValue := expectedFlagState == FlagTrue
			if expectedFlagValue != flagVal {
				t.Errorf("Flag %s incorrect, want: %t, got: %t", FlagEnumToName(flag),
					expectedFlagValue, flagVal)
			}
		}
	}
	checkFlag(Z, flagChanges.ZeroFlag)
	checkFlag(N, flagChanges.SubtractFlag)
	checkFlag(H, flagChanges.HalfCarryFlag)
	checkFlag(C, flagChanges.CarryFlag)
}

func checkCpuState(t *testing.T, cpu *Cpu, mem *memory.Memory, expectedState SystemState) {
	if cpu.A.Retrieve() != expectedState.A {
		t.Errorf("Register A incorrect, want: 0x%x, got: 0x%x", expectedState.A, cpu.A.Retrieve())
	}
	if cpu.B.Retrieve() != expectedState.B {
		t.Errorf("Register B incorrect, want: 0x%x, got: 0x%x", expectedState.B, cpu.B.Retrieve())
	}
	if cpu.C.Retrieve() != expectedState.C {
		t.Errorf("Register C incorrect, want: 0x%x, got: 0x%x", expectedState.C, cpu.C.Retrieve())
	}
	if cpu.D.Retrieve() != expectedState.D {
		t.Errorf("Register D incorrect, want: 0x%x, got: 0x%x", expectedState.D, cpu.D.Retrieve())
	}
	if cpu.E.Retrieve() != expectedState.E {
		t.Errorf("Register E incorrect, want: 0x%x, got: 0x%x", expectedState.E, cpu.E.Retrieve())
	}
	if cpu.H.Retrieve() != expectedState.H {
		t.Errorf("Register H incorrect, want: 0x%x, got: 0x%x", expectedState.H, cpu.H.Retrieve())
	}
	if cpu.L.Retrieve() != expectedState.L {
		t.Errorf("Register L incorrect, want: 0x%x, got: 0x%x", expectedState.L, cpu.L.Retrieve())
	}

	// Only check F if it's explicitly set
	if expectedState.F != 0 && cpu.F.Retrieve() != expectedState.F {
		t.Errorf("Register F incorrect, want: 0x%x, got: 0x%x", expectedState.F, cpu.F.Retrieve())
	}
	
	if cpu.SP.Retrieve() != expectedState.SP {
		t.Errorf("Register SP incorrect, want: 0x%x, got: 0x%x", expectedState.SP, cpu.SP.Retrieve())
	}
	if cpu.PC.Retrieve() != expectedState.PC {
		t.Errorf("Register PC incorrect, want: 0x%x, got: 0x%x", expectedState.PC, cpu.PC.Retrieve())
	}
	for address, expectedVal := range expectedState.memVals {
		if memVal, err := mem.Get(address); err != nil {
			t.Errorf("Error fetching memory at %x: %v", address, err)
		} else if memVal != expectedVal {
			t.Errorf("Incorrect memory at %x, want: %x, got: %x", address, expectedVal,
				memVal)
		}
	}
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

func TestInc8BitRegOpCodes(t *testing.T) {
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

			opCode := &Inc8BitRegOpCode{
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

func TestIncMemOpCodes(t *testing.T) {
	testCases := []struct{
		name string
		memoryAddress types.Word
		initialMemoryValue byte
		expectedMemoryValue byte
		expectedZeroFlag bool
		expectedHalfCarry bool
	} {
		{
			name: "Regular memory increment",
			memoryAddress: types.Word(0x33),
			initialMemoryValue: byte(0x13),
			expectedMemoryValue: byte(0x14),
			expectedZeroFlag: false,
		},
		{
			name: "Zero memory",
			memoryAddress: types.Word(0x12),
			initialMemoryValue: byte(0xFF),
			expectedMemoryValue: byte(0x00),
			expectedZeroFlag: true,
			expectedHalfCarry: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cpu, memory := setupCpu()
			cpu.BC.Assign(tc.memoryAddress)
			if err := memory.Set(tc.memoryAddress, tc.initialMemoryValue); err != nil {
				t.Fatalf("Error setting memory value: %v", err)
			}

			opCode := &IncMemOpCode{
				r1: &cpu.BC,
			}

			cycles, pcModified, err := opCode.Run(cpu)

			if err != nil {
				t.Fatalf("Error runing opcode: %v", err)
			}

			if pcModified {
				t.Error("PC incorrectly modified, want: false, got: true")
			}

			if cycles != 12 {
				t.Errorf("Incorrect number of cycles, want: 12, got: %d", cycles)
			}

			r1Val := cpu.BC.Retrieve()
			if r1Val != tc.memoryAddress {
				t.Errorf("Incorrectly modified R1 val, want: %x, got: %x", tc.memoryAddress, r1Val)
			}

			memoryValue, err := memory.Get(tc.memoryAddress)
			if err != nil {
				t.Fatalf("Error fetching memory value: %v", err)
			}
			
			if memoryValue != tc.expectedMemoryValue {
				t.Errorf("Incorrect memory value, want: %x, got: %x", tc.expectedMemoryValue, memoryValue)
			}

			if cpu.GetFlag(Z) != tc.expectedZeroFlag {
				t.Errorf("Incorrect zero flag, want: %t, got: %t", tc.expectedZeroFlag, cpu.GetFlag(Z))
			}

			if cpu.GetFlag(H) != tc.expectedHalfCarry {
				t.Errorf("Incorrect half carry flag, want: %t, got: %t", tc.expectedHalfCarry, cpu.GetFlag(H))
			}

			if cpu.GetFlag(N) {
				t.Errorf("Incorrect subtract flag, want: false, got: true")
			}
		})
	}
}


func Test8BitDecOpCode(t *testing.T) {
	testCases := []struct{
		name string
		inputValue byte
		expectedValue byte
		expectedZeroFlag bool
	} {
		{
			name: "Normal decrement",
			inputValue: byte(0x06),
			expectedValue: byte(0x05),
			expectedZeroFlag: false,
		},
		{
			name: "Zero decrement",
			inputValue: byte(0x01),
			expectedValue: byte(0x00),
			expectedZeroFlag: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cpu, _ := setupCpu()
			cpu.A.Assign(tc.inputValue)

			opCode := &Dec8BitRegOpCode{
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

			if !cpu.GetFlag(N) {
				t.Errorf("Incorrect subtract flag, want: false, got: true")
			}
		})
	}
}

func TestGenerateOpCodes(t *testing.T) {
	cpu, _ := setupCpu()
	codes := cpu.codes

	if code, exists := codes[0x41]; !exists {
		t.Errorf("Could not find expected opcode")
	} else {
		if code.Name() != "LD B,C" {
			t.Errorf("Code name incorrect, want: 'LD B,C, got: %s", code.Name())
		}
		ldOpCode, ok := code.(*Ld8BitRegisterOpCode)
		if !ok {
			t.Fatalf("Opcode was of incorrect type, want: *Ld8BitRegisterOpCode, got: %T", ldOpCode)
		}

		if ldOpCode.r1 == nil {
			t.Errorf("Opcode R1 not set, want: B, got: nil")
		} else if ldOpCode.r1 != &cpu.B {
			t.Errorf("Incorrect R1 for opcode, want: B, got: %s", ldOpCode.r1.Name)
		}

		if ldOpCode.r1 == nil {
			t.Errorf("Opcode R2 not set, want: C, got: nil")
		} else if ldOpCode.r2 != &cpu.C {
			t.Errorf("Incorrect R2 for opcdoe, want: C, got: %s", ldOpCode.r2.Name)
		}
	}
}


// New style tests that get generated opcodes\
type FlagState int
const (
	FlagUnchanged FlagState = iota
	FlagFalse
	FlagTrue
)

// Helper struct for setting CPU state in tests
type SystemState struct {
	A, B, C, D, E, F, H, L byte
	SP, PC types.Word	
	memVals map[types.Word]byte
}

type FlagChanges struct {
	ZeroFlag, SubtractFlag, HalfCarryFlag, CarryFlag FlagState
}

func TestAdd8BitRegOpCodes(t *testing.T) {
	testCases := []struct{
		name string
		code string
		startState, endState SystemState
		flagChanges FlagChanges
	} {
		{
			name: "Simple add",
			code: "ADD A,E",
			startState: SystemState{
				A: byte(0x5),
				E: byte(0x3),
			},
			endState: SystemState{
				A: byte(0x8),
				E: byte(0x3),
			},
			flagChanges: FlagChanges{
				ZeroFlag: FlagFalse,
				SubtractFlag: FlagFalse,
				HalfCarryFlag: FlagFalse,
				CarryFlag: FlagFalse,
			},
		}, {
			name: "Add with half carry",
			code: "ADD A,L",
			startState: SystemState{
				A: byte(0xA),
				L: byte(0xD),
			},
			endState: SystemState{
				A: byte(0xA + 0xD),
				L: byte(0xD),
			},
			flagChanges: FlagChanges{
				ZeroFlag: FlagFalse,
				SubtractFlag: FlagFalse,
				HalfCarryFlag: FlagTrue,
				CarryFlag: FlagFalse,
			},
		}, {
			name: "Add with full and half carry",
			code: "ADD A,H",
			startState: SystemState{
				A: byte(0xFF),
				H: byte(0xFF),
			},
			endState: SystemState{
				A: byte(0xFE),
				H: byte(0xFF),
			},
			flagChanges: FlagChanges{
				ZeroFlag: FlagFalse,
				SubtractFlag: FlagFalse,
				HalfCarryFlag: FlagTrue,
				CarryFlag: FlagTrue,
			},
		}, {
			name: "ADC with carry",
			code: "ADC A,B",
			startState: SystemState{
				A: byte(0x4),
				B: byte(0x3),
				F: byte(0x1 << 4), // Set carry flag
			},
			endState: SystemState{
				A: byte(0x8),
				B: byte(0x3),
			},
			flagChanges: FlagChanges{
				ZeroFlag: FlagFalse,
				SubtractFlag: FlagFalse,
				HalfCarryFlag: FlagFalse,
				CarryFlag: FlagFalse,
			},
		},		
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cpu, mem := setupCpuWithState(tc.startState)
			
			originalFlagState := tc.startState.F
			// Only update the flags if F isn't explicitly set
			if originalFlagState == 0 {
				originalFlagState = setupCpuFlags(cpu, tc.flagChanges)
			}
			codesByName := OpCodesByName(cpu.codes)
			code, exists := codesByName[tc.code]
			if !exists {
				t.Fatalf("Could not find code '%s'", tc.code)
			}
			cycles, pcModified, err := code.Run(cpu)
			if err != nil {
				t.Fatalf("Error running opcode: %v", err)
			}
			if cycles != 4 {
				t.Errorf("Incorrect number of cycles, want: 4, got: %d", cycles)
			}
			if pcModified {
				t.Error("PC incorrectly modified, want: false, got: true")
			}
			checkCpuFlags(t, cpu, tc.flagChanges, originalFlagState)
			checkCpuState(t, cpu, mem, tc.endState)
		})
	}
};
