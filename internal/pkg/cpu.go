package cpu

const (
	registerNames = []string{"A", "B", "C", "D", "E", "F", "H", "L"}
)

type OpCode struct {
	Name string
	Code int
}

type Register struct {
	Name string
	Value []byte
}

var (
	NOP = OpCode{"NOP", 0}
	ADD = OpCode{"ADD", 1}
)

type Cpu struct {
	// more fields to come
	memory *Memory
	registers map[string]Register
	programCounter Register
	stackPointer Register
	flagRegister Register
}

// Initializes a default CPU
func newCpu(memory *Memory) (*Cpu) {
	registerMap := make(map[string]Register)
	for _, registerName := range registerNames {
		registerMap[registerName] = Register{
			Name: registerName,
			Value: []byte{0},
		}
	}

	return &Cpu{
		memory: memory,
		registers: registerMap,
		programCounter: &Register{
			Name: "PC",
			Value: []byte{0, 100}, // might need to be hex...
		},
		stackPointer: &Register{
			Name: "SP",
			Value: []byte{255, 255} // TODO: Convert memory length into 16 bit array
		},
		flagRegister: &Register{
			Name: "FR",
			Value: []byte{0}, // TODO: this might be 16 bits
		},
	}
}

func (c *Cpu) runOp(op OpCode) (error) {
	
}
