package memory

import (
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

func (s *Memory) Get(lsb byte, msb byte) (byte, error) {
	// No bounds checking is done here
	uintAddress := binary.LittleEndian.Uint16([]byte{lsb, msb})
	return s.memory[int(uintAddress)], nil
}

func (s *Memory) GetInt(address int) (byte, error) {
	// Again no range checking yet
	return s.memory[address], nil 
}

func (s *Memory) Set(lsb byte, msb byte, value byte) error {
	uintAddress := binary.LittleEndian.Uint16([]byte{lsb, msb})
	s.memory[int(uintAddress)] = value
	return nil
}

func (s *Memory) SetInt(address int, value byte) error {
	s.memory[address] = value
	return nil
}

func SetupBlankMemory(size int) *Memory {
	return &Memory{
		memory: make([]byte, size),
	}
}

func InitializeMainMemory() *Memory {
	return &Memory{
		memory: make([]byte, mainMemorySize),
	}
}
