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
	// the shorter one is assumed to be smaller.
	if len(array1) != len(array2) {
		if len(array1) < len(array2) {
			return -1
		} else {
			return 1
		}
	}

	if len(array1) == 1 {
		if array1[0] == array2[0] {
			return 0
		} else if array1[0] < array2[0] {
			return -1
		} else {
			return 1
		}
	}

	val1 := binary.LittleEndian.Uint16(array1)
	val2 := binary.LittleEndian.Uint16(array2)
	if val1 == val2 {
		return 0
	} else if val1 < val2 {
		return -1
	}
	return 1
}
