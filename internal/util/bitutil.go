package util

// BitIsSet8 returns true if the bit with the given index (least significant bit = 0, most significant bit = 1) is set.
func BitIsSet8(value byte, bitIndex byte) bool {
	mask := byte(1 << bitIndex)
	return value&mask == mask
}

// BitIsSet16 returns true if the bit with the given index (least significant bit = 0, most significant bit = 1) is set.
func BitIsSet16(value uint16, bitIndex byte) bool {
	mask := uint16(1 << bitIndex)
	return value&mask == mask
}

// SetBit sets the bit at the given index to 1
func SetBit(value *byte, bitIndex byte) {
	mask := byte(1 << bitIndex)
	*value |= mask
}

// UnsetBit sets the bit at the given index to 0
func UnsetBit(value *byte, bitIndex byte) {
	mask := ^byte(1 << bitIndex)
	*value &= mask
}

// IsEmpty returns true if the given byte slice contains all zeroes
func IsEmpty(s []byte) bool {
	for _, b := range s {
		if b != 0x0 {
			return false
		}
	}
	return true
}
