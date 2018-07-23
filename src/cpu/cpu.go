package cpu

import (
	"fmt"
	"memory"
	"log"
	"encoding/binary"
	"utils"
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
	value []*byte
}

func (r *Register) Assign(value ...byte) error {
	if len(r.value) != len(value) {
		return fmt.Errorf("Attempted to assign a %d bit value to a %d bit register", 8*len(value), 8*len(r.value))
	}
	log.Printf("Assigning value %x into %s", value, r.Name)
	for index, val := range value {
		*r.value[index] = val
	}
	return nil
}

func (r *Register) Retrieve() []byte {
	result := make([]byte, len(r.value))
	for index, val := range r.value {
		result[index] = *val
	}
	return result
}

func (r *Register) is8Bit() bool {
	return len(r.value) == 1
}

type Cpu struct {
	// more fields to come
	memory *memory.Memory
	registers map[RegisterName]*Register
	programCounter *Register
	stackPointer *Register
	flagRegister *Register
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


// Initializes a default CPU
func NewCpu(memory *memory.Memory) (*Cpu) {
	registerMap := make(map[RegisterName]*Register)
	for _, registerName := range registers8Bit {
		var value byte
		registerMap[registerName] = &Register{
			Name: registerName,
			value: []*byte{&value},
		}
	}

	for _, registerName := range registers16Bit {
		var value1 byte
		var value2 byte
		registerMap[registerName] = &Register{
			Name: registerName,
			value: []*byte{&value1, &value2},
		}
	}

	for _, registerName := range combinationRegisters {
		register1 := registerMap[RegisterName(registerName[0])]
		register2 := registerMap[RegisterName(registerName[1])]
		registerMap[registerName] = &Register{
			Name: registerName,
			value: []*byte{register1.value[0], register2.value[0]},
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



func (c *Cpu) LoadImmediateData(length int) ([]byte, error) {
	data := make([]byte, length)
	pc := binary.LittleEndian.Uint16(c.programCounter.Retrieve())
	pc++ // skip 1 for the current instruction
	for offset := 0; offset < length; offset++ {
		memoryValue, err := c.memory.Get(utils.EncodeInt(int(pc) + offset))
		if err != nil {
			return nil, err
		}
		data[offset] = memoryValue
	}
	return data, nil

}
