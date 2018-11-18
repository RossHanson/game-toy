package cpu

import (
	"fmt"
	"utils"
)

type Modifier int
const (
	Unchanged Modifier = iota
	Increment
	Decrement
)

func ModifierToString(mod Modifier) string {
	switch mod {
	case Unchanged:
		return "Unchanged"
	case Increment:
		return "Increment"
	case Decrement:
		return "Decrement"
	default:
		return fmt.Sprintf("Unknown modifier: %d", mod)
	}
}

type Ld8BitRegisterOpCode struct {
	BaseOpCode
	r1 *Register8Bit
	r2 *Register8Bit
}

func (b *Ld8BitRegisterOpCode) Run(cpu *Cpu) (int, error) {
	b.r1.Assign(b.r2.Retrieve())
	cpu.IncrementPC(1)
	return 4, nil
}

func (b *Ld8BitRegisterOpCode) Name() string {
	return fmt.Sprintf("LD %s,%s", b.r1.Name, b.r2.Name)
}

type LdRegIntoMemOpCode struct {
	BaseOpCode
	r1 *Register16Bit
	r2 *Register8Bit
	incrementR1 bool
	decrementR1 bool
}

func (b *LdRegIntoMemOpCode) Run(cpu *Cpu) (int, error) {
	dest := b.r1.Retrieve()
	cpu.memory.Set(dest, b.r2.Retrieve())
	if b.incrementR1 {
		b.r1.Increment()
	}
	if b.decrementR1 {
		b.r1.Decrement()
	}
	cpu.IncrementPC(1)
	return 8, nil
}

func (b *LdRegIntoMemOpCode) Name() string {
	modifier := ""
	if b.incrementR1 {
		modifier = "+"
	} else if b.decrementR1 {
		modifier = "-"
	}
	return fmt.Sprintf("LD (%s%s),%s", b.r1.Name, modifier,
		b.r2.Name)
}

type LdMemIntoRegOpCode struct {
	BaseOpCode
	r1 *Register8Bit
	r2 *Register16Bit
	incrementR2 bool
	decrementR2 bool
}

func (b *LdMemIntoRegOpCode) Run(cpu *Cpu) (int, error) {
	src := b.r2.Retrieve()
	val, err := cpu.memory.Get(src)
	if err != nil {
		return -1, err
	}

	if b.incrementR2 {
		b.r2.Increment()
	}
	if b.decrementR2 {
		b.r2.Decrement()
	}

	b.r1.Assign(val)
	cpu.IncrementPC(1)
	return 8, nil
}

func (b *LdMemIntoRegOpCode) Name() string {
	modifier := ""
	if b.incrementR2 {
		modifier = "+"
	} else if b.decrementR2 {
		modifier = "-"
	}
	return fmt.Sprintf("LD %s,(%s%s)", b.r1.Name, b.r2.Name, modifier)
}

type Ld8BitImmediateOpCode struct {
	BaseOpCode
	r1 *Register8Bit
}

func (b *Ld8BitImmediateOpCode) Run(cpu *Cpu) (int, error) {
	immediateByte, err := cpu.LoadImmediateByte()
	if err != nil {
		return -1, err
	}
	b.r1.Assign(immediateByte)
	cpu.IncrementPC(2)
	return 8, nil
}

func (b *Ld8BitImmediateOpCode) Name() string {
	return fmt.Sprintf("LD %s,d8", b.r1.Name)
}

type Ld16BitImmediateOpCode struct {
	BaseOpCode
	r1 *Register16Bit
}

func (b *Ld16BitImmediateOpCode) Run(cpu *Cpu) (int, error) {
	immediateData, err := cpu.LoadImmediateWord()
	if err != nil {
		return -1, err
	}
	b.r1.Assign(immediateData)
	cpu.IncrementPC(3)
	return 12, nil
}

func (b *Ld16BitImmediateOpCode) Name() string {
	return fmt.Sprintf("LD %s,d16", b.r1.Name)
}

type LdMemoryImmediateOpCode struct {
	BaseOpCode
	r1 *Register16Bit // This is always HL?
}

func (b *LdMemoryImmediateOpCode) Run(cpu *Cpu) (int, error) {
	immediateData, err := cpu.LoadImmediateByte()
	if err != nil {
		return -1, err
	}
	targetAddress := b.r1.Retrieve()
	if err := cpu.memory.Set(targetAddress, immediateData); err != nil {
		return -1, err
	}
	cpu.IncrementPC(2)
	return 12, nil
}

func (b *LdMemoryImmediateOpCode) Name() string {
	return fmt.Sprintf("LD (%s),d8", b.r1.Name)
}

type Inc8BitRegOpCode struct {
	BaseOpCode
	r1 ByteSource
}

func (b *Inc8BitRegOpCode) Run(cpu *Cpu) (int, error) {
	zero, halfCarry, cycles:= b.r1.IncrementValue(cpu.memory)
	cpu.SetFlag(Z, zero)
	cpu.SetFlag(H, halfCarry)
	cpu.SetFlag(N, false)
	cpu.IncrementPC(1)
	return 4 + cycles, nil
}

func (b *Inc8BitRegOpCode) Name() string {
	return fmt.Sprintf("INC %s", b.r1.PrintableName())
}

type IncMemOpCode struct {
	BaseOpCode
	r1 *Register16Bit
}

func (b *IncMemOpCode) Run(cpu *Cpu) (int, error) {
	val, err := cpu.memory.Get(b.r1.Retrieve())
	if err != nil {
		return -1, err
	}
	
	incResults := utils.Add8Bit(val, 0x1)
	cpu.SetFlag(Z, incResults.Zero)
	cpu.SetFlag(H, incResults.HalfCarry)
	cpu.SetFlag(N, false)
	if err := cpu.memory.Set(b.r1.Retrieve(), incResults.Result); err != nil {
		return -1, err
	}
	cpu.IncrementPC(1)
	return 12, nil
}

func (b *IncMemOpCode) Name() string {
	return fmt.Sprintf("INC (%s)", b.r1.Name)
}

type Dec8BitRegOpCode struct {
	BaseOpCode
	r1 ByteSource
}

func (b *Dec8BitRegOpCode) Run(cpu *Cpu) (int, error) {
	zero, halfCarry, cycles := b.r1.DecrementValue(cpu.memory)
	cpu.SetFlag(Z, zero)
	cpu.SetFlag(H, halfCarry)
	cpu.SetFlag(N, true)
	cpu.IncrementPC(1)
	return 4 + cycles, nil
}

func (b *Dec8BitRegOpCode) Name() string {
	return fmt.Sprintf("DEC %s", b.r1.PrintableName())
}

type Add8BitRegOpCode struct {
	BaseOpCode
	r1 *Register8Bit // Should always be A
	r2 ByteSource
	includeCarry bool
}

func (b *Add8BitRegOpCode) Run(cpu *Cpu) (int, error) {
	r2Val, cycles := b.r2.GetValue(cpu.memory)
	result := utils.Add8BitWithCarry(b.r1.Retrieve(), r2Val,
		b.includeCarry && cpu.GetFlag(C))
	b.r1.Assign(result.Result)
	cpu.SetFlag(Z, result.Zero)
	cpu.SetFlag(H, result.HalfCarry)
	cpu.SetFlag(C, result.Carry)
	cpu.SetFlag(N, false)
	cpu.IncrementPC(1)
	return 4 + cycles, nil
}

func (b *Add8BitRegOpCode) Name() string {
	base := "ADD"
	if b.includeCarry {
		base = "ADC"
	}
	return fmt.Sprintf("%s %s,%s", base, b.r1.Name, b.r2.PrintableName())
}

type Add16BitRegOpCode struct {
	BaseOpCode
	r1 *Register16Bit
	r2 *Register16Bit
}

func (b *Add16BitRegOpCode) Run(cpu *Cpu) (int, error) {
	result := utils.Add16Bit(b.r1.Retrieve(), b.r2.Retrieve())
	b.r1.Assign(result.Result)
	cpu.SetFlag(H, result.HalfCarry)
	cpu.SetFlag(C, result.Carry)
	cpu.SetFlag(N, false)
	cpu.IncrementPC(1)
	return 4, nil
}

func (b *Add16BitRegOpCode) Name() string {
	return fmt.Sprintf("ADD %s,%s", b.r1.Name, b.r2.Name)
}

type Sub8BitRegOpCode struct {
	BaseOpCode
	r1 *Register8Bit
	r2 ByteSource
	includeCarry bool
}

func (b *Sub8BitRegOpCode) Run(cpu *Cpu) (int, error) {
	r2Val, cycles := b.r2.GetValue(cpu.memory)
	result := utils.Subtract8BitWithCarry(b.r1.Retrieve(), r2Val,
		b.includeCarry && cpu.GetFlag(C))
	b.r1.Assign(result.Result)
	cpu.SetFlag(Z, result.Zero)
	cpu.SetFlag(H, result.HalfCarry)
	cpu.SetFlag(C, result.Carry)
	cpu.SetFlag(N, true)
	cpu.IncrementPC(1)
	return 4 + cycles, nil
}

func (b *Sub8BitRegOpCode) Name() string {
	if !b.includeCarry {
		return fmt.Sprintf("SUB %s", b.r2.PrintableName())
	}
	return fmt.Sprintf("SBC A,%s", b.r2.PrintableName())
}

// INC n & DEC n for 16 bit
type IncDec16Bit struct {
	BaseOpCode
	target *Register16Bit
	mod Modifier
}

func (b *IncDec16Bit) Run(cpu *Cpu) (int, error) {
	switch b.mod {
	case Increment:
		b.target.Increment()
	case Decrement:
		b.target.Decrement()
	default:
		return -1, fmt.Errorf("Bad modifier: %d", b.mod)
	}

	cpu.IncrementPC(1)
	return 8, nil
}

func (b *IncDec16Bit) Name() string {
	base := "INC"
	if b.mod == Decrement {
		base = "DEC"
	}
	return fmt.Sprintf("%s %s", base, b.target.Name)
}

type NoOpCode struct {
	BaseOpCode
}

func (b *NoOpCode) Run(cpu *Cpu) (int, error) {
	return 4, nil
}

func (b *NoOpCode) Name() string {
	return "NOP"
}

// AND, XOR, OR, CP

type LogicalOp int
const (
	AND LogicalOp = iota
	XOR
	OR
	CP
)

type Logical8BitOp struct {
	BaseOpCode
	target *Register8Bit // Should always be A
	source ByteSource
	operation LogicalOp
}

func (b *Logical8BitOp) Run(cpu *Cpu) (int, error) {
	targetVal := b.target.Retrieve()
	sourceVal, sourceCycles := b.source.GetValue(cpu.memory)
	switch b.operation {
	case AND:
		andedVal := targetVal & sourceVal
		cpu.SetFlag(Z, andedVal == 0)
		cpu.SetFlag(N, false)
		cpu.SetFlag(H, true)
		cpu.SetFlag(C, false)
		b.target.Assign(andedVal)
	case XOR:
		xorVal := targetVal ^ sourceVal
		cpu.SetFlag(Z, xorVal == 0)
		cpu.SetFlag(N, false)
		cpu.SetFlag(H, false)
		cpu.SetFlag(C, false)
		b.target.Assign(xorVal)
	case OR:
		orVal := targetVal | sourceVal
		cpu.SetFlag(Z, orVal == 0)
		cpu.SetFlag(N, false)
		cpu.SetFlag(H, false)
		cpu.SetFlag(C, false)
	case CP:
		subtractResults := utils.Subtract8Bit(targetVal, sourceVal)
		cpu.SetFlag(Z, subtractResults.Result == 0)
		cpu.SetFlag(N, true)
		cpu.SetFlag(H, subtractResults.HalfCarry)
		cpu.SetFlag(C, subtractResults.Carry)
	default:
		return -1, fmt.Errorf("Unknown operation: %d", b.operation)
	}
	cpu.IncrementPC(1)
	return 4 + sourceCycles, nil
}

func (b *Logical8BitOp) Name() string {
	switch b.operation {
	case AND:
		return fmt.Sprintf("AND %s", b.source.PrintableName())
	case XOR:
		return fmt.Sprintf("XOR %s", b.source.PrintableName())
	case OR:
		return fmt.Sprintf("OR %s", b.source.PrintableName())
	case CP:
		return fmt.Sprintf("CP %s", b.source.PrintableName())
	default:
		return fmt.Sprintf("Unknown operation: %s", b.operation)
	}
}

type Direction int
const (
	Left Direction = iota
	Right
)

type RotateOpCode struct {
	BaseOpCode
	r1 ByteSource
	direction Direction
	includeCarry bool
	isCB bool
}

func (b *RotateOpCode) Run(cpu *Cpu) (int, error) {
	var bitVal bool
	sourceValue, cycles := b.r1.GetValue(cpu.memory)

	var calculation byte
	if b.direction == Left {
		bitVal = 0x80 & sourceValue == 0x80
		calculation = sourceValue << 1
	} else {
		bitVal = 0x01 & sourceValue == 0x01
		calculation = sourceValue >> 1
	}

	if b.includeCarry && cpu.GetFlag(C) {
		if b.direction == Left {
			calculation ^= 0x80
		} else {
			calculation ^= 0x01
		}
	}

	cpu.SetFlag(C, bitVal)
	cycles += b.r1.SetValue(cpu.memory, calculation)
	
	if !b.isCB {
		cpu.SetFlag(Z, false)
	}
	cpu.SetFlag(H, false)
	cpu.SetFlag(N, false)

	return 4 + cycles, nil
}

func (b *RotateOpCode) Name() string {
	dir := "L"
	if b.direction == Right {
		dir = "R"
	}
	name := "R" + dir

	if b.includeCarry {
		name += "C"
	}
	
	if b.isCB {
		name += " "
	}

	name += b.r1.PrintableName()
	return name
}
