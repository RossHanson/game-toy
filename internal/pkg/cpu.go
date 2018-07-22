package cpu

const (
	registers8bit = []string{"A", "B", "C", "D", "E", "F", "H", "L"}
	combinationRegisters = []string{"AF", "BC", "DE", "HL"}
	registers16Bit = []string{ "PC", "SP", "FR"}
)

type RegisterName string

type OpCode interface {
	Run(cpu Cpu) error
	Name() string
	DebugString() string
	Cycles() int
}

type Register struct {
	Name string
	Value []byte
}

type Cpu struct {
	// more fields to come
	memory *Memory
	registers map[string]Register
	programCounter Register
	stackPointer Register
	flagRegister Register
}

type BaseOpCode struct {
	name string
	code byte
	cycles int
}

func (b BaseOpCode) Name() string {
	return b.name
}

func (b BaseOpCode) DebugString() string {
	return fmt.Sprintf("%x - %s", b.code, b.name)
}

func (b BaseOpCode) Cycles() int {
	return b.cycles
}

type LdRegisterOpCode struct {
	BaseOpCode
	r1 RegisterName
	r2 RegisterName
}

func (b LdRegisterOpCode) Run(cpu Cpu) error {
	cpu.registers[r1].Value[0] = cpu.registers[r2].value[0]
}

// Initializes a default CPU
func newCpu(memory *Memory) (*Cpu) {
	registerMap := make(map[string]Register)
	for _, registerName := range registers8Bit {
		var value byte
		registerMap[registerName] = Register{
			Name: registerName,
			Value: []*byte{&value},
		}
	}

	for _, registerName := range registers16Bit {
		var value1 byte
		var value2 byte
		registerMap[registerName] = Register{
			Name: registerName,
			Value: []*byte{&value1, &value2},
		}
	}

	for _, registerName := range combinationRegisters {
		register1 := registerMap[registerName[0]]
		register2 := registerMap[registerName[1]]
		registerMap[registerName] = Register{
			Name: registerName,
			Value: []*byte{register1.Value[0], register2.Value[1]}.
		}
	}

	return &Cpu{
		memory: memory,
		registers: registerMap,
		programCounter: registerMap["PC"],
		stackPointer: registerMap["SP"],
		flagRegister: registerMap["FR"],
	}
}
