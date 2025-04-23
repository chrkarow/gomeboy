package bit

// IsSet8 returns true if the bit with the given index (least significant bit = 0, most significant bit = 1) is set.
func IsSet8(value byte, bitIndex byte) bool {
	mask := byte(1 << bitIndex)
	return value&mask == mask
}

// IsSet16 returns true if the bit with the given index (least significant bit = 0, most significant bit = 1) is set.
func IsSet16(value uint16, bitIndex byte) bool {
	mask := uint16(1 << bitIndex)
	return value&mask == mask
}

func Set(value *byte, bitIndex byte) {
	mask := byte(1 << bitIndex)
	*value |= mask
}

func Unset(value *byte, bitIndex byte) {
	mask := ^byte(1 << bitIndex)
	*value &= mask
}
