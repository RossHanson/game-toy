package cpu

type Ld8BitRegisterOpCode struct {
	BaseOpCode
	r1 *Register8Bit
	r2 *Register8Bit
}

func (b *Ld8BitRegisterOpCode) Run(cpu *Cpu) (int, error) {
	b.r1.Assign(b.r2.Retrieve())
	return 4, nil
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
	return 8, nil
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
	return 8, nil
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
	return 8, nil
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
	return 12, nil
}
