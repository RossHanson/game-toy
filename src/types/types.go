package types

import (
	"fmt"
)

type Word uint16

func (w Word) String() string {
	var zeroes string
	switch {
	case w < 0x0010:
		zeroes += "000"
	case w < 0x0100:
		zeroes += "00"
	case w < 0x100:
		zeroes += "0"
	}
	return fmt.Sprintf("0x%s%X", zeroes, uint16(w))
}

func WordFromBytes(lsb byte, msb byte) Word {
	return Word((uint16(msb) << 8) ^ uint16(lsb))
}

func (w Word) ToBytes() (lsb byte, msb byte) {
	return byte(uint16(w) & 0x00FF), byte(uint16(w) >> 8)
}
