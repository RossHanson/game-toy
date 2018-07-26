package cpu

import (
	"memory"
	"log"
	"fmt"
	"types"
)

const (
	_ = iota
	Z
	N
	H
	C
)

type OpCode interface {
	// Execute the operation and return the number of cycles consumed,
	// or an error if one occurs.
	Run(cpu *Cpu) (cycles int, pcModified bool, err error)
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

func (r *Register8Bit) Increment() (zero bool, halfCarry bool) {
	calc := r.value + 0x01

	zeroFlag := calc == 0
	// Stole this from gnomeboycolor, not sure what it does tbh
	halfCarryFlag := (calc^0x01^r.value)&0x10 == 0x10

	r.value = calc
	return zeroFlag, halfCarryFlag
}

func (r *Register8Bit) Decrement() (zero bool, halfCarry bool) {
	calc := r.value - 0x01

	zeroFlag := calc == 0
	halfCarryFlag := (calc^0x01^r.value)&0x10 == 0x10

	r.value = calc
	return zeroFlag, halfCarryFlag
}

func (r *Register8Bit) SetBit(bit byte, value bool) {
	if uint(bit) > 7 {
		log.Fatalf("Set bit maximum is 7! Got: %d", uint(bit))
	}
	if value {
		r.value |= (1 << uint(bit))
	} else {
		r.value &= ^(1 << uint(bit))
	}
}

func (r *Register8Bit) GetBit(bit byte) bool {
	if uint(bit) > 7 {
		log.Fatalf("Get bit maximum is 7! Got: %d", uint(bit))
	}
	return r.value & (1 << uint(bit)) != 0x00
}

func (r *Register16Bit) Assign(val types.Word) {
	lsb, msb := val.ToBytes()
	*r.lsb = lsb
	*r.msb =  msb
}

func (r *Register16Bit) Retrieve() (types.Word) {
	return types.WordFromBytes(*r.lsb, *r.msb)
}

func (r *Register16Bit) Increment() {
	newVal := types.WordFromBytes(*r.lsb, *r.msb) + types.Word(1)
	lsb, msb := newVal.ToBytes()
	*r.lsb = lsb
	*r.msb = msb
}

func (r *Register16Bit) Decrement() {
	newVal := types.WordFromBytes(*r.lsb, *r.msb) - types.Word(1)
	lsb, msb := newVal.ToBytes()
	*r.lsb = lsb
	*r.msb = msb
}

type Cpu struct {
	// more fields to come
	memory *memory.Memory
	A, B, C, D, E, F, H, L Register8Bit
	AF, BC, DE, HL, SP, PC Register16Bit
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

	setupMixedRegister(&cpu.AF, &cpu.A, &cpu.F)
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

func (c *Cpu) LoadImmediateByte() (byte, error) {
	pc := types.Word(c.PC.Retrieve())
	return c.memory.Get(pc + types.Word(1))
}

func (c *Cpu) LoadImmediateWord() (types.Word, error) {
	pc := types.Word(c.PC.Retrieve())
	lsb, err := c.memory.Get(pc + types.Word(1))
	if err != nil {
		return types.Word(0), err
	}
	msb, err := c.memory.Get(pc + types.Word(2))
	if err != nil {
		return types.Word(0), err
	}
	return types.WordFromBytes(lsb, msb), nil
}

func (c *Cpu) SetFlag(flag int, value bool) {
	switch flag {
	case Z:
		c.F.SetBit(7, value)
	case N:
		c.F.SetBit(6, value)
	case H:
		c.F.SetBit(5, value)
	case C:
		c.F.SetBit(4, value)
	default:
		log.Fatalf("Unknown flag: %c", flag)
	}
}

func (c *Cpu) GetFlag(flag int) bool {
	switch flag {
	case Z:
		return c.F.GetBit(7)
	case N:
		return c.F.GetBit(6)
	case H:
		return c.F.GetBit(5)
	case C:
		return c.F.GetBit(4)
	default:
		log.Fatalf("Unknown flag: %c", flag)
		return false // think I need this for the compiler
	}
}
