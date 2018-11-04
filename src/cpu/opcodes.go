package cpu

import (
	"fmt"
	"utils"
)

type Ld8BitRegisterOpCode struct {
	BaseOpCode
	r1 *Register8Bit
	r2 *Register8Bit
}

func (b *Ld8BitRegisterOpCode) Run(cpu *Cpu) (int, bool, error) {
	b.r1.Assign(b.r2.Retrieve())
	return 4, false, nil
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

func (b *LdRegIntoMemOpCode) Run(cpu *Cpu) (int, bool, error) {
	dest := b.r1.Retrieve()
	cpu.memory.Set(dest, b.r2.Retrieve())
	if b.incrementR1 {
		b.r1.Increment()
	}
	if b.decrementR1 {
		b.r1.Decrement()
	}
	return 8, false, nil
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

func (b *LdMemIntoRegOpCode) Run(cpu *Cpu) (int, bool, error) {
	src := b.r2.Retrieve()
	val, err := cpu.memory.Get(src)
	if err != nil {
		return -1, false, err
	}

	if b.incrementR2 {
		b.r2.Increment()
	}
	if b.decrementR2 {
		b.r2.Decrement()
	}

	b.r1.Assign(val)
	return 8, false, nil
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

func (b *Ld8BitImmediateOpCode) Run(cpu *Cpu) (int, bool, error) {
	immediateByte, err := cpu.LoadImmediateByte()
	if err != nil {
		return -1, false, err
	}
	b.r1.Assign(immediateByte)
	return 8, false, nil
}

func (b *Ld8BitImmediateOpCode) Name() string {
	return fmt.Sprintf("LD %s,d8", b.r1.Name)
}

type Ld16BitImmediateOpCode struct {
	BaseOpCode
	r1 *Register16Bit
}

func (b *Ld16BitImmediateOpCode) Run(cpu *Cpu) (int, bool, error) {
	immediateData, err := cpu.LoadImmediateWord()
	if err != nil {
		return -1, false, err
	}
	b.r1.Assign(immediateData)
	return 12, false, nil
}

func (b *Ld16BitImmediateOpCode) Name() string {
	return fmt.Sprintf("LD %s,d16", b.r1.Name)
}

type LdMemoryImmediateOpCode struct {
	BaseOpCode
	r1 *Register16Bit // This is always HL?
}

func (b *LdMemoryImmediateOpCode) Run(cpu *Cpu) (int, bool, error) {
	immediateData, err := cpu.LoadImmediateByte()
	if err != nil {
		return -1, false, err
	}
	targetAddress := b.r1.Retrieve()
	if err := cpu.memory.Set(targetAddress, immediateData); err != nil {
		return -1, false, err
	}
	return 12, false, nil
}

func (b *LdMemoryImmediateOpCode) Name() string {
	return fmt.Sprintf("LD (%s),d8", b.r1.Name)
}

type Inc8BitRegOpCode struct {
	BaseOpCode
	r1 *Register8Bit
}

func (b *Inc8BitRegOpCode) Run(cpu *Cpu) (int, bool, error) {
	zero, halfCarry := b.r1.Increment()
	cpu.SetFlag(Z, zero)
	cpu.SetFlag(H, halfCarry)
	cpu.SetFlag(N, false)
	return 4, false, nil
}

func (b *Inc8BitRegOpCode) Name() string {
	return fmt.Sprintf("INC %s", b.r1.Name)
}

type IncMemOpCode struct {
	BaseOpCode
	r1 *Register16Bit
}

func (b *IncMemOpCode) Run(cpu *Cpu) (int, bool, error) {
	val, err := cpu.memory.Get(b.r1.Retrieve())
	if err != nil {
		return -1, false, err
	}
	
	incResults := utils.Add8Bit(val, 0x1)
	cpu.SetFlag(Z, incResults.Zero)
	cpu.SetFlag(H, incResults.HalfCarry)
	cpu.SetFlag(N, false)
	if err := cpu.memory.Set(b.r1.Retrieve(), incResults.Result); err != nil {
		return -1, false, err
	}
	return 12, false, nil
}

func (b *IncMemOpCode) Name() string {
	return fmt.Sprintf("INC (%s)", b.r1.Name)
}

type Dec8BitRegOpCode struct {
	BaseOpCode
	r1 *Register8Bit
}

func (b *Dec8BitRegOpCode) Run(cpu *Cpu) (int, bool, error) {
	zero, halfCarry := b.r1.Decrement()
	cpu.SetFlag(Z, zero)
	cpu.SetFlag(H, halfCarry)
	cpu.SetFlag(N, true)
	return 4, false, nil
}

func (b *Dec8BitRegOpCode) Name() string {
	return fmt.Sprintf("DEC %s", b.r1.Name)
}

type Add8BitRegOpCode struct {
	BaseOpCode
	r1 *Register8Bit // Should always be A
	r2 *Register8Bit
	includeCarry bool
}

func (b *Add8BitRegOpCode) Run(cpu *Cpu) (int, bool, error) {
	result := utils.Add8BitWithCarry(b.r1.Retrieve(), b.r2.Retrieve(),
		b.includeCarry && cpu.GetFlag(C))
	b.r1.Assign(result.Result)
	cpu.SetFlag(Z, result.Zero)
	cpu.SetFlag(H, result.HalfCarry)
	cpu.SetFlag(C, result.Carry)
	cpu.SetFlag(N, false)
	return 4, false, nil
}

func (b *Add8BitRegOpCode) Name() string {
	base := "ADD"
	if b.includeCarry {
		base = "ADC"
	}
	return fmt.Sprintf("%s %s,%s", base, b.r1.Name, b.r2.Name)
}
