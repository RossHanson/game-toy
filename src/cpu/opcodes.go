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
	destLsb, destMsb := b.r1.Retrieve()
	cpu.memory.Set(destLsb, destMsb, b.r2.Retrieve())
	if b.incrementR1 {
		// TODO: figure out register increments
	}
	if b.decrementR1 {
		// TODO: figure out register decrements
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
	srcLsb, srcMsb := b.r2.Retrieve()
	val, err := cpu.memory.Get(srcLsb, srcMsb)
	if err != nil {
		return -1, err
	}

	if b.incrementR2 {
		// TODO: figure out register increments
	}
	if b.decrementR2 {
		// TODO: figure out register decrements
	}

	b.r1.Assign(val)
	return 8, nil
}

type Ld8BitImmediateOpCode struct {
	BaseOpCode
	r1 *Register8Bit
}

func (b *Ld8BitImmediateOpCode) Run(cpu *Cpu) (int, error) {
	immediateData, err := cpu.LoadImmediateData(b.length - 1)
	if err != nil {
		return 0, err
	}
	b.r1.Assign(immediateData[0])
	return 8, nil
}
