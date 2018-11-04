package utils

import (
	"encoding/binary"
)

// Encode an integer into a little-endian 16 bit byte array
func EncodeInt(value int) (lsb byte, msb byte) {
	result := make([]byte, 2)
	binary.LittleEndian.PutUint16(result, uint16(value))
	return result[0], result[1]
}

type ArithmeticResults8Bit struct {
	Result byte
	Zero bool
	HalfCarry bool
	Carry bool
}

func Add8Bit(value1 byte, value2 byte) ArithmeticResults8Bit {
	return Add8BitWithCarry(value1, value2, false)
}

func Add8BitWithCarry(value1 byte, value2 byte, carryBit bool) ArithmeticResults8Bit {
	carryVal := byte(0x0)
	if carryBit {
		carryVal = byte(0x1)
	}
	sum := value1 + value2 + carryVal
	return ArithmeticResults8Bit{
		Result: sum,
		Zero: sum == byte(0x0),
		HalfCarry: (((value1 & 0xF) + (value2 & 0xF) + carryVal) & 0x10) == 0x10,
		Carry: int(value1) + int(value2) + int(carryVal)> 0xFF,
	}
}

func Subtract8Bit(value1 byte, value2 byte) ArithmeticResults8Bit {
	res := value1 - value2
	return ArithmeticResults8Bit{
		Result: res,
		Zero: res == byte(0x0),
		HalfCarry: (((value1 & 0xF) - (value2 & 0xF)) & 0x10) == 0x10,
		Carry: value2 > value1,
	}
}
