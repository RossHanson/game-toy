package gametoy

import (
	"fmt"
)

type RegisterName string

var (
	registers8Bit = []RegisterName{"A", "B", "C", "D", "E", "F", "H", "L"}
	combinationRegisters = []RegisterName{"AF", "BC", "DE", "HL"}
	registers16Bit = []RegisterName{"PC", "SP", "FR"}
)

type OpCode interface {
	// Execute the operation and return the number of cycles consumed,
	// or an error if one occurs.
	Run(cpu *Cpu) (int, error)
	Name() string
	DebugString() string
	Cycles() int
	Length() int
}

type Register struct {
	Name RegisterName
	Value []*byte
}

func (r *Register) is8Bit() bool {
	return len(r.Value) == 1
}

type Cpu struct {
	// more fields to come
	memory *Memory
	registers map[RegisterName]Register
	programCounter Register
	stackPointer Register
	flagRegister Register
}

type BaseOpCode struct {
	name string
	code byte
	length int
}

func (b BaseOpCode) Name() string {
	return b.name
}

func (b BaseOpCode) DebugString() string {
	return fmt.Sprintf("%x - %s", b.code, b.name)
}

func (b BaseOpCode) Length() int {
	return b.length
}

type LdRegisterOpCode struct {
	BaseOpCode
	r1 RegisterName
	r2 RegisterName
	// If one of the registers is 16 bit, it's assumed to be a load from memory
}

func (b LdRegisterOpCode) Run(cpu *Cpu) (int, error) {
	r1 := cpu.registers[b.r1]
	r2 := cpu.registers[b.r2]
	if r1.is8Bit() && r2.is8Bit() {
		*r1.Value[0] = *r2.Value[0]
		return 4, nil
	}
	if !r1.is8Bit() && !r2.is8Bit() {
		return 0, fmt.Errorf("This op code is undefined")
	}
	if r1.is8Bit() {
		// Load the value in memory from r2
		memoryValue, err := cpu.memory.Get(
			[]byte{*r2.Value[0], *r2.Value[1]})
		if err != nil {
			*r1.Value[0] = memoryValue
		}
		return 8, err
	}
	// Set the memory at r1 to the value from r2
	err := cpu.memory.Set([]byte{*r1.Value[0], *r1.Value[1]}, *r2.Value[0])
	return 8, err
}

// Initializes a default CPU
func newCpu(memory *Memory) (*Cpu) {
	registerMap := make(map[RegisterName]Register)
	for _, registerName := range registers8Bit {
		var value byte
		registerMap[registerName] = Register{
			Name: registerName,
			Value: []*byte{&value},
		}
	}

	for _, registerName := range registers16Bit {
		var value1 byte
		var value2 byte
		registerMap[registerName] = Register{
			Name: registerName,
			Value: []*byte{&value1, &value2},
		}
	}

	for _, registerName := range combinationRegisters {
		register1 := registerMap[RegisterName(registerName[0])]
		register2 := registerMap[RegisterName(registerName[1])]
		registerMap[registerName] = Register{
			Name: registerName,
			Value: []*byte{register1.Value[0], register2.Value[0]},
		}
	}

	return &Cpu{
		memory: memory,
		registers: registerMap,
		programCounter: registerMap["PC"],
		stackPointer: registerMap["SP"],
		flagRegister: registerMap["FR"],
	}
}
