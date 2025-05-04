package cpu

import "gameboy-emulator/internal/util"

var extendedInstructions [256]instruction

func init() {
	extendedInstructions = [256]instruction{
		{"RLC B", func(c *CPU) { fetchCycle(c, func(c *CPU) { rotateLeft(c, &c.b) }) }}, // 0x00
		{"RLC C", func(c *CPU) { fetchCycle(c, func(c *CPU) { rotateLeft(c, &c.c) }) }}, // 0x01
		{"RLC D", func(c *CPU) { fetchCycle(c, func(c *CPU) { rotateLeft(c, &c.d) }) }}, // 0x02
		{"RLC E", func(c *CPU) { fetchCycle(c, func(c *CPU) { rotateLeft(c, &c.e) }) }}, // 0x03
		{"RLC H", func(c *CPU) { fetchCycle(c, func(c *CPU) { rotateLeft(c, &c.h) }) }}, // 0x04
		{"RLC L", func(c *CPU) { fetchCycle(c, func(c *CPU) { rotateLeft(c, &c.l) }) }}, // 0x05
		{"RLC (HL)", func(c *CPU) {
			c.ops.Push(func(c *CPU) {
				c.z = c.mmu.Read(c.hl())
			})
			c.ops.Push(func(c *CPU) {
				rotateLeft(c, &c.z)
				c.mmu.Write(c.hl(), c.z)
			})
			fetchCycle(c)
		}}, // 0x06
		{"RLC A", func(c *CPU) { fetchCycle(c, func(c *CPU) { rotateLeft(c, &c.a) }) }},  // 0x07
		{"RRC B", func(c *CPU) { fetchCycle(c, func(c *CPU) { rotateRight(c, &c.b) }) }}, // 0x08
		{"RRC C", func(c *CPU) { fetchCycle(c, func(c *CPU) { rotateRight(c, &c.c) }) }}, // 0x09
		{"RRC D", func(c *CPU) { fetchCycle(c, func(c *CPU) { rotateRight(c, &c.d) }) }}, // 0x0a
		{"RRC E", func(c *CPU) { fetchCycle(c, func(c *CPU) { rotateRight(c, &c.e) }) }}, // 0x0b
		{"RRC H", func(c *CPU) { fetchCycle(c, func(c *CPU) { rotateRight(c, &c.h) }) }}, // 0x0c
		{"RRC L", func(c *CPU) { fetchCycle(c, func(c *CPU) { rotateRight(c, &c.l) }) }}, // 0x0d
		{"RRC (HL)", func(c *CPU) {
			c.ops.Push(func(c *CPU) {
				c.z = c.mmu.Read(c.hl())
			})
			c.ops.Push(func(c *CPU) {
				rotateRight(c, &c.z)
				c.mmu.Write(c.hl(), c.z)
			})
			fetchCycle(c)
		}}, // 0x0e
		{"RRC A", func(c *CPU) { fetchCycle(c, func(c *CPU) { rotateRight(c, &c.a) }) }},           // 0x0f
		{"RL B", func(c *CPU) { fetchCycle(c, func(c *CPU) { rotateLeftThroughCarry(c, &c.b) }) }}, // 0x10
		{"RL C", func(c *CPU) { fetchCycle(c, func(c *CPU) { rotateLeftThroughCarry(c, &c.c) }) }}, // 0x11
		{"RL D", func(c *CPU) { fetchCycle(c, func(c *CPU) { rotateLeftThroughCarry(c, &c.d) }) }}, // 0x12
		{"RL E", func(c *CPU) { fetchCycle(c, func(c *CPU) { rotateLeftThroughCarry(c, &c.e) }) }}, // 0x13
		{"RL H", func(c *CPU) { fetchCycle(c, func(c *CPU) { rotateLeftThroughCarry(c, &c.h) }) }}, // 0x14
		{"RL L", func(c *CPU) { fetchCycle(c, func(c *CPU) { rotateLeftThroughCarry(c, &c.l) }) }}, // 0x15
		{"RL (HL)", func(c *CPU) {
			c.ops.Push(func(c *CPU) {
				c.z = c.mmu.Read(c.hl())
			})
			c.ops.Push(func(c *CPU) {
				rotateLeftThroughCarry(c, &c.z)
				c.mmu.Write(c.hl(), c.z)
			})
			fetchCycle(c)
		}}, // 0x16
		{"RL A", func(c *CPU) { fetchCycle(c, func(c *CPU) { rotateLeftThroughCarry(c, &c.a) }) }},  // 0x17
		{"RR B", func(c *CPU) { fetchCycle(c, func(c *CPU) { rotateRightThroughCarry(c, &c.b) }) }}, // 0x18
		{"RR C", func(c *CPU) { fetchCycle(c, func(c *CPU) { rotateRightThroughCarry(c, &c.c) }) }}, // 0x19
		{"RR D", func(c *CPU) { fetchCycle(c, func(c *CPU) { rotateRightThroughCarry(c, &c.d) }) }}, // 0x1a
		{"RR E", func(c *CPU) { fetchCycle(c, func(c *CPU) { rotateRightThroughCarry(c, &c.e) }) }}, // 0x1b
		{"RR H", func(c *CPU) { fetchCycle(c, func(c *CPU) { rotateRightThroughCarry(c, &c.h) }) }}, // 0x1c
		{"RR L", func(c *CPU) { fetchCycle(c, func(c *CPU) { rotateRightThroughCarry(c, &c.l) }) }}, // 0x1d
		{"RR (HL)", func(c *CPU) {
			c.ops.Push(func(c *CPU) {
				c.z = c.mmu.Read(c.hl())
			})
			c.ops.Push(func(c *CPU) {
				rotateRightThroughCarry(c, &c.z)
				c.mmu.Write(c.hl(), c.z)
			})
			fetchCycle(c)
		}}, // 0x1e
		{"RR A", func(c *CPU) { fetchCycle(c, func(c *CPU) { rotateRightThroughCarry(c, &c.a) }) }}, // 0x1f
		{"SLA B", func(c *CPU) { fetchCycle(c, func(c *CPU) { shiftLeft(c, &c.b) }) }},              // 0x20
		{"SLA C", func(c *CPU) { fetchCycle(c, func(c *CPU) { shiftLeft(c, &c.c) }) }},              // 0x21
		{"SLA D", func(c *CPU) { fetchCycle(c, func(c *CPU) { shiftLeft(c, &c.d) }) }},              // 0x22
		{"SLA E", func(c *CPU) { fetchCycle(c, func(c *CPU) { shiftLeft(c, &c.e) }) }},              // 0x23
		{"SLA H", func(c *CPU) { fetchCycle(c, func(c *CPU) { shiftLeft(c, &c.h) }) }},              // 0x24
		{"SLA L", func(c *CPU) { fetchCycle(c, func(c *CPU) { shiftLeft(c, &c.l) }) }},              // 0x25
		{"SLA (HL)", func(c *CPU) {
			c.ops.Push(func(c *CPU) {
				c.z = c.mmu.Read(c.hl())
			})
			c.ops.Push(func(c *CPU) {
				shiftLeft(c, &c.z)
				c.mmu.Write(c.hl(), c.z)
			})
			fetchCycle(c)
		}}, // 0x26
		{"SLA A", func(c *CPU) { fetchCycle(c, func(c *CPU) { shiftLeft(c, &c.a) }) }},        // 0x27
		{"SRA B", func(c *CPU) { fetchCycle(c, func(c *CPU) { shiftRight(c, &c.b, true) }) }}, // 0x28
		{"SRA C", func(c *CPU) { fetchCycle(c, func(c *CPU) { shiftRight(c, &c.c, true) }) }}, // 0x29
		{"SRA D", func(c *CPU) { fetchCycle(c, func(c *CPU) { shiftRight(c, &c.d, true) }) }}, // 0x2a
		{"SRA E", func(c *CPU) { fetchCycle(c, func(c *CPU) { shiftRight(c, &c.e, true) }) }}, // 0x2b
		{"SRA H", func(c *CPU) { fetchCycle(c, func(c *CPU) { shiftRight(c, &c.h, true) }) }}, // 0x2c
		{"SRA L", func(c *CPU) { fetchCycle(c, func(c *CPU) { shiftRight(c, &c.l, true) }) }}, // 0x2d
		{"SRA (HL)", func(c *CPU) {
			c.ops.Push(func(c *CPU) {
				c.z = c.mmu.Read(c.hl())
			})
			c.ops.Push(func(c *CPU) {
				shiftRight(c, &c.z, true)
				c.mmu.Write(c.hl(), c.z)
			})
			fetchCycle(c)
		}}, // 0x2e
		{"SRA A", func(c *CPU) { fetchCycle(c, func(c *CPU) { shiftRight(c, &c.a, true) }) }}, // 0x2f
		{"SWAP B", func(c *CPU) { fetchCycle(c, func(c *CPU) { swap(c, &c.b) }) }},            // 0x30
		{"SWAP C", func(c *CPU) { fetchCycle(c, func(c *CPU) { swap(c, &c.c) }) }},            // 0x31
		{"SWAP D", func(c *CPU) { fetchCycle(c, func(c *CPU) { swap(c, &c.d) }) }},            // 0x32
		{"SWAP E", func(c *CPU) { fetchCycle(c, func(c *CPU) { swap(c, &c.e) }) }},            // 0x33
		{"SWAP H", func(c *CPU) { fetchCycle(c, func(c *CPU) { swap(c, &c.h) }) }},            // 0x34
		{"SWAP L", func(c *CPU) { fetchCycle(c, func(c *CPU) { swap(c, &c.l) }) }},            // 0x35
		{"SWAP (HL)", func(c *CPU) {
			c.ops.Push(func(c *CPU) {
				c.z = c.mmu.Read(c.hl())
			})
			c.ops.Push(func(c *CPU) {
				swap(c, &c.z)
				c.mmu.Write(c.hl(), c.z)
			})
			fetchCycle(c)
		}}, // 0x36
		{"SWAP A", func(c *CPU) { fetchCycle(c, func(c *CPU) { swap(c, &c.a) }) }},             // 0x37
		{"SRL B", func(c *CPU) { fetchCycle(c, func(c *CPU) { shiftRight(c, &c.b, false) }) }}, // 0x38
		{"SRL C", func(c *CPU) { fetchCycle(c, func(c *CPU) { shiftRight(c, &c.c, false) }) }}, // 0x39
		{"SRL D", func(c *CPU) { fetchCycle(c, func(c *CPU) { shiftRight(c, &c.d, false) }) }}, // 0x3a
		{"SRL E", func(c *CPU) { fetchCycle(c, func(c *CPU) { shiftRight(c, &c.e, false) }) }}, // 0x3b
		{"SRL H", func(c *CPU) { fetchCycle(c, func(c *CPU) { shiftRight(c, &c.h, false) }) }}, // 0x3c
		{"SRL L", func(c *CPU) { fetchCycle(c, func(c *CPU) { shiftRight(c, &c.l, false) }) }}, // 0x3d
		{"SRL (HL)", func(c *CPU) {
			c.ops.Push(func(c *CPU) {
				c.z = c.mmu.Read(c.hl())
			})
			c.ops.Push(func(c *CPU) {
				shiftRight(c, &c.z, false)
				c.mmu.Write(c.hl(), c.z)
			})
			fetchCycle(c)
		}}, // 0x3e
		{"SRL A", func(c *CPU) { fetchCycle(c, func(c *CPU) { shiftRight(c, &c.a, false) }) }}, // 0x3f
		{"BIT 0, B", func(c *CPU) { fetchCycle(c, func(c *CPU) { bit(c, c.b, 0) }) }},          // 0x40
		{"BIT 0, C", func(c *CPU) { fetchCycle(c, func(c *CPU) { bit(c, c.c, 0) }) }},          // 0x41
		{"BIT 0, D", func(c *CPU) { fetchCycle(c, func(c *CPU) { bit(c, c.d, 0) }) }},          // 0x42
		{"BIT 0, E", func(c *CPU) { fetchCycle(c, func(c *CPU) { bit(c, c.e, 0) }) }},          // 0x43
		{"BIT 0, H", func(c *CPU) { fetchCycle(c, func(c *CPU) { bit(c, c.h, 0) }) }},          // 0x44
		{"BIT 0, L", func(c *CPU) { fetchCycle(c, func(c *CPU) { bit(c, c.l, 0) }) }},          // 0x45
		{"BIT 0, (HL)", func(c *CPU) {
			c.ops.Push(func(c *CPU) {
				c.z = c.mmu.Read(c.hl())
				bit(c, c.z, 0)
			})
			fetchCycle(c)
		}}, // 0x46
		{"BIT 0, A", func(c *CPU) { fetchCycle(c, func(c *CPU) { bit(c, c.a, 0) }) }}, // 0x47
		{"BIT 1, B", func(c *CPU) { fetchCycle(c, func(c *CPU) { bit(c, c.b, 1) }) }}, // 0x48
		{"BIT 1, C", func(c *CPU) { fetchCycle(c, func(c *CPU) { bit(c, c.c, 1) }) }}, // 0x49
		{"BIT 1, D", func(c *CPU) { fetchCycle(c, func(c *CPU) { bit(c, c.d, 1) }) }}, // 0x4a
		{"BIT 1, E", func(c *CPU) { fetchCycle(c, func(c *CPU) { bit(c, c.e, 1) }) }}, // 0x4b
		{"BIT 1, H", func(c *CPU) { fetchCycle(c, func(c *CPU) { bit(c, c.h, 1) }) }}, // 0x4c
		{"BIT 1, L", func(c *CPU) { fetchCycle(c, func(c *CPU) { bit(c, c.l, 1) }) }}, // 0x4d
		{"BIT 1, (HL)", func(c *CPU) {
			c.ops.Push(func(c *CPU) {
				c.z = c.mmu.Read(c.hl())
				bit(c, c.z, 1)
			})
			fetchCycle(c)
		}}, // 0x4e
		{"BIT 1, A", func(c *CPU) { fetchCycle(c, func(c *CPU) { bit(c, c.a, 1) }) }}, // 0x4f
		{"BIT 2, B", func(c *CPU) { fetchCycle(c, func(c *CPU) { bit(c, c.b, 2) }) }}, // 0x50
		{"BIT 2, C", func(c *CPU) { fetchCycle(c, func(c *CPU) { bit(c, c.c, 2) }) }}, // 0x51
		{"BIT 2, D", func(c *CPU) { fetchCycle(c, func(c *CPU) { bit(c, c.d, 2) }) }}, // 0x52
		{"BIT 2, E", func(c *CPU) { fetchCycle(c, func(c *CPU) { bit(c, c.e, 2) }) }}, // 0x53
		{"BIT 2, H", func(c *CPU) { fetchCycle(c, func(c *CPU) { bit(c, c.h, 2) }) }}, // 0x54
		{"BIT 2, L", func(c *CPU) { fetchCycle(c, func(c *CPU) { bit(c, c.l, 2) }) }}, // 0x55
		{"BIT 2, (HL)", func(c *CPU) {
			c.ops.Push(func(c *CPU) {
				c.z = c.mmu.Read(c.hl())
				bit(c, c.z, 2)
			})
			fetchCycle(c)
		}}, // 0x56
		{"BIT 2, A", func(c *CPU) { fetchCycle(c, func(c *CPU) { bit(c, c.a, 2) }) }}, // 0x57
		{"BIT 3, B", func(c *CPU) { fetchCycle(c, func(c *CPU) { bit(c, c.b, 3) }) }}, // 0x58
		{"BIT 3, C", func(c *CPU) { fetchCycle(c, func(c *CPU) { bit(c, c.c, 3) }) }}, // 0x59
		{"BIT 3, D", func(c *CPU) { fetchCycle(c, func(c *CPU) { bit(c, c.d, 3) }) }}, // 0x5a
		{"BIT 3, E", func(c *CPU) { fetchCycle(c, func(c *CPU) { bit(c, c.e, 3) }) }}, // 0x5b
		{"BIT 3, H", func(c *CPU) { fetchCycle(c, func(c *CPU) { bit(c, c.h, 3) }) }}, // 0x5c
		{"BIT 3, L", func(c *CPU) { fetchCycle(c, func(c *CPU) { bit(c, c.l, 3) }) }}, // 0x5d
		{"BIT 3, (HL)", func(c *CPU) {
			c.ops.Push(func(c *CPU) {
				c.z = c.mmu.Read(c.hl())
				bit(c, c.z, 3)
			})
			fetchCycle(c)
		}}, // 0x5e
		{"BIT 3, A", func(c *CPU) { fetchCycle(c, func(c *CPU) { bit(c, c.a, 3) }) }}, // 0x5f
		{"BIT 4, B", func(c *CPU) { fetchCycle(c, func(c *CPU) { bit(c, c.b, 4) }) }}, // 0x60
		{"BIT 4, C", func(c *CPU) { fetchCycle(c, func(c *CPU) { bit(c, c.c, 4) }) }}, // 0x61
		{"BIT 4, D", func(c *CPU) { fetchCycle(c, func(c *CPU) { bit(c, c.d, 4) }) }}, // 0x62
		{"BIT 4, E", func(c *CPU) { fetchCycle(c, func(c *CPU) { bit(c, c.e, 4) }) }}, // 0x63
		{"BIT 4, H", func(c *CPU) { fetchCycle(c, func(c *CPU) { bit(c, c.h, 4) }) }}, // 0x64
		{"BIT 4, L", func(c *CPU) { fetchCycle(c, func(c *CPU) { bit(c, c.l, 4) }) }}, // 0x65
		{"BIT 4, (HL)", func(c *CPU) {
			c.ops.Push(func(c *CPU) {
				c.z = c.mmu.Read(c.hl())
				bit(c, c.z, 4)
			})
			fetchCycle(c)
		}}, // 0x66
		{"BIT 4, A", func(c *CPU) { fetchCycle(c, func(c *CPU) { bit(c, c.a, 4) }) }}, // 0x67
		{"BIT 5, B", func(c *CPU) { fetchCycle(c, func(c *CPU) { bit(c, c.b, 5) }) }}, // 0x68
		{"BIT 5, C", func(c *CPU) { fetchCycle(c, func(c *CPU) { bit(c, c.c, 5) }) }}, // 0x69
		{"BIT 5, D", func(c *CPU) { fetchCycle(c, func(c *CPU) { bit(c, c.d, 5) }) }}, // 0x6a
		{"BIT 5, E", func(c *CPU) { fetchCycle(c, func(c *CPU) { bit(c, c.e, 5) }) }}, // 0x6b
		{"BIT 5, H", func(c *CPU) { fetchCycle(c, func(c *CPU) { bit(c, c.h, 5) }) }}, // 0x6c
		{"BIT 5, L", func(c *CPU) { fetchCycle(c, func(c *CPU) { bit(c, c.l, 5) }) }}, // 0x6d
		{"BIT 5, (HL)", func(c *CPU) {
			c.ops.Push(func(c *CPU) {
				c.z = c.mmu.Read(c.hl())
				bit(c, c.z, 5)
			})
			fetchCycle(c)
		}}, // 0x6e
		{"BIT 5, A", func(c *CPU) { fetchCycle(c, func(c *CPU) { bit(c, c.a, 5) }) }}, // 0x6f
		{"BIT 6, B", func(c *CPU) { fetchCycle(c, func(c *CPU) { bit(c, c.b, 6) }) }}, // 0x70
		{"BIT 6, C", func(c *CPU) { fetchCycle(c, func(c *CPU) { bit(c, c.c, 6) }) }}, // 0x71
		{"BIT 6, D", func(c *CPU) { fetchCycle(c, func(c *CPU) { bit(c, c.d, 6) }) }}, // 0x72
		{"BIT 6, E", func(c *CPU) { fetchCycle(c, func(c *CPU) { bit(c, c.e, 6) }) }}, // 0x73
		{"BIT 6, H", func(c *CPU) { fetchCycle(c, func(c *CPU) { bit(c, c.h, 6) }) }}, // 0x74
		{"BIT 6, L", func(c *CPU) { fetchCycle(c, func(c *CPU) { bit(c, c.l, 6) }) }}, // 0x75
		{"BIT 6, (HL)", func(c *CPU) {
			c.ops.Push(func(c *CPU) {
				c.z = c.mmu.Read(c.hl())
				bit(c, c.z, 6)
			})
			fetchCycle(c)
		}}, // 0x76
		{"BIT 6, A", func(c *CPU) { fetchCycle(c, func(c *CPU) { bit(c, c.a, 6) }) }}, // 0x77
		{"BIT 7, B", func(c *CPU) { fetchCycle(c, func(c *CPU) { bit(c, c.b, 7) }) }}, // 0x78
		{"BIT 7, C", func(c *CPU) { fetchCycle(c, func(c *CPU) { bit(c, c.c, 7) }) }}, // 0x79
		{"BIT 7, D", func(c *CPU) { fetchCycle(c, func(c *CPU) { bit(c, c.d, 7) }) }}, // 0x7a
		{"BIT 7, E", func(c *CPU) { fetchCycle(c, func(c *CPU) { bit(c, c.e, 7) }) }}, // 0x7b
		{"BIT 7, H", func(c *CPU) { fetchCycle(c, func(c *CPU) { bit(c, c.h, 7) }) }}, // 0x7c
		{"BIT 7, L", func(c *CPU) { fetchCycle(c, func(c *CPU) { bit(c, c.l, 7) }) }}, // 0x7d
		{"BIT 7, (HL)", func(c *CPU) {
			c.ops.Push(func(c *CPU) {
				c.z = c.mmu.Read(c.hl())
				bit(c, c.z, 7)
			})
			fetchCycle(c)
		}}, // 0x7e
		{"BIT 7, A", func(c *CPU) { fetchCycle(c, func(c *CPU) { bit(c, c.a, 7) }) }}, // 0x7f
		{"RES 0, B", func(c *CPU) { fetchCycle(c, func(c *CPU) { res(&c.b, 0) }) }},   // 0x80
		{"RES 0, C", func(c *CPU) { fetchCycle(c, func(c *CPU) { res(&c.c, 0) }) }},   // 0x81
		{"RES 0, D", func(c *CPU) { fetchCycle(c, func(c *CPU) { res(&c.d, 0) }) }},   // 0x82
		{"RES 0, E", func(c *CPU) { fetchCycle(c, func(c *CPU) { res(&c.e, 0) }) }},   // 0x83
		{"RES 0, H", func(c *CPU) { fetchCycle(c, func(c *CPU) { res(&c.h, 0) }) }},   // 0x84
		{"RES 0, L", func(c *CPU) { fetchCycle(c, func(c *CPU) { res(&c.l, 0) }) }},   // 0x85
		{"RES 0, (HL)", func(c *CPU) {
			c.ops.Push(func(c *CPU) {
				c.z = c.mmu.Read(c.hl())
			})
			c.ops.Push(func(c *CPU) {
				res(&c.z, 0)
				c.mmu.Write(c.hl(), c.z)
			})
			fetchCycle(c)
		}}, // 0x86
		{"RES 0, A", func(c *CPU) { fetchCycle(c, func(c *CPU) { res(&c.a, 0) }) }}, // 0x87
		{"RES 1, B", func(c *CPU) { fetchCycle(c, func(c *CPU) { res(&c.b, 1) }) }}, // 0x88
		{"RES 1, C", func(c *CPU) { fetchCycle(c, func(c *CPU) { res(&c.c, 1) }) }}, // 0x89
		{"RES 1, D", func(c *CPU) { fetchCycle(c, func(c *CPU) { res(&c.d, 1) }) }}, // 0x8a
		{"RES 1, E", func(c *CPU) { fetchCycle(c, func(c *CPU) { res(&c.e, 1) }) }}, // 0x8b
		{"RES 1, H", func(c *CPU) { fetchCycle(c, func(c *CPU) { res(&c.h, 1) }) }}, // 0x8c
		{"RES 1, L", func(c *CPU) { fetchCycle(c, func(c *CPU) { res(&c.l, 1) }) }}, // 0x8d
		{"RES 1, (HL)", func(c *CPU) {
			c.ops.Push(func(c *CPU) {
				c.z = c.mmu.Read(c.hl())
			})
			c.ops.Push(func(c *CPU) {
				res(&c.z, 1)
				c.mmu.Write(c.hl(), c.z)
			})
			fetchCycle(c)
		}}, // 0x8e
		{"RES 1, A", func(c *CPU) { fetchCycle(c, func(c *CPU) { res(&c.a, 1) }) }}, // 0x8f
		{"RES 2, B", func(c *CPU) { fetchCycle(c, func(c *CPU) { res(&c.b, 2) }) }}, // 0x90
		{"RES 2, C", func(c *CPU) { fetchCycle(c, func(c *CPU) { res(&c.c, 2) }) }}, // 0x91
		{"RES 2, D", func(c *CPU) { fetchCycle(c, func(c *CPU) { res(&c.d, 2) }) }}, // 0x92
		{"RES 2, E", func(c *CPU) { fetchCycle(c, func(c *CPU) { res(&c.e, 2) }) }}, // 0x93
		{"RES 2, H", func(c *CPU) { fetchCycle(c, func(c *CPU) { res(&c.h, 2) }) }}, // 0x94
		{"RES 2, L", func(c *CPU) { fetchCycle(c, func(c *CPU) { res(&c.l, 2) }) }}, // 0x95
		{"RES 2, (HL)", func(c *CPU) {
			c.ops.Push(func(c *CPU) {
				c.z = c.mmu.Read(c.hl())
			})
			c.ops.Push(func(c *CPU) {
				res(&c.z, 2)
				c.mmu.Write(c.hl(), c.z)
			})
			fetchCycle(c)
		}}, // 0x96
		{"RES 2, A", func(c *CPU) { fetchCycle(c, func(c *CPU) { res(&c.a, 2) }) }}, // 0x97
		{"RES 3, B", func(c *CPU) { fetchCycle(c, func(c *CPU) { res(&c.b, 3) }) }}, // 0x98
		{"RES 3, C", func(c *CPU) { fetchCycle(c, func(c *CPU) { res(&c.c, 3) }) }}, // 0x99
		{"RES 3, D", func(c *CPU) { fetchCycle(c, func(c *CPU) { res(&c.d, 3) }) }}, // 0x9a
		{"RES 3, E", func(c *CPU) { fetchCycle(c, func(c *CPU) { res(&c.e, 3) }) }}, // 0x9b
		{"RES 3, H", func(c *CPU) { fetchCycle(c, func(c *CPU) { res(&c.h, 3) }) }}, // 0x9c
		{"RES 3, L", func(c *CPU) { fetchCycle(c, func(c *CPU) { res(&c.l, 3) }) }}, // 0x9d
		{"RES 3, (HL)", func(c *CPU) {
			c.ops.Push(func(c *CPU) {
				c.z = c.mmu.Read(c.hl())
			})
			c.ops.Push(func(c *CPU) {
				res(&c.z, 3)
				c.mmu.Write(c.hl(), c.z)
			})
			fetchCycle(c)
		}}, // 0x9e
		{"RES 3, A", func(c *CPU) { fetchCycle(c, func(c *CPU) { res(&c.a, 3) }) }}, // 0x9f
		{"RES 4, B", func(c *CPU) { fetchCycle(c, func(c *CPU) { res(&c.b, 4) }) }}, // 0xa0
		{"RES 4, C", func(c *CPU) { fetchCycle(c, func(c *CPU) { res(&c.c, 4) }) }}, // 0xa1
		{"RES 4, D", func(c *CPU) { fetchCycle(c, func(c *CPU) { res(&c.d, 4) }) }}, // 0xa2
		{"RES 4, E", func(c *CPU) { fetchCycle(c, func(c *CPU) { res(&c.e, 4) }) }}, // 0xa3
		{"RES 4, H", func(c *CPU) { fetchCycle(c, func(c *CPU) { res(&c.h, 4) }) }}, // 0xa4
		{"RES 4, L", func(c *CPU) { fetchCycle(c, func(c *CPU) { res(&c.l, 4) }) }}, // 0xa5
		{"RES 4, (HL)", func(c *CPU) {
			c.ops.Push(func(c *CPU) {
				c.z = c.mmu.Read(c.hl())
			})
			c.ops.Push(func(c *CPU) {
				res(&c.z, 4)
				c.mmu.Write(c.hl(), c.z)
			})
			fetchCycle(c)
		}}, // 0xa6
		{"RES 4, A", func(c *CPU) { fetchCycle(c, func(c *CPU) { res(&c.a, 4) }) }}, // 0xa7
		{"RES 5, B", func(c *CPU) { fetchCycle(c, func(c *CPU) { res(&c.b, 5) }) }}, // 0xa8
		{"RES 5, C", func(c *CPU) { fetchCycle(c, func(c *CPU) { res(&c.c, 5) }) }}, // 0xa9
		{"RES 5, D", func(c *CPU) { fetchCycle(c, func(c *CPU) { res(&c.d, 5) }) }}, // 0xaa
		{"RES 5, E", func(c *CPU) { fetchCycle(c, func(c *CPU) { res(&c.e, 5) }) }}, // 0xab
		{"RES 5, H", func(c *CPU) { fetchCycle(c, func(c *CPU) { res(&c.h, 5) }) }}, // 0xac
		{"RES 5, L", func(c *CPU) { fetchCycle(c, func(c *CPU) { res(&c.l, 5) }) }}, // 0xad
		{"RES 5, (HL)", func(c *CPU) {
			c.ops.Push(func(c *CPU) {
				c.z = c.mmu.Read(c.hl())
			})
			c.ops.Push(func(c *CPU) {
				res(&c.z, 5)
				c.mmu.Write(c.hl(), c.z)
			})
			fetchCycle(c)
		}}, // 0xae
		{"RES 5, A", func(c *CPU) { fetchCycle(c, func(c *CPU) { res(&c.a, 5) }) }}, // 0xaf
		{"RES 6, B", func(c *CPU) { fetchCycle(c, func(c *CPU) { res(&c.b, 6) }) }}, // 0xb0
		{"RES 6, C", func(c *CPU) { fetchCycle(c, func(c *CPU) { res(&c.c, 6) }) }}, // 0xb1
		{"RES 6, D", func(c *CPU) { fetchCycle(c, func(c *CPU) { res(&c.d, 6) }) }}, // 0xb2
		{"RES 6, E", func(c *CPU) { fetchCycle(c, func(c *CPU) { res(&c.e, 6) }) }}, // 0xb3
		{"RES 6, H", func(c *CPU) { fetchCycle(c, func(c *CPU) { res(&c.h, 6) }) }}, // 0xb4
		{"RES 6, L", func(c *CPU) { fetchCycle(c, func(c *CPU) { res(&c.l, 6) }) }}, // 0xb5
		{"RES 6, (HL)", func(c *CPU) {
			c.ops.Push(func(c *CPU) {
				c.z = c.mmu.Read(c.hl())
			})
			c.ops.Push(func(c *CPU) {
				res(&c.z, 6)
				c.mmu.Write(c.hl(), c.z)
			})
			fetchCycle(c)
		}}, // 0xb6
		{"RES 6, A", func(c *CPU) { fetchCycle(c, func(c *CPU) { res(&c.a, 6) }) }}, // 0xb7
		{"RES 7, B", func(c *CPU) { fetchCycle(c, func(c *CPU) { res(&c.b, 7) }) }}, // 0xb8
		{"RES 7, C", func(c *CPU) { fetchCycle(c, func(c *CPU) { res(&c.c, 7) }) }}, // 0xb9
		{"RES 7, D", func(c *CPU) { fetchCycle(c, func(c *CPU) { res(&c.d, 7) }) }}, // 0xba
		{"RES 7, E", func(c *CPU) { fetchCycle(c, func(c *CPU) { res(&c.e, 7) }) }}, // 0xbb
		{"RES 7, H", func(c *CPU) { fetchCycle(c, func(c *CPU) { res(&c.h, 7) }) }}, // 0xbc
		{"RES 7, L", func(c *CPU) { fetchCycle(c, func(c *CPU) { res(&c.l, 7) }) }}, // 0xbd
		{"RES 7, (HL)", func(c *CPU) {
			c.ops.Push(func(c *CPU) {
				c.z = c.mmu.Read(c.hl())
			})
			c.ops.Push(func(c *CPU) {
				res(&c.z, 7)
				c.mmu.Write(c.hl(), c.z)
			})
			fetchCycle(c)
		}}, // 0xbe
		{"RES 7, A", func(c *CPU) { fetchCycle(c, func(c *CPU) { res(&c.a, 7) }) }}, // 0xbf
		{"SET 0, B", func(c *CPU) { fetchCycle(c, func(c *CPU) { set(&c.b, 0) }) }}, // 0xc0
		{"SET 0, C", func(c *CPU) { fetchCycle(c, func(c *CPU) { set(&c.c, 0) }) }}, // 0xc1
		{"SET 0, D", func(c *CPU) { fetchCycle(c, func(c *CPU) { set(&c.d, 0) }) }}, // 0xc2
		{"SET 0, E", func(c *CPU) { fetchCycle(c, func(c *CPU) { set(&c.e, 0) }) }}, // 0xc3
		{"SET 0, H", func(c *CPU) { fetchCycle(c, func(c *CPU) { set(&c.h, 0) }) }}, // 0xc4
		{"SET 0, L", func(c *CPU) { fetchCycle(c, func(c *CPU) { set(&c.l, 0) }) }}, // 0xc5
		{"SET 0, (HL)", func(c *CPU) {
			c.ops.Push(func(c *CPU) {
				c.z = c.mmu.Read(c.hl())
			})
			c.ops.Push(func(c *CPU) {
				set(&c.z, 0)
				c.mmu.Write(c.hl(), c.z)
			})
			fetchCycle(c)
		}}, // 0xc6
		{"SET 0, A", func(c *CPU) { fetchCycle(c, func(c *CPU) { set(&c.a, 0) }) }}, // 0xc7
		{"SET 1, B", func(c *CPU) { fetchCycle(c, func(c *CPU) { set(&c.b, 1) }) }}, // 0xc8
		{"SET 1, C", func(c *CPU) { fetchCycle(c, func(c *CPU) { set(&c.c, 1) }) }}, // 0xc9
		{"SET 1, D", func(c *CPU) { fetchCycle(c, func(c *CPU) { set(&c.d, 1) }) }}, // 0xca
		{"SET 1, E", func(c *CPU) { fetchCycle(c, func(c *CPU) { set(&c.e, 1) }) }}, // 0xcb
		{"SET 1, H", func(c *CPU) { fetchCycle(c, func(c *CPU) { set(&c.h, 1) }) }}, // 0xcc
		{"SET 1, L", func(c *CPU) { fetchCycle(c, func(c *CPU) { set(&c.l, 1) }) }}, // 0xcd
		{"SET 1, (HL)", func(c *CPU) {
			c.ops.Push(func(c *CPU) {
				c.z = c.mmu.Read(c.hl())
			})
			c.ops.Push(func(c *CPU) {
				set(&c.z, 1)
				c.mmu.Write(c.hl(), c.z)
			})
			fetchCycle(c)
		}}, // 0xce
		{"SET 1, A", func(c *CPU) { fetchCycle(c, func(c *CPU) { set(&c.a, 1) }) }}, // 0xcf
		{"SET 2, B", func(c *CPU) { fetchCycle(c, func(c *CPU) { set(&c.b, 2) }) }}, // 0xd0
		{"SET 2, C", func(c *CPU) { fetchCycle(c, func(c *CPU) { set(&c.c, 2) }) }}, // 0xd1
		{"SET 2, D", func(c *CPU) { fetchCycle(c, func(c *CPU) { set(&c.d, 2) }) }}, // 0xd2
		{"SET 2, E", func(c *CPU) { fetchCycle(c, func(c *CPU) { set(&c.e, 2) }) }}, // 0xd3
		{"SET 2, H", func(c *CPU) { fetchCycle(c, func(c *CPU) { set(&c.h, 2) }) }}, // 0xd4
		{"SET 2, L", func(c *CPU) { fetchCycle(c, func(c *CPU) { set(&c.l, 2) }) }}, // 0xd5
		{"SET 2, (HL)", func(c *CPU) {
			c.ops.Push(func(c *CPU) {
				c.z = c.mmu.Read(c.hl())
			})
			c.ops.Push(func(c *CPU) {
				set(&c.z, 2)
				c.mmu.Write(c.hl(), c.z)
			})
			fetchCycle(c)
		}}, // 0xd6
		{"SET 2, A", func(c *CPU) { fetchCycle(c, func(c *CPU) { set(&c.a, 2) }) }}, // 0xd7
		{"SET 3, B", func(c *CPU) { fetchCycle(c, func(c *CPU) { set(&c.b, 3) }) }}, // 0xd8
		{"SET 3, C", func(c *CPU) { fetchCycle(c, func(c *CPU) { set(&c.c, 3) }) }}, // 0xd9
		{"SET 3, D", func(c *CPU) { fetchCycle(c, func(c *CPU) { set(&c.d, 3) }) }}, // 0xda
		{"SET 3, E", func(c *CPU) { fetchCycle(c, func(c *CPU) { set(&c.e, 3) }) }}, // 0xdb
		{"SET 3, H", func(c *CPU) { fetchCycle(c, func(c *CPU) { set(&c.h, 3) }) }}, // 0xdc
		{"SET 3, L", func(c *CPU) { fetchCycle(c, func(c *CPU) { set(&c.l, 3) }) }}, // 0xdd
		{"SET 3, (HL)", func(c *CPU) {
			c.ops.Push(func(c *CPU) {
				c.z = c.mmu.Read(c.hl())
			})
			c.ops.Push(func(c *CPU) {
				set(&c.z, 3)
				c.mmu.Write(c.hl(), c.z)
			})
			fetchCycle(c)
		}}, // 0xde
		{"SET 3, A", func(c *CPU) { fetchCycle(c, func(c *CPU) { set(&c.a, 3) }) }}, // 0xdf
		{"SET 4, B", func(c *CPU) { fetchCycle(c, func(c *CPU) { set(&c.b, 4) }) }}, // 0xe0
		{"SET 4, C", func(c *CPU) { fetchCycle(c, func(c *CPU) { set(&c.c, 4) }) }}, // 0xe1
		{"SET 4, D", func(c *CPU) { fetchCycle(c, func(c *CPU) { set(&c.d, 4) }) }}, // 0xe2
		{"SET 4, E", func(c *CPU) { fetchCycle(c, func(c *CPU) { set(&c.e, 4) }) }}, // 0xe3
		{"SET 4, H", func(c *CPU) { fetchCycle(c, func(c *CPU) { set(&c.h, 4) }) }}, // 0xe4
		{"SET 4, L", func(c *CPU) { fetchCycle(c, func(c *CPU) { set(&c.l, 4) }) }}, // 0xe5
		{"SET 4, (HL)", func(c *CPU) {
			c.ops.Push(func(c *CPU) {
				c.z = c.mmu.Read(c.hl())
			})
			c.ops.Push(func(c *CPU) {
				set(&c.z, 4)
				c.mmu.Write(c.hl(), c.z)
			})
			fetchCycle(c)
		}}, // 0xe6
		{"SET 4, A", func(c *CPU) { fetchCycle(c, func(c *CPU) { set(&c.a, 4) }) }}, // 0xe7
		{"SET 5, B", func(c *CPU) { fetchCycle(c, func(c *CPU) { set(&c.b, 5) }) }}, // 0xe8
		{"SET 5, C", func(c *CPU) { fetchCycle(c, func(c *CPU) { set(&c.c, 5) }) }}, // 0xe9
		{"SET 5, D", func(c *CPU) { fetchCycle(c, func(c *CPU) { set(&c.d, 5) }) }}, // 0xea
		{"SET 5, E", func(c *CPU) { fetchCycle(c, func(c *CPU) { set(&c.e, 5) }) }}, // 0xeb
		{"SET 5, H", func(c *CPU) { fetchCycle(c, func(c *CPU) { set(&c.h, 5) }) }}, // 0xec
		{"SET 5, L", func(c *CPU) { fetchCycle(c, func(c *CPU) { set(&c.l, 5) }) }}, // 0xed
		{"SET 5, (HL)", func(c *CPU) {
			c.ops.Push(func(c *CPU) {
				c.z = c.mmu.Read(c.hl())
			})
			c.ops.Push(func(c *CPU) {
				set(&c.z, 5)
				c.mmu.Write(c.hl(), c.z)
			})
			fetchCycle(c)
		}}, // 0xee
		{"SET 5, A", func(c *CPU) { fetchCycle(c, func(c *CPU) { set(&c.a, 5) }) }}, // 0xef
		{"SET 6, B", func(c *CPU) { fetchCycle(c, func(c *CPU) { set(&c.b, 6) }) }}, // 0xf0
		{"SET 6, C", func(c *CPU) { fetchCycle(c, func(c *CPU) { set(&c.c, 6) }) }}, // 0xf1
		{"SET 6, D", func(c *CPU) { fetchCycle(c, func(c *CPU) { set(&c.d, 6) }) }}, // 0xf2
		{"SET 6, E", func(c *CPU) { fetchCycle(c, func(c *CPU) { set(&c.e, 6) }) }}, // 0xf3
		{"SET 6, H", func(c *CPU) { fetchCycle(c, func(c *CPU) { set(&c.h, 6) }) }}, // 0xf4
		{"SET 6, L", func(c *CPU) { fetchCycle(c, func(c *CPU) { set(&c.l, 6) }) }}, // 0xf5
		{"SET 6, (HL)", func(c *CPU) {
			c.ops.Push(func(c *CPU) {
				c.z = c.mmu.Read(c.hl())
			})
			c.ops.Push(func(c *CPU) {
				set(&c.z, 6)
				c.mmu.Write(c.hl(), c.z)
			})
			fetchCycle(c)
		}}, // 0xf6
		{"SET 6, A", func(c *CPU) { fetchCycle(c, func(c *CPU) { set(&c.a, 6) }) }}, // 0xf7
		{"SET 7, B", func(c *CPU) { fetchCycle(c, func(c *CPU) { set(&c.b, 7) }) }}, // 0xf8
		{"SET 7, C", func(c *CPU) { fetchCycle(c, func(c *CPU) { set(&c.c, 7) }) }}, // 0xf9
		{"SET 7, D", func(c *CPU) { fetchCycle(c, func(c *CPU) { set(&c.d, 7) }) }}, // 0xfa
		{"SET 7, E", func(c *CPU) { fetchCycle(c, func(c *CPU) { set(&c.e, 7) }) }}, // 0xfb
		{"SET 7, H", func(c *CPU) { fetchCycle(c, func(c *CPU) { set(&c.h, 7) }) }}, // 0xfc
		{"SET 7, L", func(c *CPU) { fetchCycle(c, func(c *CPU) { set(&c.l, 7) }) }}, // 0xfd
		{"SET 7, (HL)", func(c *CPU) {
			c.ops.Push(func(c *CPU) {
				c.z = c.mmu.Read(c.hl())
			})
			c.ops.Push(func(c *CPU) {
				set(&c.z, 7)
				c.mmu.Write(c.hl(), c.z)
			})
			fetchCycle(c)
		}}, // 0xfe
		{"SET 7, A", func(c *CPU) { fetchCycle(c, func(c *CPU) { set(&c.a, 7) }) }}, // 0xff
	}
}

func shiftLeft(c *CPU, reg *byte) {
	// get the highest bit of value by masking it with 0b10000000 (aka 0x80 or 0d128) and shifting
	// it 7 bits down
	car := (*reg & 0x80) >> 7

	if car == 1 {
		c.f.setFlag(carry)
	} else {
		c.f.unsetFlag(carry)
	}

	// shift value 1bit up
	*reg <<= 1

	if *reg == 0 {
		c.f.setFlag(zero)
	} else {
		c.f.unsetFlag(zero)
	}

	c.f.unsetFlag(negative)
	c.f.unsetFlag(halfCarry)
}

func rotateLeft(c *CPU, reg *byte) {
	// get the highest bit of value by masking it with 0b10000000 (aka 0x80 or 0d128) and shifting
	// it 7 bits down
	car := (*reg & 0x80) >> 7

	if car == 1 {
		c.f.setFlag(carry)
	} else {
		c.f.unsetFlag(carry)
	}

	// shift value 1bit up
	*reg <<= 1
	// append carry at its end
	*reg += car

	if *reg == 0 {
		c.f.setFlag(zero)
	} else {
		c.f.unsetFlag(zero)
	}

	c.f.unsetFlag(negative)
	c.f.unsetFlag(halfCarry)
}

func rotateLeftThroughCarry(c *CPU, reg *byte) {
	carryFlagWasSet := c.f.isSet(carry)

	// get the highest bit of value by masking it with 0b10000000 (aka 0x80 or 0d128) and shifting
	// it 7 bits down
	car := (*reg & 0x80) >> 7

	if car == 1 {
		c.f.setFlag(carry)
	} else {
		c.f.unsetFlag(carry)
	}

	// shift value 1bit up
	*reg <<= 1
	if carryFlagWasSet {
		*reg++
	}

	if *reg == 0 {
		c.f.setFlag(zero)
	} else {
		c.f.unsetFlag(zero)
	}

	c.f.unsetFlag(negative)
	c.f.unsetFlag(halfCarry)
}

func shiftRight(c *CPU, reg *byte, arithmetic bool) {
	highestBit := *reg & 0x80

	if util.BitIsSet8(*reg, 0) {
		c.f.setFlag(carry)
	} else {
		c.f.unsetFlag(carry)
	}

	// shift value 1bit down
	*reg >>= 1
	if arithmetic {
		// highest bit must stay unchanged
		*reg += highestBit
	}

	if *reg == 0 {
		c.f.setFlag(zero)
	} else {
		c.f.unsetFlag(zero)
	}

	c.f.unsetFlag(negative)
	c.f.unsetFlag(halfCarry)
}

func rotateRight(c *CPU, reg *byte) {
	car := *reg & 0x01

	if car == 1 {
		c.f.setFlag(carry)
	} else {
		c.f.unsetFlag(carry)
	}

	// shift value 1bit down
	*reg >>= 1
	*reg += car << 7

	if *reg == 0 {
		c.f.setFlag(zero)
	} else {
		c.f.unsetFlag(zero)
	}

	c.f.unsetFlag(negative)
	c.f.unsetFlag(halfCarry)
}

func rotateRightThroughCarry(c *CPU, reg *byte) {
	carryFlagWasSet := c.f.isSet(carry)

	if util.BitIsSet8(*reg, 0) {
		c.f.setFlag(carry)
	} else {
		c.f.unsetFlag(carry)
	}

	// shift value 1bit down
	*reg >>= 1
	if carryFlagWasSet {
		*reg += 1 << 7
	}

	if *reg == 0 {
		c.f.setFlag(zero)
	} else {
		c.f.unsetFlag(zero)
	}

	c.f.unsetFlag(negative)
	c.f.unsetFlag(halfCarry)
}

func swap(c *CPU, reg *byte) {
	c.f.unsetFlag(negative)
	c.f.unsetFlag(halfCarry)
	c.f.unsetFlag(carry)

	if *reg == 0 {
		c.f.setFlag(zero)
		return
	} else {
		c.f.unsetFlag(zero)
	}

	upperNibble := *reg & 0xf0
	lowerNibble := *reg & 0x0f

	*reg = upperNibble>>4 | lowerNibble<<4
}

func bit(c *CPU, reg byte, index byte) {
	if !util.BitIsSet8(reg, index) {
		c.f.setFlag(zero)
	} else {
		c.f.unsetFlag(zero)
	}

	c.f.setFlag(halfCarry)
	c.f.unsetFlag(negative)
}

func res(reg *byte, index byte) {
	*reg &= ^(byte(1) << index)
}

func set(reg *byte, index byte) {
	*reg |= byte(1) << index
}
