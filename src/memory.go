package gametoy

import (
	"fmt"
	"encoding/binary"
)

const (
	mainMemorySize = 8 * 1024
	videoMemorySize = 8 * 1024
)

type Memory struct {
	memory []byte
}

func (s *Memory) Size() int {
	return len(s.memory)
}

func (s *Memory) Get(address []byte) (byte, error) {
	// No bounds checking is done here
	if len(address) != 2 {
		return 0x00, fmt.Errorf("addresses must be 16 bits")
	}
	uintAddress := binary.BigEndian.Uint16(address)
	return s.memory[int(uintAddress)], nil
}

func (s *Memory) Set(address []byte, value byte) error {
	if len(address) != 2 {
		return fmt.Errorf("addresses must be 16 bits")
	}
	uintAddress := binary.BigEndian.Uint16(address)
	s.memory[int(uintAddress)] = value
	return nil
}
