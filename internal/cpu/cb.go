package cpu

import (
	log "go.uber.org/zap"
)

type extendedInstruction struct {
	disassembly string
	execute     func(*CPU)
	ticks       byte
}

func (instr extendedInstruction) String() string {
	return instr.disassembly
}

var extendedInstructions = [256]extendedInstruction{
	{"RLC B", func(cpu *CPU) { cpu.bc.Hi.SetValue(rotateLeft(cpu.bc.Hi.GetValue(), &cpu.f)) }, 8}, // 0x00
	{"RLC C", func(cpu *CPU) { cpu.bc.Lo.SetValue(rotateLeft(cpu.bc.Lo.GetValue(), &cpu.f)) }, 8}, // 0x01
	{"RLC D", func(cpu *CPU) { cpu.de.Hi.SetValue(rotateLeft(cpu.de.Hi.GetValue(), &cpu.f)) }, 8}, // 0x02
	{"RLC E", func(cpu *CPU) { cpu.de.Lo.SetValue(rotateLeft(cpu.de.Lo.GetValue(), &cpu.f)) }, 8}, // 0x03
	{"RLC H", func(cpu *CPU) { cpu.hl.Hi.SetValue(rotateLeft(cpu.hl.Hi.GetValue(), &cpu.f)) }, 8}, // 0x04
	{"RLC L", func(cpu *CPU) { cpu.hl.Lo.SetValue(rotateLeft(cpu.hl.Lo.GetValue(), &cpu.f)) }, 8}, // 0x05
	{"RLC (HL)", func(cpu *CPU) {
		cpu.memory.Write8BitValue(cpu.hl.GetValue(), rotateLeft(cpu.memory.Read8BitValue(cpu.hl.GetValue()), &cpu.f))
	}, 16}, // 0x06
	{"RLC A", func(cpu *CPU) { (&cpu.a).SetValue(rotateLeft(cpu.a.GetValue(), &cpu.f)) }, 8},       // 0x07
	{"RRC B", func(cpu *CPU) { cpu.bc.Hi.SetValue(rotateRight(cpu.bc.Hi.GetValue(), &cpu.f)) }, 8}, // 0x08
	{"RRC C", func(cpu *CPU) { cpu.bc.Lo.SetValue(rotateRight(cpu.bc.Lo.GetValue(), &cpu.f)) }, 8}, // 0x09
	{"RRC D", func(cpu *CPU) { cpu.de.Hi.SetValue(rotateRight(cpu.de.Hi.GetValue(), &cpu.f)) }, 8}, // 0x0a
	{"RRC E", func(cpu *CPU) { cpu.de.Lo.SetValue(rotateRight(cpu.de.Lo.GetValue(), &cpu.f)) }, 8}, // 0x0b
	{"RRC H", func(cpu *CPU) { cpu.hl.Hi.SetValue(rotateRight(cpu.hl.Hi.GetValue(), &cpu.f)) }, 8}, // 0x0c
	{"RRC L", func(cpu *CPU) { cpu.hl.Lo.SetValue(rotateRight(cpu.hl.Lo.GetValue(), &cpu.f)) }, 8}, // 0x0d
	{"RRC (HL)", func(cpu *CPU) {
		cpu.memory.Write8BitValue(cpu.hl.GetValue(), rotateRight(cpu.memory.Read8BitValue(cpu.hl.GetValue()), &cpu.f))
	}, 16}, // 0x0e
	{"RRC A", func(cpu *CPU) { cpu.a.SetValue(rotateRight(cpu.a.GetValue(), &cpu.f)) }, 8},                   // 0x0f
	{"RL B", func(cpu *CPU) { cpu.bc.Hi.SetValue(rotateLeftThroughCarry(cpu.bc.Hi.GetValue(), &cpu.f)) }, 8}, // 0x10
	{"RL C", func(cpu *CPU) { cpu.bc.Lo.SetValue(rotateLeftThroughCarry(cpu.bc.Lo.GetValue(), &cpu.f)) }, 8}, // 0x11
	{"RL D", func(cpu *CPU) { cpu.de.Hi.SetValue(rotateLeftThroughCarry(cpu.de.Hi.GetValue(), &cpu.f)) }, 8}, // 0x12
	{"RL E", func(cpu *CPU) { cpu.de.Lo.SetValue(rotateLeftThroughCarry(cpu.de.Lo.GetValue(), &cpu.f)) }, 8}, // 0x13
	{"RL H", func(cpu *CPU) { cpu.hl.Hi.SetValue(rotateLeftThroughCarry(cpu.hl.Hi.GetValue(), &cpu.f)) }, 8}, // 0x14
	{"RL L", func(cpu *CPU) { cpu.hl.Lo.SetValue(rotateLeftThroughCarry(cpu.hl.Lo.GetValue(), &cpu.f)) }, 8}, // 0x15
	{"RL (HL)", func(cpu *CPU) {
		cpu.memory.Write8BitValue(cpu.hl.GetValue(), rotateLeftThroughCarry(cpu.memory.Read8BitValue(cpu.hl.GetValue()), &cpu.f))
	}, 16}, // 0x16
	{"RL A", func(cpu *CPU) { cpu.a.SetValue(rotateLeftThroughCarry(cpu.a.GetValue(), &cpu.f)) }, 8},          // 0x17
	{"RR B", func(cpu *CPU) { cpu.bc.Hi.SetValue(rotateRightThroughCarry(cpu.bc.Hi.GetValue(), &cpu.f)) }, 8}, // 0x18
	{"RR C", func(cpu *CPU) { cpu.bc.Lo.SetValue(rotateRightThroughCarry(cpu.bc.Lo.GetValue(), &cpu.f)) }, 8}, // 0x19
	{"RR D", func(cpu *CPU) { cpu.de.Hi.SetValue(rotateRightThroughCarry(cpu.de.Hi.GetValue(), &cpu.f)) }, 8}, // 0x1a
	{"RR E", func(cpu *CPU) { cpu.de.Lo.SetValue(rotateRightThroughCarry(cpu.de.Lo.GetValue(), &cpu.f)) }, 8}, // 0x1b
	{"RR H", func(cpu *CPU) { cpu.hl.Hi.SetValue(rotateRightThroughCarry(cpu.hl.Hi.GetValue(), &cpu.f)) }, 8}, // 0x1c
	{"RR L", func(cpu *CPU) { cpu.hl.Lo.SetValue(rotateRightThroughCarry(cpu.hl.Lo.GetValue(), &cpu.f)) }, 8}, // 0x1d
	{"RR (HL)", func(cpu *CPU) {
		cpu.memory.Write8BitValue(cpu.hl.GetValue(), rotateRightThroughCarry(cpu.memory.Read8BitValue(cpu.hl.GetValue()), &cpu.f))
	}, 16}, // 0x1e
	{"RR A", func(cpu *CPU) { cpu.a.SetValue(rotateRightThroughCarry(cpu.a.GetValue(), &cpu.f)) }, 8}, // 0x1f
	{"SLA B", func(cpu *CPU) { cpu.bc.Hi.SetValue(shiftLeft(cpu.bc.Hi.GetValue(), &cpu.f)) }, 8},      // 0x20
	{"SLA C", func(cpu *CPU) { cpu.bc.Lo.SetValue(shiftLeft(cpu.bc.Lo.GetValue(), &cpu.f)) }, 8},      // 0x21
	{"SLA D", func(cpu *CPU) { cpu.de.Hi.SetValue(shiftLeft(cpu.de.Hi.GetValue(), &cpu.f)) }, 8},      // 0x22
	{"SLA E", func(cpu *CPU) { cpu.de.Lo.SetValue(shiftLeft(cpu.de.Lo.GetValue(), &cpu.f)) }, 8},      // 0x23
	{"SLA H", func(cpu *CPU) { cpu.hl.Hi.SetValue(shiftLeft(cpu.hl.Hi.GetValue(), &cpu.f)) }, 8},      // 0x24
	{"SLA L", func(cpu *CPU) { cpu.hl.Lo.SetValue(shiftLeft(cpu.hl.Lo.GetValue(), &cpu.f)) }, 8},      // 0x25
	{"SLA (HL)", func(cpu *CPU) {
		cpu.memory.Write8BitValue(cpu.hl.GetValue(), shiftLeft(cpu.memory.Read8BitValue(cpu.hl.GetValue()), &cpu.f))
	}, 16}, // 0x26
	{"SLA A", func(cpu *CPU) { cpu.a.SetValue(shiftLeft(cpu.a.GetValue(), &cpu.f)) }, 8},                // 0x27
	{"SRA B", func(cpu *CPU) { cpu.bc.Hi.SetValue(shiftRight(cpu.bc.Hi.GetValue(), &cpu.f, true)) }, 8}, // 0x28
	{"SRA C", func(cpu *CPU) { cpu.bc.Lo.SetValue(shiftRight(cpu.bc.Lo.GetValue(), &cpu.f, true)) }, 8}, // 0x29
	{"SRA D", func(cpu *CPU) { cpu.de.Hi.SetValue(shiftRight(cpu.de.Hi.GetValue(), &cpu.f, true)) }, 8}, // 0x2a
	{"SRA E", func(cpu *CPU) { cpu.de.Lo.SetValue(shiftRight(cpu.de.Lo.GetValue(), &cpu.f, true)) }, 8}, // 0x2b
	{"SRA H", func(cpu *CPU) { cpu.hl.Hi.SetValue(shiftRight(cpu.hl.Hi.GetValue(), &cpu.f, true)) }, 8}, // 0x2c
	{"SRA L", func(cpu *CPU) { cpu.hl.Lo.SetValue(shiftRight(cpu.hl.Lo.GetValue(), &cpu.f, true)) }, 8}, // 0x2d
	{"SRA (HL)", func(cpu *CPU) {
		cpu.memory.Write8BitValue(cpu.hl.GetValue(), shiftRight(cpu.memory.Read8BitValue(cpu.hl.GetValue()), &cpu.f, true))
	}, 16}, // 0x2e
	{"SRA A", func(cpu *CPU) { cpu.a.SetValue(shiftRight(cpu.a.GetValue(), &cpu.f, true)) }, 8}, // 0x2f
	{"SWAP B", func(cpu *CPU) { cpu.bc.Hi.SetValue(swap(cpu.bc.Hi.GetValue(), &cpu.f)) }, 8},    // 0x30
	{"SWAP C", func(cpu *CPU) { cpu.bc.Lo.SetValue(swap(cpu.bc.Lo.GetValue(), &cpu.f)) }, 8},    // 0x31
	{"SWAP D", func(cpu *CPU) { cpu.de.Hi.SetValue(swap(cpu.de.Hi.GetValue(), &cpu.f)) }, 8},    // 0x32
	{"SWAP E", func(cpu *CPU) { cpu.de.Lo.SetValue(swap(cpu.de.Lo.GetValue(), &cpu.f)) }, 8},    // 0x33
	{"SWAP H", func(cpu *CPU) { cpu.hl.Hi.SetValue(swap(cpu.hl.Hi.GetValue(), &cpu.f)) }, 8},    // 0x34
	{"SWAP L", func(cpu *CPU) { cpu.hl.Lo.SetValue(swap(cpu.hl.Lo.GetValue(), &cpu.f)) }, 8},    // 0x35
	{"SWAP (HL)", func(cpu *CPU) {
		cpu.memory.Write8BitValue(cpu.hl.GetValue(), swap(cpu.memory.Read8BitValue(cpu.hl.GetValue()), &cpu.f))
	}, 16}, // 0x36
	{"SWAP A", func(cpu *CPU) { cpu.a.SetValue(swap(cpu.a.GetValue(), &cpu.f)) }, 8},                     // 0x37
	{"SRL B", func(cpu *CPU) { cpu.bc.Hi.SetValue(shiftRight(cpu.bc.Hi.GetValue(), &cpu.f, false)) }, 8}, // 0x38
	{"SRL C", func(cpu *CPU) { cpu.bc.Lo.SetValue(shiftRight(cpu.bc.Lo.GetValue(), &cpu.f, false)) }, 8}, // 0x39
	{"SRL D", func(cpu *CPU) { cpu.de.Hi.SetValue(shiftRight(cpu.de.Hi.GetValue(), &cpu.f, false)) }, 8}, // 0x3a
	{"SRL E", func(cpu *CPU) { cpu.de.Lo.SetValue(shiftRight(cpu.de.Lo.GetValue(), &cpu.f, false)) }, 8}, // 0x3b
	{"SRL H", func(cpu *CPU) { cpu.hl.Hi.SetValue(shiftRight(cpu.hl.Hi.GetValue(), &cpu.f, false)) }, 8}, // 0x3c
	{"SRL L", func(cpu *CPU) { cpu.hl.Lo.SetValue(shiftRight(cpu.hl.Lo.GetValue(), &cpu.f, false)) }, 8}, // 0x3d
	{"SRL (HL)", func(cpu *CPU) {
		cpu.memory.Write8BitValue(cpu.hl.GetValue(), shiftRight(cpu.memory.Read8BitValue(cpu.hl.GetValue()), &cpu.f, false))
	}, 16}, // 0x3e
	{"SRL A", func(cpu *CPU) { cpu.a.SetValue(shiftRight(cpu.a.GetValue(), &cpu.f, false)) }, 8},        // 0x3f
	{"BIT 0, B", func(cpu *CPU) { bit(cpu.bc.Hi.GetValue(), 0, &cpu.f) }, 8},                            // 0x40
	{"BIT 0, C", func(cpu *CPU) { bit(cpu.bc.Lo.GetValue(), 0, &cpu.f) }, 8},                            // 0x41
	{"BIT 0, D", func(cpu *CPU) { bit(cpu.de.Hi.GetValue(), 0, &cpu.f) }, 8},                            // 0x42
	{"BIT 0, E", func(cpu *CPU) { bit(cpu.de.Lo.GetValue(), 0, &cpu.f) }, 8},                            // 0x43
	{"BIT 0, H", func(cpu *CPU) { bit(cpu.hl.Hi.GetValue(), 0, &cpu.f) }, 8},                            // 0x44
	{"BIT 0, L", func(cpu *CPU) { bit(cpu.hl.Lo.GetValue(), 0, &cpu.f) }, 8},                            // 0x45
	{"BIT 0, (HL)", func(cpu *CPU) { bit(cpu.memory.Read8BitValue(cpu.hl.GetValue()), 0, &cpu.f) }, 16}, // 0x46
	{"BIT 0, A", func(cpu *CPU) { bit(cpu.a.GetValue(), 0, &cpu.f) }, 8},                                // 0x47
	{"BIT 1, B", func(cpu *CPU) { bit(cpu.bc.Hi.GetValue(), 1, &cpu.f) }, 8},                            // 0x48
	{"BIT 1, C", func(cpu *CPU) { bit(cpu.bc.Lo.GetValue(), 1, &cpu.f) }, 8},                            // 0x49
	{"BIT 1, D", func(cpu *CPU) { bit(cpu.de.Hi.GetValue(), 1, &cpu.f) }, 8},                            // 0x4a
	{"BIT 1, E", func(cpu *CPU) { bit(cpu.de.Lo.GetValue(), 1, &cpu.f) }, 8},                            // 0x4b
	{"BIT 1, H", func(cpu *CPU) { bit(cpu.hl.Hi.GetValue(), 1, &cpu.f) }, 8},                            // 0x4c
	{"BIT 1, L", func(cpu *CPU) { bit(cpu.hl.Lo.GetValue(), 1, &cpu.f) }, 8},                            // 0x4d
	{"BIT 1, (HL)", func(cpu *CPU) { bit(cpu.memory.Read8BitValue(cpu.hl.GetValue()), 1, &cpu.f) }, 16}, // 0x4e
	{"BIT 1, A", func(cpu *CPU) { bit(cpu.a.GetValue(), 1, &cpu.f) }, 8},                                // 0x4f
	{"BIT 2, B", func(cpu *CPU) { bit(cpu.bc.Hi.GetValue(), 2, &cpu.f) }, 8},                            // 0x50
	{"BIT 2, C", func(cpu *CPU) { bit(cpu.bc.Lo.GetValue(), 2, &cpu.f) }, 8},                            // 0x51
	{"BIT 2, D", func(cpu *CPU) { bit(cpu.de.Hi.GetValue(), 2, &cpu.f) }, 8},                            // 0x52
	{"BIT 2, E", func(cpu *CPU) { bit(cpu.de.Lo.GetValue(), 2, &cpu.f) }, 8},                            // 0x53
	{"BIT 2, H", func(cpu *CPU) { bit(cpu.hl.Hi.GetValue(), 2, &cpu.f) }, 8},                            // 0x54
	{"BIT 2, L", func(cpu *CPU) { bit(cpu.hl.Lo.GetValue(), 2, &cpu.f) }, 8},                            // 0x55
	{"BIT 2, (HL)", func(cpu *CPU) { bit(cpu.memory.Read8BitValue(cpu.hl.GetValue()), 2, &cpu.f) }, 16}, // 0x56
	{"BIT 2, A", func(cpu *CPU) { bit(cpu.a.GetValue(), 2, &cpu.f) }, 8},                                // 0x57
	{"BIT 3, B", func(cpu *CPU) { bit(cpu.bc.Hi.GetValue(), 3, &cpu.f) }, 8},                            // 0x58
	{"BIT 3, C", func(cpu *CPU) { bit(cpu.bc.Lo.GetValue(), 3, &cpu.f) }, 8},                            // 0x59
	{"BIT 3, D", func(cpu *CPU) { bit(cpu.de.Hi.GetValue(), 3, &cpu.f) }, 8},                            // 0x5a
	{"BIT 3, E", func(cpu *CPU) { bit(cpu.de.Lo.GetValue(), 3, &cpu.f) }, 8},                            // 0x5b
	{"BIT 3, H", func(cpu *CPU) { bit(cpu.hl.Hi.GetValue(), 3, &cpu.f) }, 8},                            // 0x5c
	{"BIT 3, L", func(cpu *CPU) { bit(cpu.hl.Lo.GetValue(), 3, &cpu.f) }, 8},                            // 0x5d
	{"BIT 3, (HL)", func(cpu *CPU) { bit(cpu.memory.Read8BitValue(cpu.hl.GetValue()), 3, &cpu.f) }, 16}, // 0x5e
	{"BIT 3, A", func(cpu *CPU) { bit(cpu.a.GetValue(), 3, &cpu.f) }, 8},                                // 0x5f
	{"BIT 4, B", func(cpu *CPU) { bit(cpu.bc.Hi.GetValue(), 4, &cpu.f) }, 8},                            // 0x60
	{"BIT 4, C", func(cpu *CPU) { bit(cpu.bc.Lo.GetValue(), 4, &cpu.f) }, 8},                            // 0x61
	{"BIT 4, D", func(cpu *CPU) { bit(cpu.de.Hi.GetValue(), 4, &cpu.f) }, 8},                            // 0x62
	{"BIT 4, E", func(cpu *CPU) { bit(cpu.de.Lo.GetValue(), 4, &cpu.f) }, 8},                            // 0x63
	{"BIT 4, H", func(cpu *CPU) { bit(cpu.hl.Hi.GetValue(), 4, &cpu.f) }, 8},                            // 0x64
	{"BIT 4, L", func(cpu *CPU) { bit(cpu.hl.Lo.GetValue(), 4, &cpu.f) }, 8},                            // 0x65
	{"BIT 4, (HL)", func(cpu *CPU) { bit(cpu.memory.Read8BitValue(cpu.hl.GetValue()), 4, &cpu.f) }, 16}, // 0x66
	{"BIT 4, A", func(cpu *CPU) { bit(cpu.a.GetValue(), 4, &cpu.f) }, 8},                                // 0x67
	{"BIT 5, B", func(cpu *CPU) { bit(cpu.bc.Hi.GetValue(), 5, &cpu.f) }, 8},                            // 0x68
	{"BIT 5, C", func(cpu *CPU) { bit(cpu.bc.Lo.GetValue(), 5, &cpu.f) }, 8},                            // 0x69
	{"BIT 5, D", func(cpu *CPU) { bit(cpu.de.Hi.GetValue(), 5, &cpu.f) }, 8},                            // 0x6a
	{"BIT 5, E", func(cpu *CPU) { bit(cpu.de.Lo.GetValue(), 5, &cpu.f) }, 8},                            // 0x6b
	{"BIT 5, H", func(cpu *CPU) { bit(cpu.hl.Hi.GetValue(), 5, &cpu.f) }, 8},                            // 0x6c
	{"BIT 5, L", func(cpu *CPU) { bit(cpu.hl.Lo.GetValue(), 5, &cpu.f) }, 8},                            // 0x6d
	{"BIT 5, (HL)", func(cpu *CPU) { bit(cpu.memory.Read8BitValue(cpu.hl.GetValue()), 5, &cpu.f) }, 16}, // 0x6e
	{"BIT 5, A", func(cpu *CPU) { bit(cpu.a.GetValue(), 5, &cpu.f) }, 8},                                // 0x6f
	{"BIT 6, B", func(cpu *CPU) { bit(cpu.bc.Hi.GetValue(), 6, &cpu.f) }, 8},                            // 0x70
	{"BIT 6, C", func(cpu *CPU) { bit(cpu.bc.Lo.GetValue(), 6, &cpu.f) }, 8},                            // 0x71
	{"BIT 6, D", func(cpu *CPU) { bit(cpu.de.Hi.GetValue(), 6, &cpu.f) }, 8},                            // 0x72
	{"BIT 6, E", func(cpu *CPU) { bit(cpu.de.Lo.GetValue(), 6, &cpu.f) }, 8},                            // 0x73
	{"BIT 6, H", func(cpu *CPU) { bit(cpu.hl.Hi.GetValue(), 6, &cpu.f) }, 8},                            // 0x74
	{"BIT 6, L", func(cpu *CPU) { bit(cpu.hl.Lo.GetValue(), 6, &cpu.f) }, 8},                            // 0x75
	{"BIT 6, (HL)", func(cpu *CPU) { bit(cpu.memory.Read8BitValue(cpu.hl.GetValue()), 6, &cpu.f) }, 16}, // 0x76
	{"BIT 6, A", func(cpu *CPU) { bit(cpu.a.GetValue(), 6, &cpu.f) }, 8},                                // 0x77
	{"BIT 7, B", func(cpu *CPU) { bit(cpu.bc.Hi.GetValue(), 7, &cpu.f) }, 8},                            // 0x78
	{"BIT 7, C", func(cpu *CPU) { bit(cpu.bc.Lo.GetValue(), 7, &cpu.f) }, 8},                            // 0x79
	{"BIT 7, D", func(cpu *CPU) { bit(cpu.de.Hi.GetValue(), 7, &cpu.f) }, 8},                            // 0x7a
	{"BIT 7, E", func(cpu *CPU) { bit(cpu.de.Lo.GetValue(), 7, &cpu.f) }, 8},                            // 0x7b
	{"BIT 7, H", func(cpu *CPU) { bit(cpu.hl.Hi.GetValue(), 7, &cpu.f) }, 8},                            // 0x7c
	{"BIT 7, L", func(cpu *CPU) { bit(cpu.hl.Lo.GetValue(), 7, &cpu.f) }, 8},                            // 0x7d
	{"BIT 7, (HL)", func(cpu *CPU) { bit(cpu.memory.Read8BitValue(cpu.hl.GetValue()), 7, &cpu.f) }, 16}, // 0x7e
	{"BIT 7, A", func(cpu *CPU) { bit(cpu.a.GetValue(), 7, &cpu.f) }, 8},                                // 0x7f
	{"RES 0, B", func(cpu *CPU) { cpu.bc.Hi.SetValue(res(cpu.bc.Hi.GetValue(), 0)) }, 8},                // 0x80
	{"RES 0, C", func(cpu *CPU) { cpu.bc.Lo.SetValue(res(cpu.bc.Lo.GetValue(), 0)) }, 8},                // 0x81
	{"RES 0, D", func(cpu *CPU) { cpu.de.Hi.SetValue(res(cpu.de.Hi.GetValue(), 0)) }, 8},                // 0x82
	{"RES 0, E", func(cpu *CPU) { cpu.de.Lo.SetValue(res(cpu.de.Lo.GetValue(), 0)) }, 8},                // 0x83
	{"RES 0, H", func(cpu *CPU) { cpu.hl.Hi.SetValue(res(cpu.hl.Hi.GetValue(), 0)) }, 8},                // 0x84
	{"RES 0, L", func(cpu *CPU) { cpu.hl.Lo.SetValue(res(cpu.hl.Lo.GetValue(), 0)) }, 8},                // 0x85
	{"RES 0, (HL)", func(cpu *CPU) {
		cpu.memory.Write8BitValue(cpu.hl.GetValue(), res(cpu.memory.Read8BitValue(cpu.hl.GetValue()), 0))
	}, 16}, // 0x86
	{"RES 0, A", func(cpu *CPU) { cpu.a.SetValue(res(cpu.a.GetValue(), 0)) }, 8},         // 0x87
	{"RES 1, B", func(cpu *CPU) { cpu.bc.Hi.SetValue(res(cpu.bc.Hi.GetValue(), 1)) }, 8}, // 0x88
	{"RES 1, C", func(cpu *CPU) { cpu.bc.Lo.SetValue(res(cpu.bc.Lo.GetValue(), 1)) }, 8}, // 0x89
	{"RES 1, D", func(cpu *CPU) { cpu.de.Hi.SetValue(res(cpu.de.Hi.GetValue(), 1)) }, 8}, // 0x8a
	{"RES 1, E", func(cpu *CPU) { cpu.de.Lo.SetValue(res(cpu.de.Lo.GetValue(), 1)) }, 8}, // 0x8b
	{"RES 1, H", func(cpu *CPU) { cpu.hl.Hi.SetValue(res(cpu.hl.Hi.GetValue(), 1)) }, 8}, // 0x8c
	{"RES 1, L", func(cpu *CPU) { cpu.hl.Lo.SetValue(res(cpu.hl.Lo.GetValue(), 1)) }, 8}, // 0x8d
	{"RES 1, (HL)", func(cpu *CPU) {
		cpu.memory.Write8BitValue(cpu.hl.GetValue(), res(cpu.memory.Read8BitValue(cpu.hl.GetValue()), 1))
	}, 16}, // 0x8e
	{"RES 1, A", func(cpu *CPU) { cpu.a.SetValue(res(cpu.a.GetValue(), 1)) }, 8},         // 0x8f
	{"RES 2, B", func(cpu *CPU) { cpu.bc.Hi.SetValue(res(cpu.bc.Hi.GetValue(), 2)) }, 8}, // 0x90
	{"RES 2, C", func(cpu *CPU) { cpu.bc.Lo.SetValue(res(cpu.bc.Lo.GetValue(), 2)) }, 8}, // 0x91
	{"RES 2, D", func(cpu *CPU) { cpu.de.Hi.SetValue(res(cpu.de.Hi.GetValue(), 2)) }, 8}, // 0x92
	{"RES 2, E", func(cpu *CPU) { cpu.de.Lo.SetValue(res(cpu.de.Lo.GetValue(), 2)) }, 8}, // 0x93
	{"RES 2, H", func(cpu *CPU) { cpu.hl.Hi.SetValue(res(cpu.hl.Hi.GetValue(), 2)) }, 8}, // 0x94
	{"RES 2, L", func(cpu *CPU) { cpu.hl.Lo.SetValue(res(cpu.hl.Lo.GetValue(), 2)) }, 8}, // 0x95
	{"RES 2, (HL)", func(cpu *CPU) {
		cpu.memory.Write8BitValue(cpu.hl.GetValue(), res(cpu.memory.Read8BitValue(cpu.hl.GetValue()), 2))
	}, 16}, // 0x96
	{"RES 2, A", func(cpu *CPU) { cpu.a.SetValue(res(cpu.a.GetValue(), 2)) }, 8},         // 0x97
	{"RES 3, B", func(cpu *CPU) { cpu.bc.Hi.SetValue(res(cpu.bc.Hi.GetValue(), 3)) }, 8}, // 0x98
	{"RES 3, C", func(cpu *CPU) { cpu.bc.Lo.SetValue(res(cpu.bc.Lo.GetValue(), 3)) }, 8}, // 0x99
	{"RES 3, D", func(cpu *CPU) { cpu.de.Hi.SetValue(res(cpu.de.Hi.GetValue(), 3)) }, 8}, // 0x9a
	{"RES 3, E", func(cpu *CPU) { cpu.de.Lo.SetValue(res(cpu.de.Lo.GetValue(), 3)) }, 8}, // 0x9b
	{"RES 3, H", func(cpu *CPU) { cpu.hl.Hi.SetValue(res(cpu.hl.Hi.GetValue(), 3)) }, 8}, // 0x9c
	{"RES 3, L", func(cpu *CPU) { cpu.hl.Lo.SetValue(res(cpu.hl.Lo.GetValue(), 3)) }, 8}, // 0x9d
	{"RES 3, (HL)", func(cpu *CPU) {
		cpu.memory.Write8BitValue(cpu.hl.GetValue(), res(cpu.memory.Read8BitValue(cpu.hl.GetValue()), 3))
	}, 16}, // 0x9e
	{"RES 3, A", func(cpu *CPU) { cpu.a.SetValue(res(cpu.a.GetValue(), 3)) }, 8},         // 0x9f
	{"RES 4, B", func(cpu *CPU) { cpu.bc.Hi.SetValue(res(cpu.bc.Hi.GetValue(), 4)) }, 8}, // 0xa0
	{"RES 4, C", func(cpu *CPU) { cpu.bc.Lo.SetValue(res(cpu.bc.Lo.GetValue(), 4)) }, 8}, // 0xa1
	{"RES 4, D", func(cpu *CPU) { cpu.de.Hi.SetValue(res(cpu.de.Hi.GetValue(), 4)) }, 8}, // 0xa2
	{"RES 4, E", func(cpu *CPU) { cpu.de.Lo.SetValue(res(cpu.de.Lo.GetValue(), 4)) }, 8}, // 0xa3
	{"RES 4, H", func(cpu *CPU) { cpu.hl.Hi.SetValue(res(cpu.hl.Hi.GetValue(), 4)) }, 8}, // 0xa4
	{"RES 4, L", func(cpu *CPU) { cpu.hl.Lo.SetValue(res(cpu.hl.Lo.GetValue(), 4)) }, 8}, // 0xa5
	{"RES 4, (HL)", func(cpu *CPU) {
		cpu.memory.Write8BitValue(cpu.hl.GetValue(), res(cpu.memory.Read8BitValue(cpu.hl.GetValue()), 4))
	}, 16}, // 0xa6
	{"RES 4, A", func(cpu *CPU) { cpu.a.SetValue(res(cpu.a.GetValue(), 4)) }, 8},         // 0xa7
	{"RES 5, B", func(cpu *CPU) { cpu.bc.Hi.SetValue(res(cpu.bc.Hi.GetValue(), 5)) }, 8}, // 0xa8
	{"RES 5, C", func(cpu *CPU) { cpu.bc.Lo.SetValue(res(cpu.bc.Lo.GetValue(), 5)) }, 8}, // 0xa9
	{"RES 5, D", func(cpu *CPU) { cpu.de.Hi.SetValue(res(cpu.de.Hi.GetValue(), 5)) }, 8}, // 0xaa
	{"RES 5, E", func(cpu *CPU) { cpu.de.Lo.SetValue(res(cpu.de.Lo.GetValue(), 5)) }, 8}, // 0xab
	{"RES 5, H", func(cpu *CPU) { cpu.hl.Hi.SetValue(res(cpu.hl.Hi.GetValue(), 5)) }, 8}, // 0xac
	{"RES 5, L", func(cpu *CPU) { cpu.hl.Lo.SetValue(res(cpu.hl.Lo.GetValue(), 5)) }, 8}, // 0xad
	{"RES 5, (HL)", func(cpu *CPU) {
		cpu.memory.Write8BitValue(cpu.hl.GetValue(), res(cpu.memory.Read8BitValue(cpu.hl.GetValue()), 5))
	}, 16}, // 0xae
	{"RES 5, A", func(cpu *CPU) { cpu.a.SetValue(res(cpu.a.GetValue(), 5)) }, 8},         // 0xaf
	{"RES 6, B", func(cpu *CPU) { cpu.bc.Hi.SetValue(res(cpu.bc.Hi.GetValue(), 6)) }, 8}, // 0xb0
	{"RES 6, C", func(cpu *CPU) { cpu.bc.Lo.SetValue(res(cpu.bc.Lo.GetValue(), 6)) }, 8}, // 0xb1
	{"RES 6, D", func(cpu *CPU) { cpu.de.Hi.SetValue(res(cpu.de.Hi.GetValue(), 6)) }, 8}, // 0xb2
	{"RES 6, E", func(cpu *CPU) { cpu.de.Lo.SetValue(res(cpu.de.Lo.GetValue(), 6)) }, 8}, // 0xb3
	{"RES 6, H", func(cpu *CPU) { cpu.hl.Hi.SetValue(res(cpu.hl.Hi.GetValue(), 6)) }, 8}, // 0xb4
	{"RES 6, L", func(cpu *CPU) { cpu.hl.Lo.SetValue(res(cpu.hl.Lo.GetValue(), 6)) }, 8}, // 0xb5
	{"RES 6, (HL)", func(cpu *CPU) {
		cpu.memory.Write8BitValue(cpu.hl.GetValue(), res(cpu.memory.Read8BitValue(cpu.hl.GetValue()), 6))
	}, 16}, // 0xb6
	{"RES 6, A", func(cpu *CPU) { cpu.a.SetValue(res(cpu.a.GetValue(), 6)) }, 8},         // 0xb7
	{"RES 7, B", func(cpu *CPU) { cpu.bc.Hi.SetValue(res(cpu.bc.Hi.GetValue(), 7)) }, 8}, // 0xb8
	{"RES 7, C", func(cpu *CPU) { cpu.bc.Lo.SetValue(res(cpu.bc.Lo.GetValue(), 7)) }, 8}, // 0xb9
	{"RES 7, D", func(cpu *CPU) { cpu.de.Hi.SetValue(res(cpu.de.Hi.GetValue(), 7)) }, 8}, // 0xba
	{"RES 7, E", func(cpu *CPU) { cpu.de.Lo.SetValue(res(cpu.de.Lo.GetValue(), 7)) }, 8}, // 0xbb
	{"RES 7, H", func(cpu *CPU) { cpu.hl.Hi.SetValue(res(cpu.hl.Hi.GetValue(), 7)) }, 8}, // 0xbc
	{"RES 7, L", func(cpu *CPU) { cpu.hl.Lo.SetValue(res(cpu.hl.Lo.GetValue(), 7)) }, 8}, // 0xbd
	{"RES 7, (HL)", func(cpu *CPU) {
		cpu.memory.Write8BitValue(cpu.hl.GetValue(), res(cpu.memory.Read8BitValue(cpu.hl.GetValue()), 7))
	}, 16}, // 0xbe
	{"RES 7, A", func(cpu *CPU) { cpu.a.SetValue(res(cpu.a.GetValue(), 7)) }, 8},         // 0xbf
	{"SET 0, B", func(cpu *CPU) { cpu.bc.Hi.SetValue(set(cpu.bc.Hi.GetValue(), 0)) }, 8}, // 0xc0
	{"SET 0, C", func(cpu *CPU) { cpu.bc.Lo.SetValue(set(cpu.bc.Lo.GetValue(), 0)) }, 8}, // 0xc1
	{"SET 0, D", func(cpu *CPU) { cpu.de.Hi.SetValue(set(cpu.de.Hi.GetValue(), 0)) }, 8}, // 0xc2
	{"SET 0, E", func(cpu *CPU) { cpu.de.Lo.SetValue(set(cpu.de.Lo.GetValue(), 0)) }, 8}, // 0xc3
	{"SET 0, H", func(cpu *CPU) { cpu.hl.Hi.SetValue(set(cpu.hl.Hi.GetValue(), 0)) }, 8}, // 0xc4
	{"SET 0, L", func(cpu *CPU) { cpu.hl.Lo.SetValue(set(cpu.hl.Lo.GetValue(), 0)) }, 8}, // 0xc5
	{"SET 0, (HL)", func(cpu *CPU) {
		cpu.memory.Write8BitValue(cpu.hl.GetValue(), set(cpu.memory.Read8BitValue(cpu.hl.GetValue()), 0))
	}, 16}, // 0xc6
	{"SET 0, A", func(cpu *CPU) { cpu.a.SetValue(set(cpu.a.GetValue(), 0)) }, 8},         // 0xc7
	{"SET 1, B", func(cpu *CPU) { cpu.bc.Hi.SetValue(set(cpu.bc.Hi.GetValue(), 1)) }, 8}, // 0xc8
	{"SET 1, C", func(cpu *CPU) { cpu.bc.Lo.SetValue(set(cpu.bc.Lo.GetValue(), 1)) }, 8}, // 0xc9
	{"SET 1, D", func(cpu *CPU) { cpu.de.Hi.SetValue(set(cpu.de.Hi.GetValue(), 1)) }, 8}, // 0xca
	{"SET 1, E", func(cpu *CPU) { cpu.de.Lo.SetValue(set(cpu.de.Lo.GetValue(), 1)) }, 8}, // 0xcb
	{"SET 1, H", func(cpu *CPU) { cpu.hl.Hi.SetValue(set(cpu.hl.Hi.GetValue(), 1)) }, 8}, // 0xcc
	{"SET 1, L", func(cpu *CPU) { cpu.hl.Lo.SetValue(set(cpu.hl.Lo.GetValue(), 1)) }, 8}, // 0xcd
	{"SET 1, (HL)", func(cpu *CPU) {
		cpu.memory.Write8BitValue(cpu.hl.GetValue(), set(cpu.memory.Read8BitValue(cpu.hl.GetValue()), 1))
	}, 16}, // 0xce
	{"SET 1, A", func(cpu *CPU) { cpu.a.SetValue(set(cpu.a.GetValue(), 1)) }, 8},         // 0xcf
	{"SET 2, B", func(cpu *CPU) { cpu.bc.Hi.SetValue(set(cpu.bc.Hi.GetValue(), 2)) }, 8}, // 0xd0
	{"SET 2, C", func(cpu *CPU) { cpu.bc.Lo.SetValue(set(cpu.bc.Lo.GetValue(), 2)) }, 8}, // 0xd1
	{"SET 2, D", func(cpu *CPU) { cpu.de.Hi.SetValue(set(cpu.de.Hi.GetValue(), 2)) }, 8}, // 0xd2
	{"SET 2, E", func(cpu *CPU) { cpu.de.Lo.SetValue(set(cpu.de.Lo.GetValue(), 2)) }, 8}, // 0xd3
	{"SET 2, H", func(cpu *CPU) { cpu.hl.Hi.SetValue(set(cpu.hl.Hi.GetValue(), 2)) }, 8}, // 0xd4
	{"SET 2, L", func(cpu *CPU) { cpu.hl.Lo.SetValue(set(cpu.hl.Lo.GetValue(), 2)) }, 8}, // 0xd5
	{"SET 2, (HL)", func(cpu *CPU) {
		cpu.memory.Write8BitValue(cpu.hl.GetValue(), set(cpu.memory.Read8BitValue(cpu.hl.GetValue()), 2))
	}, 16}, // 0xd6
	{"SET 2, A", func(cpu *CPU) { cpu.a.SetValue(set(cpu.a.GetValue(), 2)) }, 8},         // 0xd7
	{"SET 3, B", func(cpu *CPU) { cpu.bc.Hi.SetValue(set(cpu.bc.Hi.GetValue(), 3)) }, 8}, // 0xd8
	{"SET 3, C", func(cpu *CPU) { cpu.bc.Lo.SetValue(set(cpu.bc.Lo.GetValue(), 3)) }, 8}, // 0xd9
	{"SET 3, D", func(cpu *CPU) { cpu.de.Hi.SetValue(set(cpu.de.Hi.GetValue(), 3)) }, 8}, // 0xda
	{"SET 3, E", func(cpu *CPU) { cpu.de.Lo.SetValue(set(cpu.de.Lo.GetValue(), 3)) }, 8}, // 0xdb
	{"SET 3, H", func(cpu *CPU) { cpu.hl.Hi.SetValue(set(cpu.hl.Hi.GetValue(), 3)) }, 8}, // 0xdc
	{"SET 3, L", func(cpu *CPU) { cpu.hl.Lo.SetValue(set(cpu.hl.Lo.GetValue(), 3)) }, 8}, // 0xdd
	{"SET 3, (HL)", func(cpu *CPU) {
		cpu.memory.Write8BitValue(cpu.hl.GetValue(), set(cpu.memory.Read8BitValue(cpu.hl.GetValue()), 3))
	}, 16}, // 0xde
	{"SET 3, A", func(cpu *CPU) { cpu.a.SetValue(set(cpu.a.GetValue(), 3)) }, 8},          // 0xdf
	{"SET 4, B", func(cpu *CPU) { cpu.bc.Hi.SetValue(set(cpu.bc.Hi.GetValue(), 4)) }, 16}, // 0xe0
	{"SET 4, C", func(cpu *CPU) { cpu.bc.Lo.SetValue(set(cpu.bc.Lo.GetValue(), 4)) }, 16}, // 0xe1
	{"SET 4, D", func(cpu *CPU) { cpu.de.Hi.SetValue(set(cpu.de.Hi.GetValue(), 4)) }, 16}, // 0xe2
	{"SET 4, E", func(cpu *CPU) { cpu.de.Lo.SetValue(set(cpu.de.Lo.GetValue(), 4)) }, 16}, // 0xe3
	{"SET 4, H", func(cpu *CPU) { cpu.hl.Hi.SetValue(set(cpu.hl.Hi.GetValue(), 4)) }, 16}, // 0xe4
	{"SET 4, L", func(cpu *CPU) { cpu.hl.Lo.SetValue(set(cpu.hl.Lo.GetValue(), 4)) }, 16}, // 0xe5
	{"SET 4, (HL)", func(cpu *CPU) {
		cpu.memory.Write8BitValue(cpu.hl.GetValue(), set(cpu.memory.Read8BitValue(cpu.hl.GetValue()), 4))
	}, 8}, // 0xe6
	{"SET 4, A", func(cpu *CPU) { cpu.a.SetValue(set(cpu.a.GetValue(), 4)) }, 8},         // 0xe7
	{"SET 5, B", func(cpu *CPU) { cpu.bc.Hi.SetValue(set(cpu.bc.Hi.GetValue(), 5)) }, 8}, // 0xe8
	{"SET 5, C", func(cpu *CPU) { cpu.bc.Lo.SetValue(set(cpu.bc.Lo.GetValue(), 5)) }, 8}, // 0xe9
	{"SET 5, D", func(cpu *CPU) { cpu.de.Hi.SetValue(set(cpu.de.Hi.GetValue(), 5)) }, 8}, // 0xea
	{"SET 5, E", func(cpu *CPU) { cpu.de.Lo.SetValue(set(cpu.de.Lo.GetValue(), 5)) }, 8}, // 0xeb
	{"SET 5, H", func(cpu *CPU) { cpu.hl.Hi.SetValue(set(cpu.hl.Hi.GetValue(), 5)) }, 8}, // 0xec
	{"SET 5, L", func(cpu *CPU) { cpu.hl.Lo.SetValue(set(cpu.hl.Lo.GetValue(), 5)) }, 8}, // 0xed
	{"SET 5, (HL)", func(cpu *CPU) {
		cpu.memory.Write8BitValue(cpu.hl.GetValue(), set(cpu.memory.Read8BitValue(cpu.hl.GetValue()), 5))
	}, 16}, // 0xee
	{"SET 5, A", func(cpu *CPU) { cpu.a.SetValue(set(cpu.a.GetValue(), 5)) }, 8},         // 0xef
	{"SET 6, B", func(cpu *CPU) { cpu.bc.Hi.SetValue(set(cpu.bc.Hi.GetValue(), 6)) }, 8}, // 0xf0
	{"SET 6, C", func(cpu *CPU) { cpu.bc.Lo.SetValue(set(cpu.bc.Lo.GetValue(), 6)) }, 8}, // 0xf1
	{"SET 6, D", func(cpu *CPU) { cpu.de.Hi.SetValue(set(cpu.de.Hi.GetValue(), 6)) }, 8}, // 0xf2
	{"SET 6, E", func(cpu *CPU) { cpu.de.Lo.SetValue(set(cpu.de.Lo.GetValue(), 6)) }, 8}, // 0xf3
	{"SET 6, H", func(cpu *CPU) { cpu.hl.Hi.SetValue(set(cpu.hl.Hi.GetValue(), 6)) }, 8}, // 0xf4
	{"SET 6, L", func(cpu *CPU) { cpu.hl.Lo.SetValue(set(cpu.hl.Lo.GetValue(), 6)) }, 8}, // 0xf5
	{"SET 6, (HL)", func(cpu *CPU) {
		cpu.memory.Write8BitValue(cpu.hl.GetValue(), set(cpu.memory.Read8BitValue(cpu.hl.GetValue()), 6))
	}, 16}, // 0xf6
	{"SET 6, A", func(cpu *CPU) { cpu.a.SetValue(set(cpu.a.GetValue(), 6)) }, 8},         // 0xf7
	{"SET 7, B", func(cpu *CPU) { cpu.bc.Hi.SetValue(set(cpu.bc.Hi.GetValue(), 7)) }, 8}, // 0xf8
	{"SET 7, C", func(cpu *CPU) { cpu.bc.Lo.SetValue(set(cpu.bc.Lo.GetValue(), 7)) }, 8}, // 0xf9
	{"SET 7, D", func(cpu *CPU) { cpu.de.Hi.SetValue(set(cpu.de.Hi.GetValue(), 7)) }, 8}, // 0xfa
	{"SET 7, E", func(cpu *CPU) { cpu.de.Lo.SetValue(set(cpu.de.Lo.GetValue(), 7)) }, 8}, // 0xfb
	{"SET 7, H", func(cpu *CPU) { cpu.hl.Hi.SetValue(set(cpu.hl.Hi.GetValue(), 7)) }, 8}, // 0xfc
	{"SET 7, L", func(cpu *CPU) { cpu.hl.Lo.SetValue(set(cpu.hl.Lo.GetValue(), 7)) }, 8}, // 0xfd
	{"SET 7, (HL)", func(cpu *CPU) {
		cpu.memory.Write8BitValue(cpu.hl.GetValue(), set(cpu.memory.Read8BitValue(cpu.hl.GetValue()), 7))
	}, 16}, // 0xfe
	{"SET 7, A", func(cpu *CPU) { cpu.a.SetValue(set(cpu.a.GetValue(), 7)) }, 8}, // 0xff
}

func executeExtendedInstruction(cpu *CPU, ticks *int, optCode byte) {
	instr := extendedInstructions[optCode]
	log.L().Info(instr.disassembly)

	instr.execute(cpu)
	*ticks += int(instr.ticks)
}

func shiftLeft(value byte, flags *flags) byte {
	// get the highest bit of value by masking it with 0b10000000 (aka 0x80 or 0d128) and shifting
	// it 7 bits down
	carry := (value & 0x80) >> 7

	if carry == 1 {
		flags.setFlag(c)
	} else {
		flags.unsetFlag(c)
	}

	// shift value 1bit up
	value <<= 1

	if value == 0 {
		flags.setFlag(z)
	} else {
		flags.unsetFlag(z)
	}

	flags.unsetFlag(n)
	flags.unsetFlag(h)

	return value
}

func rotateLeft(value byte, flags *flags) byte {
	// get the highest bit of value by masking it with 0b10000000 (aka 0x80 or 0d128) and shifting
	// it 7 bits down
	carry := (value & 0x80) >> 7

	if carry == 1 {
		flags.setFlag(c)
	} else {
		flags.unsetFlag(c)
	}

	// shift value 1bit up
	value <<= 1
	// append carry at its end
	value += carry

	if value == 0 {
		flags.setFlag(z)
	} else {
		flags.unsetFlag(z)
	}

	flags.unsetFlag(n)
	flags.unsetFlag(h)

	return value
}

func rotateLeftThroughCarry(value byte, flags *flags) byte {
	carryFlagWasSet := flags.isSet(c)

	// get the highest bit of value by masking it with 0b10000000 (aka 0x80 or 0d128) and shifting
	// it 7 bits down
	carry := (value & 0x80) >> 7

	if carry == 1 {
		flags.setFlag(c)
	} else {
		flags.unsetFlag(c)
	}

	// shift value 1bit up
	value <<= 1
	if carryFlagWasSet {
		value++
	}

	if value == 0 {
		flags.setFlag(z)
	} else {
		flags.unsetFlag(z)
	}

	flags.unsetFlag(n)
	flags.unsetFlag(h)

	return value
}

func shiftRight(value byte, flags *flags, arithmetic bool) byte {
	carry := value & 0x01
	highestBit := value & 0x80

	if carry == 1 {
		flags.setFlag(c)
	} else {
		flags.unsetFlag(c)
	}

	// shift value 1bit down
	value >>= 1
	if arithmetic {
		// highest bit must stay unchanged
		value += highestBit
	}

	if value == 0 {
		flags.setFlag(z)
	} else {
		flags.unsetFlag(z)
	}

	flags.unsetFlag(n)
	flags.unsetFlag(h)

	return value
}

func rotateRight(value byte, flags *flags) byte {
	carry := value & 0x01

	if carry == 1 {
		flags.setFlag(c)
	} else {
		flags.unsetFlag(c)
	}

	// shift value 1bit down
	value >>= 1
	value += carry << 7

	if value == 0 {
		flags.setFlag(z)
	} else {
		flags.unsetFlag(z)
	}

	flags.unsetFlag(n)
	flags.unsetFlag(h)

	return value
}

func rotateRightThroughCarry(value byte, flags *flags) byte {
	carryFlagWasSet := flags.isSet(c)

	carry := value & 0x01

	if carry == 1 {
		flags.setFlag(c)
	} else {
		flags.unsetFlag(c)
	}

	// shift value 1bit down
	value >>= 1
	if carryFlagWasSet {
		value += 1 << 7
	}

	if value == 0 {
		flags.setFlag(z)
	} else {
		flags.unsetFlag(z)
	}

	flags.unsetFlag(n)
	flags.unsetFlag(h)
	return value
}

func swap(value byte, flags *flags) byte {
	if value == 0 {
		flags.setFlag(z)
		return value

	} else {
		flags.unsetFlag(z)
	}

	upperNibble := value & 0xf0
	lowerNibble := value & 0x0f

	return upperNibble>>4 | lowerNibble<<4
}

func bit(value byte, index byte, flags *flags) {
	checkMask := byte(1) << index

	if value&checkMask == 0 {
		flags.setFlag(z)
	} else {
		flags.unsetFlag(z)
	}

	flags.setFlag(h)
	flags.unsetFlag(n)
}

func res(value byte, index byte) byte {
	return value & ^(byte(1) << index)
}

func set(value byte, index byte) byte {
	return value | (byte(1) << index)
}
