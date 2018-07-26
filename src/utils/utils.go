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
