package cpu

import (
	"fmt"
)

type LdRegisterOpCode struct {
	BaseOpCode
	r1 RegisterName
	r2 RegisterName
	// If one of the registers is 16 bit, it's assumed to be a load from memory
}

func (b *LdRegisterOpCode) Run(cpu *Cpu) (int, error) {
	r1 := cpu.registers[b.r1]
	r2 := cpu.registers[b.r2]
	if r1.is8Bit() && r2.is8Bit() {
		err := r1.Assign(r2.Retrieve()...)
		return 4, err
	}
	if !r1.is8Bit() && !r2.is8Bit() {
		return 0, fmt.Errorf("This op code is undefined")
	}
	if r1.is8Bit() {
		// Load the value in memory from r2
		memoryValue, err := cpu.memory.Get(r2.Retrieve())
		
		if err == nil {
			err = r1.Assign(memoryValue)
		}
		return 8, err
	}
	// Set the memory at r1 to the value from r2
	err := cpu.memory.Set(r1.Retrieve(), r2.Retrieve()[0])
	return 8, err
}

type LdImmediateOpCode struct {
	BaseOpCode
	r1 RegisterName
}

func (b *LdImmediateOpCode) Run(cpu *Cpu) (int, error) {
	r1 := cpu.registers[b.r1]
	immediateData, err := cpu.LoadImmediateData(b.length - 1)
	if err != nil {
		return 0, err
	}
	r1.Assign(immediateData...)
	if r1.is8Bit() {
		return 8, nil
	} else {
		return 12, nil
	}
}
