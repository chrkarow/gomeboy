package cpu

type flags byte

const (
	carry     flags = 1 << 4 // Carry
	halfCarry flags = 1 << 5 // Half carry (BCD)
	negative  flags = 1 << 6 // Subtraction (BCD)
	zero      flags = 1 << 7 // Zero
)

// setFlag sets the given flag to 1
func (flags *flags) setFlag(f flags) {
	*flags |= f
}

// unsetFlag sets the given flag to 0
func (flags *flags) unsetFlag(f flags) {
	*flags &= ^f
}

// isSet returns true if the given flag is set
func (flags *flags) isSet(f flags) bool {
	return *flags&f != 0
}

// reset set all flags to zero
func (flags *flags) reset() {
	*flags = 0x00
}
