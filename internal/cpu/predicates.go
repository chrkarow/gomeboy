package cpu

func always() func() bool {
	return func() bool { return true }
}

func onCSet(cpu *CPU) func() bool {
	return func() bool { return cpu.f.isSet(c) }
}

func onCNotSet(cpu *CPU) func() bool {
	return func() bool { return !cpu.f.isSet(c) }
}

func onZSet(cpu *CPU) func() bool {
	return func() bool { return cpu.f.isSet(z) }
}

func onZNotSet(cpu *CPU) func() bool {
	return func() bool { return !cpu.f.isSet(z) }
}
