package utils

import (
	"encoding/binary"
)

// Encode an integer into a little-endian 16 bit byte array
func EncodeInt(value int) []byte {
	result := make([]byte, 2)
	binary.LittleEndian.PutUint16(result, uint16(value))
	return result
}

func CompareByteArrays(array1 []byte, array2 []byte) int {
	// Returns -1 if array 1 is less than 2, 0 if equal, and
	// 1 if array 2 is larger. If the arrays of unequal size,
	// the smaller one is assumed to be the LSB of the larger.
	val1 := binary.LittleEndian.Uint16(array1)
	val2 := binary.LittleEndian.Uint16(array2)
	if val1 == val2 {
		return 0
	} else if val1 < val2 {
		return -1
	}
	return 1
}
