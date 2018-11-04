package cpu

type Ld8BitRegisterOpCode struct {
	BaseOpCode
	r1 *Register8Bit
	r2 *Register8Bit
}

func (b *Ld8BitRegisterOpCode) Run(cpu *Cpu) (int, bool, error) {
	b.r1.Assign(b.r2.Retrieve())
	return 4, false, nil
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


type IncRegOpCode struct {
	BaseOpCode
	r1 *Register8Bit
}

func (b *IncRegOpCode) Run(cpu *Cpu) (int, bool, error) {
	zero, halfCarry := b.r1.Increment()
	cpu.SetFlag(Z, zero)
	cpu.SetFlag(H, halfCarry)
	cpu.SetFlag(N, false)
	return 4, false, nil
}

type IncMemOpCode struct {
	BaseOpCode
	r1 *Register16Bit
}

func (b *IncMemOpCode) Run(cpu *Cpu) (int, bool, error) {
	_, err := cpu.memory.Get(b.r1.Retrieve())
	if err != nil {
		return -1, false, err
	}

	// TODO: set flags properly
	panic("Unimplemented!")
}

type DecRegOpCode struct {
	BaseOpCode
	r1 *Register8Bit
}

func (b *DecRegOpCode) Run(cpu *Cpu) (int, bool, error) {
	zero, halfCarry := b.r1.Decrement()
	cpu.SetFlag(Z, zero)
	cpu.SetFlag(H, halfCarry)
	cpu.SetFlag(N, true)
	return 4, false, nil
}
