package cpu

import (
	"memory"
	"log"
	"encoding/binary"
	"utils"
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

type Register8Bit struct {
	Name string
	value byte
}

type Register16Bit struct {
	Name string
	lsb *byte
	msb *byte
}

func (r *Register8Bit) Assign(value byte) {
	log.Printf("Assigning value %x into %s", value, r.Name)
	r.value = value
}

func (r *Register8Bit) Retrieve() byte {
	return r.value
}

func (r *Register16Bit) Assign(lsb byte, msb byte) {
	*r.lsb = lsb
	*r.msb =  msb
}

func (r *Register16Bit) Retrieve() (lsb byte, msb byte) {
	return *r.lsb, *r.msb
}

func (r *Register8Bit) Increment() (zero bool, halfCarry bool) {
	// This could probably be more efficient
	//val := UInt8(r.value)
	//val++
	//r.value = byte(val)
	// TODO figure out return types
	return false, false
	//return val == 0, false
}

type Cpu struct {
	// more fields to come
	memory *memory.Memory
	A, B, C, D, E, F, H, L Register8Bit
	BC, DE, HL, SP, PC Register16Bit
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
	cpu := &Cpu{
		memory: memory,
	}

	cpu.A.Name = "A"
	cpu.B.Name = "B"
	cpu.C.Name = "C"
	cpu.D.Name = "D"
	cpu.E.Name = "E"
	cpu.F.Name = "F"

	setupMixedRegister := func(mixed *Register16Bit, lsb *Register8Bit, msb *Register8Bit) {
		mixed.lsb = &lsb.value
		mixed.msb = &msb.value
		mixed.Name = lsb.Name + msb.Name
	}

	setupMixedRegister(&cpu.BC, &cpu.B, &cpu.C)
	setupMixedRegister(&cpu.DE, &cpu.D, &cpu.E)
	setupMixedRegister(&cpu.HL, &cpu.H, &cpu.L)

	var spLsb, spMsb, pcLsb, pcMsb byte

	cpu.SP.lsb = &spLsb
	cpu.SP.msb = &spMsb
	cpu.PC.lsb = &pcLsb
	cpu.PC.msb = &pcMsb

	return cpu
}

func (c *Cpu) LoadImmediateData(length int) ([]byte, error) {
	data := make([]byte, length)
	pcValLsb, pcValMsb := c.PC.Retrieve()
	pc := binary.LittleEndian.Uint16([]byte{pcValLsb, pcValMsb})
	pc++ // skip 1 for the current instruction
	for offset := 0; offset < length; offset++ {
		lsb, msb := utils.EncodeInt(int(pc) + offset)
		memoryValue, err := c.memory.Get(lsb, msb)
		if err != nil {
			return nil, err
		}
		data[offset] = memoryValue
	}
	return data, nil

}
