package cpu

type flags struct {
	value byte
}
type flag byte

const (
	c flag = 1 << 4 // Carry
	h flag = 1 << 5 // Half carry (BCD)
	n flag = 1 << 6 // Subtraction (BCD)
	z flag = 1 << 7 // Zero
)

// setFlag sets the given flag to 1
func (flags *flags) setFlag(f flag) {
	flags.value |= byte(f)
}

// unsetFlag sets the given flag to 0
func (flags *flags) unsetFlag(f flag) {
	flags.value &= ^byte(f)
}

// isSet returns true if the given flag is set
func (flags *flags) isSet(f flag) bool {
	return flags.value&byte(f) != 0
}

// reset set all flags to zero
func (flags *flags) reset() {
	flags.value = 0x00
}

func (flags *flags) setValue(val byte) {
	flags.value = val
}

func (flags *flags) getValue() byte {
	return flags.value
}
