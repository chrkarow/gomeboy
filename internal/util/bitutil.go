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

func SetBit(value *byte, bitIndex byte) {
	mask := byte(1 << bitIndex)
	*value |= mask
}

func UnsetBit(value *byte, bitIndex byte) {
	mask := ^byte(1 << bitIndex)
	*value &= mask
}
