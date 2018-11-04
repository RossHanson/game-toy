package cpu

import (
	"memory"
	"log"
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
	codes map[byte]OpCode
}

type BaseOpCode struct {
	code byte
	length int
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

	cpu.codes = cpu.generateOpCodes()

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

func (c *Cpu) generateOpCodes() map[byte]OpCode {
	codes := make(map[byte]OpCode)
	// 8 bit ld op codes
	{
		// Order of source registers
		sourceRegisters := []*Register8Bit{
			&c.B, &c.C, &c.D, &c.H, &c.L, nil, &c.A}
		destRegisters := []*Register8Bit{
			&c.B, &c.C, &c.D, &c.E, &c.H, &c.L, &c.A}

		codeMsb := 0x04
		codeLsb := 0x00

		for index, destRegister := range destRegisters {
			for _, srcRegister := range sourceRegisters {
				code := byte(codeLsb + (codeMsb * 16))
				codeLsb++
				if srcRegister == nil || destRegister == nil {
					continue
				}
				op := &Ld8BitRegisterOpCode{
					BaseOpCode: BaseOpCode{
						code: code,
						length: 1,
					},
					r1: destRegister,
					r2: srcRegister,
				}
				codes[code] = op
			}
			if index % 2 == 1 {
				codeMsb++
			}
		}
	}
	// 8 bit immediate loads
	{
		// Map of code LSB to dest registers
		lsbToDestRegisters := map[int][]*Register8Bit{
			0x6: {&c.B, &c.D, &c.H},
			0xE: {&c.C, &c.E, &c.L, &c.A},
		}
		for codeLsb, destRegisters := range lsbToDestRegisters {
			codeMsb := 0x0
			for _, destRegister := range destRegisters {
				code := byte(codeLsb + (codeMsb * 16))
				codeMsb++
				op := &Ld8BitImmediateOpCode{
					BaseOpCode: BaseOpCode{
						code: code,
						length: 2,
					},
					r1: destRegister,
				}
				codes[code] = op
			}
		}
	}

	// Memory immediate load
	{
		code := byte(0x36)
		codes[code] = &LdMemoryImmediateOpCode{
			BaseOpCode: BaseOpCode {
				code: code,
				length: 2,
			},
			r1: &c.HL,
		}
	}

	// Reg into memory loads for BC and DE
	{
		srcRegisters := []*Register16Bit{&c.BC, &c.DE}
		codeLsb := 0x2
		codeMsb := 0x0
		for _, srcRegister := range srcRegisters {
			code := byte(codeLsb + (codeMsb * 16))
			codeMsb++
			codes[code] = &LdRegIntoMemOpCode{
				BaseOpCode: BaseOpCode{
					code: code,
					length: 1,
				},
				r1: srcRegister,
				r2: &c.A,
			}
		}
	}

	// Reg into memory loads for HL
	{
		srcRegisters := []*Register8Bit{&c.B, &c.C, &c.D, &c.E, &c.H, &c.L, nil, &c.A}
		codeLsb := 0x0
		codeMsb := 0x7
		for _, srcRegister := range srcRegisters {
			code := byte(codeLsb + (codeMsb * 16))
			codeLsb++
			if srcRegister == nil {
				continue
			}
			codes[code] = &LdRegIntoMemOpCode{
				BaseOpCode: BaseOpCode{
					code: code,
					length: 1,
				},
				r1: &c.HL,
				r2: srcRegister,
			}
		}
	}

	// Reg into memory loads for HL+
	{
		code := byte(0x22)
		codes[code] = &LdRegIntoMemOpCode{
			BaseOpCode: BaseOpCode{
				code: code,
				length: 1,
			},
			incrementR1: true,
			r1: &c.HL,
			r2: &c.A,
		}
	}

	// Reg into memory loads for HL-
	{
		code := byte(0x32)
		codes[code] = &LdRegIntoMemOpCode{
			BaseOpCode: BaseOpCode{
				code: code,
				length: 1,
			},
			decrementR1: true,
			r1: &c.HL,
			r2: &c.A,
		}
	}

	// Memory into reg for BC and DE
	{
		codeMsb := 0x0
		codeLsb := 0xA
		srcRegisters := []*Register16Bit{&c.BC, &c.DE}
		for _, srcRegister := range srcRegisters {
			code := byte(codeLsb + 16 * codeMsb)
			codeMsb++
			codes[code] = &LdMemIntoRegOpCode{
				BaseOpCode: BaseOpCode{
					code: code,
					length: 1,
				},
				r1: &c.A,
				r2: srcRegister,
			}
		}
	}

	// Memory into reg for regular HL
	{
		lsbToRegisters := map[int][]*Register8Bit{
			0x6: {&c.B, &c.D, &c.H},
			0xE: {&c.C, &c.E, &c.L, &c.A},
		}
		for codeLsb, destRegisters := range lsbToRegisters {
			codeMsb := 0x4
			for _, destRegister := range destRegisters {
				code := byte(codeLsb + 16 * codeMsb)
				codeMsb++
				codes[code] = &LdMemIntoRegOpCode{
					BaseOpCode: BaseOpCode{
						code: code,
						length: 1,
					},
					r1: destRegister,
					r2: &c.HL,
				}
			}
		}
	}

	// LD A,(HL+)
	{
		code := byte(0x2A)
		codes[code] = &LdMemIntoRegOpCode{
			BaseOpCode: BaseOpCode{
				code: code,
				length: 1,
			},
			r1: &c.A,
			r2: &c.HL,
			incrementR2: true,
		}
	}

	// LD A,(HL-)
	{
		code := byte(0x3A)
		codes[code] = &LdMemIntoRegOpCode{
			BaseOpCode: BaseOpCode{
				code: code,
				length: 1,
			},
			r1: &c.A,
			r2: &c.HL,
			decrementR2: true,
		}
	}

	
	return codes
}
