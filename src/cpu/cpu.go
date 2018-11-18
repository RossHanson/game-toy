package cpu

import (
	"memory"
	"log"
	"types"
	"fmt"
	"utils"
	"text/tabwriter"
	"os"
)

const (
	_ = iota
	Z
	N
	H
	C
)

func FlagEnumToName(flag int) string {
	switch flag {
	case Z:
		return "Z"
	case N:
		return "N"
	case H:
		return "H"
	case C:
		return "C"
	default:
		return fmt.Sprintf("Unknown: %d", flag)
	}
}

type OpCode interface {
	// Execute the operation and return the number of cycles consumed,
	// or an error if one occurs.
	Run(cpu *Cpu) (cycles int, err error)
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
	r.value = value
}

func (r *Register8Bit) Retrieve() byte {
	return r.value
}

func (r *Register8Bit) Increment() (zero bool, halfCarry bool) {
	result := utils.Add8Bit(r.value, 0x1)
	r.value = result.Result
	return result.Zero, result.HalfCarry
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


func (r *Register8Bit) IncrementValue(*memory.Memory) (bool, bool, int) {
	zero, halfCarry := r.Increment()
	return zero, halfCarry, 0
}

func (r *Register8Bit) DecrementValue(*memory.Memory) (bool, bool, int) {
	zero, halfCarry := r.Decrement()
	return zero, halfCarry, 0
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

func (r *Register16Bit) IncrementValue(mem *memory.Memory) (zero bool, halfCarry bool, cycles int) {
	val, _ := mem.Get(r.Retrieve())
	result := utils.Add8Bit(val, byte(1))
	mem.Set(r.Retrieve(), result.Result)
	return result.Zero, result.HalfCarry, 8
}

func (r *Register16Bit) DecrementValue(mem *memory.Memory) (zero bool, halfCarry bool, cycles int) {
	val, _ := mem.Get(r.Retrieve())
	result := utils.Subtract8Bit(val, byte(1))
	mem.Set(r.Retrieve(), result.Result)
	return result.Zero, result.HalfCarry, 8
}

// Wrapper that lets us use either an 8 bit register directly or a 16bit one as an address
type ByteSource interface {
	GetValue(mem *memory.Memory) (value byte, cycles int)
	// Annoyingly need to use something besides Name() because both registers already have that field
	PrintableName() string
	SetValue(mem *memory.Memory, value byte) (cycles int)
	IncrementValue(mem *memory.Memory) (zero bool, halfCarry bool, cycles int)
	DecrementValue(mem *memory.Memory) (zero bool, halfCarry bool, cycles int)
}

func (r *Register8Bit) GetValue(*memory.Memory) (byte, int) {
	return r.value, 0 // No extra cycle cost
}

func (r *Register8Bit) SetValue(_ *memory.Memory, val byte) (int) {
	r.Assign(val)
	return 0
}

func (r *Register8Bit) PrintableName() string {
	return r.Name
}

func (r *Register16Bit) GetValue(mem *memory.Memory) (byte, int) {
	// TODO handle errors?
	val, _ := mem.Get(r.Retrieve())
	return val, 4
}

func (r *Register16Bit) SetValue(mem *memory.Memory, value byte) (int) {
	mem.Set(r.Retrieve(), value)
	return 4
}

func (r *Register16Bit) PrintableName() string {
	return fmt.Sprintf("(%s)", r.Name)
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
	cpu.H.Name = "H"
	cpu.L.Name = "L"

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

func (c *Cpu) IncrementPC(instructionSize int) {
	// Probably a better way to do this
	newVal := c.PC.Retrieve() + types.Word(instructionSize)
	c.PC.Assign(newVal)
	// TODO: check if we're wrapping around?
}

func (c *Cpu) generateOpCodes() map[byte]OpCode {
	codes := make(map[byte]OpCode)
	// LD n,n
	{
		// Order of source registers
		sourceRegisters := []*Register8Bit{
			&c.B, &c.C, &c.D, &c.E, &c.H, &c.L, nil, &c.A}
		destRegisters := []*Register8Bit{
			&c.B, &c.C, &c.D, &c.E, &c.H, &c.L, nil, &c.A}

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
				codeLsb = 0
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

	// ADD A,n
	{
		otherRegs := []ByteSource{&c.B, &c.C, &c.D, &c.E, &c.H, &c.L, &c.HL, &c.A}
		codeMsb := 0x8
		codeLsb := 0x0
		for _, otherReg := range otherRegs {
			code := byte(codeLsb + 16*codeMsb)
			codeLsb++
			codes[code] = &Add8BitRegOpCode{
				BaseOpCode: BaseOpCode{
					code: code,
					length: 1,
				},
				r1: &c.A,
				r2: otherReg,
			}
		}
	}

	// SUB n, SBC n
	{
		otherRegs := []ByteSource{&c.B, &c.C, &c.D, &c.E, &c.H, &c.L, &c.HL, &c.A}
		codeMsb := 0x9
		codeLsb := 0x0
		for offset, otherReg := range otherRegs {
			codeNoCarry := byte(codeLsb + codeMsb * 16 + offset)
			codeWithCarry := byte(codeLsb + codeMsb * 16 + len(otherRegs) + offset)
			codes[codeNoCarry] = &Sub8BitRegOpCode{
				BaseOpCode: BaseOpCode{
					code: codeNoCarry,
				},
				r1: &c.A,
				r2: otherReg,
				includeCarry: false,
			}
			codes[codeWithCarry] = &Sub8BitRegOpCode{
				BaseOpCode: BaseOpCode{
					code: codeWithCarry,
				},
				r1: &c.A,
				r2: otherReg,
				includeCarry: true,
			}
		}
	}


	// ADC A,n
	{
		otherRegs := []ByteSource{&c.B, &c.C, &c.D, &c.E, &c.H, &c.L, &c.HL, &c.A}
		codeMsb := 0x8
		codeLsb := 0x8
		for _, otherReg := range otherRegs {
			code := byte(codeLsb + 16*codeMsb)
			codeLsb++
			if otherReg == nil {
				continue
			}
			codes[code] = &Add8BitRegOpCode{
				BaseOpCode: BaseOpCode{
					code: code,
					length: 1,
				},
				r1: &c.A,
				r2: otherReg,
				includeCarry: true,
			}
		}
	}

	// Logical ops, AND, XOR, OR, and CP
	{
		buildOp := func(code byte, sourceReg ByteSource, op LogicalOp) *Logical8BitOp {
			return &Logical8BitOp{
				BaseOpCode: BaseOpCode{
					code: code,
					length: 1,
				},
				target: &c.A,
				source: sourceReg,
				operation: op,
			}
		}
		sourceRegs := []ByteSource{&c.B, &c.C, &c.D, &c.E, &c.H, &c.L, &c.HL, &c.A}
		codeMsb := 0xA
		codeLsb := 0x0
		for _, sourceReg := range sourceRegs {
			code := byte(codeLsb + 16*codeMsb)
			codeLsb++
			codes[code] = buildOp(code, sourceReg, AND)
		}

		for _, sourceReg := range sourceRegs {
			code := byte(codeLsb + 16 * codeMsb)
			codeLsb++
			codes[code] = buildOp(code, sourceReg, XOR)
		}

		for _, sourceReg := range sourceRegs {
			code := byte(codeLsb + 16 * codeMsb)
			codeLsb++
			codes[code] = buildOp(code, sourceReg, OR)
		}

		for _, sourceReg := range sourceRegs {
			code := byte(codeLsb + 16 * codeMsb)
			codeLsb++
			codes[code] = buildOp(code, sourceReg, CP)
		}
	}

	// LD n,d16
	{
		sourceRegs := []*Register16Bit{&c.BC, &c.DE, &c.HL, &c.SP}
		codeLsb := 0x1
		for codeMsb, sourceReg := range sourceRegs {
			code := byte(codeLsb + 16 * codeMsb)
			codes[code] = &Ld16BitImmediateOpCode{
				BaseOpCode: BaseOpCode{
					code: code,
				},
				r1: sourceReg,
			}
		}
	}

	// INC and DEC 16 bit
	{
		regs := []*Register16Bit{&c.BC, &c.DE, &c.HL, &c.SP}
		incLsb := 0x3
		decLsb := 0xB
		for codeMsb, reg := range regs {
			incCode := byte(incLsb + 16 * codeMsb)
			decCode := byte(decLsb + 16 * codeMsb)
			codes[incCode] = &IncDec16Bit{
				BaseOpCode: BaseOpCode{
					code: incCode,
				},
				target: reg,
				mod: Increment,
			}
			codes[decCode] = &IncDec16Bit{
				BaseOpCode: BaseOpCode{
					code: decCode,
				},
				target: reg,
				mod: Decrement,
			}
		}
	}

	{
		lsbToRegs := map[int][]ByteSource{
			0x4: []ByteSource{&c.B, &c.D, &c.H, &c.HL},
			0xC: []ByteSource{&c.C, &c.E, &c.L, &c.A},
		}
		// DEC regs LSB is +1 from INC
		for codeLsb, regs := range lsbToRegs {
			for codeMsb, reg := range regs {
				incCode := byte(codeLsb + 16 * codeMsb)
				decCode := byte(codeLsb + 1 + 16 * codeMsb)
				codes[incCode] = &Inc8BitRegOpCode{
					BaseOpCode: BaseOpCode{
						code: incCode,
					},
					r1: reg,
				}
				codes[decCode] = &Dec8BitRegOpCode{
					BaseOpCode: BaseOpCode{
						code: decCode,
					},
					r1: reg,
				}
			}
		}
	}

	// NOP
	{
		code := byte(0x00)
		codes[code] = &NoOpCode{
			BaseOpCode: BaseOpCode{
				code: code,
			},
		}
	}

	// RL{C}A and RR{C}A
	{
		makeOp := func(code byte, dir Direction, carry bool) {
			codes[code] = &RotateOpCode{
				BaseOpCode: BaseOpCode{
					code: code,
				},
				r1: &c.A,
				direction: dir,
				includeCarry: carry,
				isCB: false,
			}
		}
		makeOp(0x07, Left, true)
		makeOp(0x17, Left, false)
		makeOp(0x0F, Right, true)
		makeOp(0x1F, Right, false)
	}

	// ADD nn,nn for 16 bit
	{
		sourceRegisters := []*Register16Bit{&c.BC, &c.DE, &c.HL, &c.SP}
		codeLsb := 0x9
		for codeMsb, reg := range sourceRegisters {
			code := byte(codeMsb * 16 + codeLsb)
			codes[code] = &Add16BitRegOpCode{
				BaseOpCode: BaseOpCode{
					code: code,
				},
				r1: &c.HL,
				r2: reg,
			}
		}
	}
	

	return codes
}


// Utility function
func OpCodesByName(codes map[byte]OpCode) map[string]OpCode {
	codesByName := make(map[string]OpCode)
	for _, code := range codes {
		codesByName[code.Name()] = code
	}
	return codesByName
}

func (c *Cpu) PrintKnownOpCodes() {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.Debug)
	for msb := 0; msb < 16; msb++ {
		for lsb := 0; lsb < 16; lsb++ {
			index := byte(16*msb + lsb)
			if code, exists := c.codes[index]; exists {
				fmt.Fprintf(w, "0x%02x - %s", index, code.Name())
			} else {
				fmt.Fprintf(w, "0x%02x - Missing", index)
			}
			if lsb != 15 {
				fmt.Fprintf(w, "\t")
			}
		}
		fmt.Fprintf(w, "\n")
	}
	w.Flush()
}
