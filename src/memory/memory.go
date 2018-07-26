package memory

import (
	"types"
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

func (s *Memory) Get(address types.Word) (byte, error) {
	// No bounds checking is done here
	return s.memory[int(address)], nil
}

func (s *Memory) Set(address types.Word, value byte) error {
	s.memory[int(address)] = value
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
