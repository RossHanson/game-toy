package gametoy

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

func (s *Memory) Get(address int) byte {
	// No bounds checking is done here
	return s.memory[address]
}

func (s *Memory) Set(address int, value byte) {
	s.memory[address] = value
}
