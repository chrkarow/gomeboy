package cpu

import (
	"gameboy-emulator/internal/cycle/interrupts"
	"gameboy-emulator/internal/util"
)

var instructions [256]instruction

// Do initialization of array in init() to prevent circular dependencies.
func init() {
	instructions = [256]instruction{
		{"NOP", nop},                    // 0x00
		{"LD BC, nn", ldBCd16},          // 0x01
		{"LD (BC), A", ldBCpA},          // 0x02
		{"INC BC", incBC},               // 0x03
		{"INC B", incB},                 // 0x04
		{"DEC B", decB},                 // 0x05
		{"LD B, n", ldBd8},              // 0x06
		{"RLCA", rlca},                  // 0x07
		{"LD (nn), SP", ldd16pSP},       // 0x08
		{"ADD HL, BC", addHLBC},         // 0x09
		{"LD A, (BC)", ldABCp},          // 0x0A
		{"DEC BC", decBC},               // 0x0B
		{"INC C", incC},                 // 0x0C
		{"DEC C", decC},                 // 0x0D
		{"LD C, n", ldCd8},              // 0x0E
		{"RRCA", rrca},                  // 0x0F
		{"STOP", stop},                  // 0x10
		{"LD DE, nn", ldDEd16},          // 0x11
		{"LD (DE), A", ldDEpA},          // 0x12
		{"INC DE", incDE},               // 0x13
		{"INC D", incD},                 // 0x14
		{"DEC D", decD},                 // 0x15
		{"LD D, n", ldDd8},              // 0x16
		{"RLA", rla},                    // 0x17
		{"JR e", jre},                   // 0x18
		{"ADD HL, DE", addHLDE},         // 0x19
		{"LD A, (DE)", ldADEp},          // 0x1A
		{"DEC DE", decDE},               // 0x1B
		{"INC E", incE},                 // 0x1C
		{"DEC E", decE},                 // 0x1D
		{"LD E, n", ldEd8},              // 0x1E
		{"RRA", rra},                    // 0x1F
		{"JR NZ, e", jrnze},             // 0x20
		{"LD HL, nn", ldHLd16},          // 0x21
		{"LD (HL+), A", ldiHLpA},        // 0x22
		{"INC HL", incHL},               // 0x23
		{"INC H", incH},                 // 0x24
		{"DEC H", decH},                 // 0x25
		{"LD H, n", ldHd8},              // 0x26
		{"DAA", daa},                    // 0x27
		{"JR Z, e", jrze},               // 0x28
		{"ADD HL, HL", addHLHL},         // 0x29
		{"LD A, (HL+)", ldiAHLp},        // 0x2A
		{"DEC HL", decHL},               // 0x2B
		{"INC L", incL},                 // 0x2C
		{"DEC L", decL},                 // 0x2D
		{"LD L, n", ldLd8},              // 0x2E
		{"CPL", cpl},                    // 0x2F
		{"JR NC, e", jrnce},             // 0x30
		{"LD SP, nn", ldSPd16},          // 0x31
		{"LD (HL-), A", lddHLpA},        // 0x32
		{"INC SP", incSP},               // 0x33
		{"INC (HL)", incHLp},            // 0x34
		{"DEC (HL)", decHLp},            // 0x35
		{"LD (HL), n", ldHLpd8},         // 0x36
		{"SCF", scf},                    // 0x37
		{"JR C, e", jrce},               // 0x38
		{"ADD HL, SP", addHLSP},         // 0x39
		{"LD A, (HL-)", lddAHLp},        // 0x3A
		{"DEC SP", decSP},               // 0x3B
		{"INC A", incA},                 // 0x3C
		{"DEC A", decA},                 // 0x3D
		{"LD A, n", ldAd8},              // 0x3E
		{"CCF", ccf},                    // 0x3F
		{"LD B, B", nop},                // 0x40
		{"LD B, C", ldBC},               // 0x41
		{"LD B, D", ldBD},               // 0x42
		{"LD B, E", ldBE},               // 0x43
		{"LD B, H", ldBH},               // 0x44
		{"LD B, L", ldBL},               // 0x45
		{"LD B, (HL)", ldBHLp},          // 0x46
		{"LD B, A", ldBA},               // 0x47
		{"LD C, B", ldCB},               // 0x48
		{"LD C, C", nop},                // 0x49
		{"LD C, D", ldCD},               // 0x4A
		{"LD C, E", ldCE},               // 0x4B
		{"LD C, H", ldCH},               // 0x4C
		{"LD C, L", ldCL},               // 0x4D
		{"LD C, (HL)", ldCHLp},          // 0x4E
		{"LD C, A", ldCA},               // 0x4F
		{"LD D, B", ldDB},               // 0x50
		{"LD D, C", ldDC},               // 0x51
		{"LD D, D", nop},                // 0x52
		{"LD D, E", ldDE},               // 0x53
		{"LD D, H", ldDH},               // 0x54
		{"LD D, L", ldDL},               // 0x55
		{"LD D, (HL)", ldDHLp},          // 0x56
		{"LD D, A", ldDA},               // 0x57
		{"LD E, B", ldEB},               // 0x58
		{"LD E, C", ldEC},               // 0x59
		{"LD E, D", ldED},               // 0x5A
		{"LD E, E", nop},                // 0x5B
		{"LD E, H", ldEH},               // 0x5C
		{"LD E, L", ldEL},               // 0x5D
		{"LD E, (HL)", ldEHLp},          // 0x5E
		{"LD E, A", ldEA},               // 0x5F
		{"LD H, B", ldHB},               // 0x60
		{"LD H, C", ldHC},               // 0x61
		{"LD H, D", ldHD},               // 0x62
		{"LD H, E", ldHE},               // 0x63
		{"LD H, H", nop},                // 0x64
		{"LD H, L", ldHL},               // 0x65
		{"LD H, (HL)", ldHHLp},          // 0x66
		{"LD H, A", ldHA},               // 0x67
		{"LD L, B", ldLB},               // 0x68
		{"LD L, C", ldLC},               // 0x69
		{"LD L, D", ldLD},               // 0x6A
		{"LD L, E", ldLE},               // 0x6B
		{"LD L, H", ldLH},               // 0x6C
		{"LD L, L", nop},                // 0x6D
		{"LD L, (HL)", ldLHLp},          // 0x6E
		{"LD L, A", ldLA},               // 0x6F
		{"LD (HL), B", ldHLpB},          // 0x70
		{"LD (HL), C", ldHLpC},          // 0x71
		{"LD (HL), D", ldHLpD},          // 0x72
		{"LD (HL), E", ldHLpE},          // 0x73
		{"LD (HL), H", ldHLpH},          // 0x74
		{"LD (HL), L", ldHLpL},          // 0x75
		{"HALT", halt},                  // 0x76
		{"LD (HL), A", ldHLpA},          // 0x77
		{"LD A, B", ldAB},               // 0x78
		{"LD A, C", ldAC},               // 0x79
		{"LD A, D", ldAD},               // 0x7A
		{"LD A, E", ldAE},               // 0x7B
		{"LD A, H", ldAH},               // 0x7C
		{"LD A, L", ldAL},               // 0x7D
		{"LD A, (HL)", ldAHLp},          // 0x7E
		{"LD A, A", nop},                // 0x7F
		{"ADD B", addB},                 // 0x80
		{"ADD C", addC},                 // 0x81
		{"ADD D", addD},                 // 0x82
		{"ADD E", addE},                 // 0x83
		{"ADD H", addH},                 // 0x84
		{"ADD L", addL},                 // 0x85
		{"ADD (HL)", addHLp},            // 0x86
		{"ADD A", addA},                 // 0x87
		{"ADC B", adcB},                 // 0x88
		{"ADC C", adcC},                 // 0x89
		{"ADC D", adcD},                 // 0x8A
		{"ADC E", adcE},                 // 0x8B
		{"ADC H", adcH},                 // 0x8C
		{"ADC L", adcL},                 // 0x8D
		{"ADC (HL)", adcHLp},            // 0x8E
		{"ADC A", adcA},                 // 0x8F
		{"SUB B", subB},                 // 0x90
		{"SUB C", subC},                 // 0x91
		{"SUB D", subD},                 // 0x92
		{"SUB E", subE},                 // 0x93
		{"SUB H", subH},                 // 0x94
		{"SUB L", subL},                 // 0x95
		{"SUB (HL)", subHLp},            // 0x96
		{"SUB A", subA},                 // 0x97
		{"SBC B", sbcB},                 // 0x98
		{"SBC C", sbcC},                 // 0x99
		{"SBC D", sbcD},                 // 0x9A
		{"SBC E", sbcE},                 // 0x9B
		{"SBC H", sbcH},                 // 0x9C
		{"SBC L", sbcL},                 // 0x9D
		{"SBC (HL)", sbcHLp},            // 0x9E
		{"SBC A", sbcA},                 // 0x9F
		{"AND B", andB},                 // 0xA0
		{"AND C", andC},                 // 0xA1
		{"AND D", andD},                 // 0xA2
		{"AND E", andE},                 // 0xA3
		{"AND H", andH},                 // 0xA4
		{"AND L", andL},                 // 0xA5
		{"AND (HL)", andHLp},            // 0xA6
		{"AND A", andA},                 // 0xA7
		{"XOR B", xorB},                 // 0xA8
		{"XOR C", xorC},                 // 0xA9
		{"XOR D", xorD},                 // 0xAA
		{"XOR E", xorE},                 // 0xAB
		{"XOR H", xorH},                 // 0xAC
		{"XOR L", xorL},                 // 0xAD
		{"XOR (HL)", xorHLp},            // 0xAE
		{"XOR A", xorA},                 // 0xAF
		{"OR B", orB},                   // 0xB0
		{"OR C", orC},                   // 0xB1
		{"OR D", orD},                   // 0xB2
		{"OR E", orE},                   // 0xB3
		{"OR H", orH},                   // 0xB4
		{"OR L", orL},                   // 0xB5
		{"OR (HL)", orHLp},              // 0xB6
		{"OR A", orA},                   // 0xB7
		{"CP B", cpB},                   // 0xB8
		{"CP C", cpC},                   // 0xB9
		{"CP D", cpD},                   // 0xBA
		{"CP E", cpE},                   // 0xBB
		{"CP H", cpH},                   // 0xBC
		{"CP L", cpL},                   // 0xBD
		{"CP (HL)", cpHLp},              // 0xBE
		{"CP A", cpA},                   // 0xBF
		{"RET NZ", retnz},               // 0xC0
		{"POP BC", popBC},               // 0xC1
		{"JP NZ, nn", jpnzd16},          // 0xC2
		{"JP nn", jpd16},                // 0xC3
		{"CALL NZ, nn", callnzd16},      // 0xC4
		{"PUSH BC", pushBC},             // 0xC5
		{"ADD n", addd8},                // 0xC6
		{"RST 0x00", rst00},             // 0xC7
		{"RET Z", retz},                 // 0xC8
		{"RET", ret},                    // 0xC9
		{"JP Z, nn", jpzd16},            // 0xCA
		{"CB", cb},                      // 0xCB
		{"CALL w, nn", callzd16},        // 0xCC
		{"CALL nn", calld16},            // 0xCD
		{"ADC n", adcd8},                // 0xCE
		{"RST 0x08", rst08},             // 0xCF
		{"RET NC", retnc},               // 0xD0
		{"POP DE", popDE},               // 0xD1
		{"JP NC, nn", jpncd16},          // 0xD2
		{"UNDEFINED", nil},              // 0xD3
		{"CALL NC, nn", callncd16},      // 0xD4
		{"PUSH DE", pushDE},             // 0xD5
		{"SUB n", subd8},                // 0xD6
		{"RST 0x10", rst10},             // 0xD7
		{"RET C", retc},                 // 0xD8
		{"RETI", reti},                  // 0xD9
		{"JP C, nn", jpcd16},            // 0xDA
		{"UNDEFINED", nil},              // 0xDB
		{"CALL C, nn", callcd16},        // 0xDC
		{"UNDEFINED", nil},              // 0xDD
		{"SBC n", sbcd8},                // 0xDE
		{"RST 0x18", rst18},             // 0xDF
		{"LD (0xFF00+n), A", ldff00d8A}, // 0xE0
		{"POP HL", popHL},               // 0xE1
		{"LD (0xFF00+C), A", ldff00CA},  // 0xE2
		{"UNDEFINED", nil},              // 0xE3
		{"UNDEFINED", nil},              // 0xE4
		{"PUSH HL", pushHL},             // 0xE5
		{"AND n", andd8},                // 0xE6
		{"RST 0x20", rst20},             // 0xE7
		{"ADD SP, e", addSPr8},          // 0xE8
		{"JP HL", jpHL},                 // 0xE9
		{"LD (nn), A", ldd16pA},         // 0xEA
		{"UNDEFINED", nil},              // 0xEB
		{"UNDEFINED", nil},              // 0xEC
		{"UNDEFINED", nil},              // 0xED
		{"XOR n", xord8},                // 0xEE
		{"RST 0x28", rst28},             // 0xEF
		{"LD A, (0xFF00+n)", ldAff00d8}, // 0xF0
		{"POP AF", popAF},               // 0xF1
		{"LD A, (0xFF00+C)", ldAff00C},  // 0xF2
		{"DI", di},                      // 0xF3
		{"UNDEFINED", nil},              // 0xF4
		{"PUSH AF", pushAF},             // 0xF5
		{"OR n", ord8},                  // 0xF6
		{"RST 0x30", rst30},             // 0xF7
		{"LD HL, SP+e", ldHLSPr8},       // 0xF8
		{"LD SP, HL", ldSPHL},           // 0xF9
		{"LD A, (nn)", ldAd16p},         // 0xFA
		{"EI", ei},                      // 0xFB
		{"UNDEFINED", nil},              // 0xFC
		{"UNDEFINED", nil},              // 0xFD
		{"CP n", cpd8},                  // 0xFE
		{"RST 0x38", rst38},             // 0xFF
	}
}

// 0x00
func nop(c *CPU) {
	fetchCycle(c)
}

// 0x01
func ldBCd16(c *CPU) {
	ldrrd16(c, &c.b, &c.c)
}

// 0x02
func ldBCpA(c *CPU) {
	ldpr(c, c.bc(), c.a)
}

// 0x03
func incBC(c *CPU) {
	c.ops.Push(func(c *CPU) {
		c.writeBC(c.bc() + 1)
	})
	fetchCycle(c)
}

// 0x04
func incB(c *CPU) {
	incr(c, &c.b)
}

// 0x05
func decB(c *CPU) {
	decr(c, &c.b)
}

// 0x06
func ldBd8(c *CPU) {
	ldrd8(c, &c.b)
}

// 0x07
func rlca(c *CPU) {
	fetchCycle(c, func(c *CPU) {
		c.f.unsetFlag(halfCarry)
		c.f.unsetFlag(zero)
		c.f.unsetFlag(negative)

		// check bit 7 to determine carry flag after rotation
		car := c.a&0x80 != 0
		if car {
			c.f.setFlag(carry)
		} else {
			c.f.unsetFlag(carry)
		}

		c.a <<= 1
		if car {
			c.a++
		}
	})
}

// 0x08
func ldd16pSP(c *CPU) {
	readd16(c)

	c.ops.Push(func(c *CPU) {
		c.mmu.Write(c.wz(), byte(c.sp))
	})

	c.ops.Push(func(c *CPU) {
		c.mmu.Write(c.wz()+1, byte(c.sp>>8))
	})

	fetchCycle(c)
}

// 0x09
func addHLBC(c *CPU) {
	addHLrr(c, c.b, c.c)
}

// 0x0A
func ldABCp(c *CPU) {
	ldrp(c, &c.a, c.bc())
}

// 0x0B
func decBC(c *CPU) {
	c.ops.Push(func(c *CPU) {
		c.writeBC(c.bc() - 1)
	})

	fetchCycle(c)
}

// 0x0C
func incC(c *CPU) {
	incr(c, &c.c)
}

// 0x0D
func decC(c *CPU) {
	decr(c, &c.c)
}

// 0x0E
func ldCd8(c *CPU) {
	ldrd8(c, &c.c)
}

// 0x0F
func rrca(c *CPU) {
	fetchCycle(c, func(c *CPU) {
		c.f.unsetFlag(halfCarry)
		c.f.unsetFlag(zero)
		c.f.unsetFlag(negative)

		// check bit 0 to determine carry flag after rotation
		car := c.a&0x01 != 0
		if car {
			c.f.setFlag(carry)
		} else {
			c.f.unsetFlag(carry)
		}

		c.a >>= 1
		if car {
			c.a |= 0x80
		}
	})
}

// 0x10
// Should skip a byte after it
func stop(c *CPU) {
	c.ops.Push(func(c *CPU) {
		c.interrupts.SetMasterEnable(false)
		c.state = stopped
		c.pc++
	})
}

// 0x11
func ldDEd16(c *CPU) {
	ldrrd16(c, &c.d, &c.e)
}

// 0x12
func ldDEpA(c *CPU) {
	ldpr(c, c.de(), c.a)
}

// 0x13
func incDE(c *CPU) {
	c.ops.Push(func(c *CPU) {
		c.writeDE(c.de() + 1)
	})
	fetchCycle(c)
}

// 0x14
func incD(c *CPU) {
	incr(c, &c.d)
}

// 0x15
func decD(c *CPU) {
	decr(c, &c.d)
}

// 0x16
func ldDd8(c *CPU) {
	ldrd8(c, &c.d)
}

// 0x17
func rla(c *CPU) {
	fetchCycle(c, func(c *CPU) {
		c.f.unsetFlag(halfCarry)
		c.f.unsetFlag(zero)
		c.f.unsetFlag(negative)

		oldCarry := c.f.isSet(carry)

		// check bit 7 to determine carry flag after rotation
		car := c.a&0x80 != 0
		if car {
			c.f.setFlag(carry)
		} else {
			c.f.unsetFlag(carry)
		}

		c.a <<= 1
		if oldCarry {
			c.a++
		}
	})
}

// 0x18
func jre(c *CPU) {
	jrcc(c, true)
}

// 0x19
func addHLDE(c *CPU) {
	addHLrr(c, c.d, c.e)
}

// 0x1A
func ldADEp(c *CPU) {
	ldrp(c, &c.a, c.de())
}

// 0x1B
func decDE(c *CPU) {
	c.ops.Push(func(c *CPU) {
		c.writeDE(c.de() - 1)
	})
	fetchCycle(c)
}

// 0x1C
func incE(c *CPU) {
	incr(c, &c.e)
}

// 0x1D
func decE(c *CPU) {
	decr(c, &c.e)
}

// 0x1E
func ldEd8(c *CPU) {
	ldrd8(c, &c.e)
}

// 0x1F
func rra(c *CPU) {
	fetchCycle(c, func(c *CPU) {
		c.f.unsetFlag(halfCarry)
		c.f.unsetFlag(zero)
		c.f.unsetFlag(negative)

		oldCarry := c.f.isSet(carry)

		// check bit 0 to determine carry flag after rotation
		car := c.a&0x01 != 0
		if car {
			c.f.setFlag(carry)
		} else {
			c.f.unsetFlag(carry)
		}

		c.a >>= 1
		if oldCarry {
			c.a |= 0x80
		}
	})
}

// 0x20
func jrnze(c *CPU) {
	jrcc(c, !c.f.isSet(zero))
}

// 0x21
func ldHLd16(c *CPU) {
	ldrrd16(c, &c.h, &c.l)
}

// 0x22
func ldiHLpA(c *CPU) {
	c.ops.Push(func(c *CPU) {
		c.mmu.Write(c.hl(), c.a)
		c.writeHL(c.hl() + 1)
	})
	fetchCycle(c)
}

// 0x23
func incHL(c *CPU) {
	c.ops.Push(func(c *CPU) {
		c.writeHL(c.hl() + 1)
	})
	fetchCycle(c)
}

// 0x24
func incH(c *CPU) {
	incr(c, &c.h)
}

// 0x25
func decH(c *CPU) {
	decr(c, &c.h)
}

// 0x26
func ldHd8(c *CPU) {
	ldrd8(c, &c.h)
}

// 0x27
// Implemented according to https://ehaskins.com/2018-01-30%20Z80%20DAA/
func daa(c *CPU) {
	fetchCycle(c, func(c *CPU) {
		var correction byte

		value := c.a

		if c.f.isSet(halfCarry) || (!c.f.isSet(negative) && value&0xf > 0x9) {
			correction |= 0x6
		}

		if c.f.isSet(carry) || (!c.f.isSet(negative) && value > 0x99) {
			correction |= 0x60
			c.f.setFlag(carry)
		}

		if c.f.isSet(negative) {
			value -= correction
		} else {
			value += correction
		}

		if value == 0 {
			c.f.setFlag(zero)
		} else {
			c.f.unsetFlag(zero)
		}

		c.f.unsetFlag(halfCarry)

		c.a = value
	})
}

// 0x28
func jrze(c *CPU) {
	jrcc(c, c.f.isSet(zero))
}

// 0x29
func addHLHL(c *CPU) {
	addHLrr(c, c.h, c.l)
}

// 0x2A
func ldiAHLp(c *CPU) {
	c.ops.Push(func(c *CPU) {
		c.a = c.mmu.Read(c.hl())
		c.writeHL(c.hl() + 1)
	})
	fetchCycle(c)
}

// 0x2B
func decHL(c *CPU) {
	c.ops.Push(func(c *CPU) {
		c.writeHL(c.hl() - 1)
	})
	fetchCycle(c)
}

// 0x2C
func incL(c *CPU) {
	incr(c, &c.l)
}

// 0x2D
func decL(c *CPU) {
	decr(c, &c.l)
}

// 0x2E
func ldLd8(c *CPU) {
	ldrd8(c, &c.l)
}

// 0x2F
func cpl(c *CPU) {
	fetchCycle(c, func(c *CPU) {
		c.a = ^c.a
		c.f.setFlag(negative)
		c.f.setFlag(halfCarry)
	})
}

// 0x30
func jrnce(c *CPU) {
	jrcc(c, !c.f.isSet(carry))
}

// 0x31
func ldSPd16(c *CPU) {
	readd16(c)
	fetchCycle(c, func(c *CPU) {
		c.sp = c.wz()
	})
}

// 0x32
func lddHLpA(c *CPU) {
	c.ops.Push(func(c *CPU) {
		c.mmu.Write(c.hl(), c.a)
		c.writeHL(c.hl() - 1)
	})
	fetchCycle(c)
}

// 0x33
func incSP(c *CPU) {
	c.ops.Push(func(c *CPU) {
		c.sp++
	})
	fetchCycle(c)
}

// 0x34
func incHLp(c *CPU) {
	c.ops.Push(func(c *CPU) {
		c.z = c.mmu.Read(c.hl())
	})

	// can't use incrr because the timing is different
	c.ops.Push(func(c *CPU) {
		c.f.unsetFlag(negative)

		// overflow from bit 3 to 4
		if c.z&0x0F == 0x0F {
			c.f.setFlag(halfCarry)
		} else {
			c.f.unsetFlag(halfCarry)
		}

		c.z++

		if c.z == 0x00 {
			c.f.setFlag(zero)
		} else {
			c.f.unsetFlag(zero)
		}

		c.mmu.Write(c.hl(), c.z)
	})

	fetchCycle(c)
}

// 0x35
func decHLp(c *CPU) {
	c.ops.Push(func(c *CPU) {
		c.z = c.mmu.Read(c.hl())
	})

	// can't use decrr because the timing is different
	c.ops.Push(func(c *CPU) {
		c.f.setFlag(negative)

		// decrement will wrap around lower nibble
		if c.z&0xF == 0x00 {
			c.f.setFlag(halfCarry)
		} else {
			c.f.unsetFlag(halfCarry)
		}

		c.z--

		if c.z == 0x00 {
			c.f.setFlag(zero)
		} else {
			c.f.unsetFlag(zero)
		}

		c.mmu.Write(c.hl(), c.z)
	})

	fetchCycle(c)
}

// 0x36
func ldHLpd8(c *CPU) {
	readd8(c)

	c.ops.Push(func(c *CPU) {
		c.mmu.Write(c.hl(), c.z)
	})

	fetchCycle(c)
}

// 0x37
func scf(c *CPU) {
	fetchCycle(c, func(c *CPU) {
		c.f.setFlag(carry)
		c.f.unsetFlag(negative)
		c.f.unsetFlag(halfCarry)
	})
}

// 0x38
func jrce(c *CPU) {
	jrcc(c, c.f.isSet(carry))
}

// 0x39
func addHLSP(c *CPU) {
	addHLrr(c, byte(c.sp>>8), byte(c.sp))
}

// 0x3A
func lddAHLp(c *CPU) {
	c.ops.Push(func(c *CPU) {
		c.a = c.mmu.Read(c.hl())
		c.writeHL(c.hl() - 1)
	})
	fetchCycle(c)
}

// 0x3B
func decSP(c *CPU) {
	c.ops.Push(func(c *CPU) {
		c.sp--
	})
	fetchCycle(c)
}

// 0x3C
func incA(c *CPU) {
	incr(c, &c.a)
}

// 0x3D
func decA(c *CPU) {
	decr(c, &c.a)
}

// 0x3E
func ldAd8(c *CPU) {
	ldrd8(c, &c.a)
}

// 0x3F
func ccf(c *CPU) {
	fetchCycle(c, func(c *CPU) {
		c.f.unsetFlag(negative)
		c.f.unsetFlag(halfCarry)
		if c.f.isSet(carry) {
			c.f.unsetFlag(carry)
		} else {
			c.f.setFlag(carry)
		}
	})
}

// 0x40 NOP

// 0x41
func ldBC(c *CPU) {
	fetchCycle(c, func(c *CPU) { c.b = c.c })
}

// 0x42
func ldBD(c *CPU) {
	fetchCycle(c, func(c *CPU) { c.b = c.d })
}

// 0x43
func ldBE(c *CPU) {
	fetchCycle(c, func(c *CPU) { c.b = c.e })
}

// 0x44
func ldBH(c *CPU) {
	fetchCycle(c, func(c *CPU) { c.b = c.h })
}

// 0x45
func ldBL(c *CPU) {
	fetchCycle(c, func(c *CPU) { c.b = c.l })
}

// 0x46
func ldBHLp(c *CPU) {
	ldrp(c, &c.b, c.hl())
}

// 0x47
func ldBA(c *CPU) {
	fetchCycle(c, func(c *CPU) { c.b = c.a })
}

// 0x48
func ldCB(c *CPU) {
	fetchCycle(c, func(c *CPU) { c.c = c.b })
}

// 0x49 NOP

// 0x4A
func ldCD(c *CPU) {
	fetchCycle(c, func(c *CPU) { c.c = c.d })
}

// 0x4B
func ldCE(c *CPU) {
	fetchCycle(c, func(c *CPU) { c.c = c.e })
}

// 0x4C
func ldCH(c *CPU) {
	fetchCycle(c, func(c *CPU) { c.c = c.h })
}

// 0x4D
func ldCL(c *CPU) {
	fetchCycle(c, func(c *CPU) { c.c = c.l })
}

// 0x4E
func ldCHLp(c *CPU) {
	ldrp(c, &c.c, c.hl())
}

// 0x4F
func ldCA(c *CPU) {
	fetchCycle(c, func(c *CPU) { c.c = c.a })
}

// 0x50
func ldDB(c *CPU) {
	fetchCycle(c, func(c *CPU) { c.d = c.b })
}

// 0x51
func ldDC(c *CPU) {
	fetchCycle(c, func(c *CPU) { c.d = c.c })
}

// 0x52 NOP

// 0x53
func ldDE(c *CPU) {
	fetchCycle(c, func(c *CPU) { c.d = c.e })
}

// 0x54
func ldDH(c *CPU) {
	fetchCycle(c, func(c *CPU) { c.d = c.h })
}

// 0x55
func ldDL(c *CPU) {
	fetchCycle(c, func(c *CPU) { c.d = c.l })
}

// 0x56
func ldDHLp(c *CPU) {
	ldrp(c, &c.d, c.hl())
}

// 0x57
func ldDA(c *CPU) {
	fetchCycle(c, func(c *CPU) { c.d = c.a })
}

// 0x58
func ldEB(c *CPU) {
	fetchCycle(c, func(c *CPU) { c.e = c.b })
}

// 0x59
func ldEC(c *CPU) {
	fetchCycle(c, func(c *CPU) { c.e = c.c })
}

// 0x5A
func ldED(c *CPU) {
	fetchCycle(c, func(c *CPU) { c.e = c.d })
}

// 0x5B NOP

// 0x5C
func ldEH(c *CPU) {
	fetchCycle(c, func(c *CPU) { c.e = c.h })
}

// 0x5D
func ldEL(c *CPU) {
	fetchCycle(c, func(c *CPU) { c.e = c.l })
}

// 0x5E
func ldEHLp(c *CPU) {
	ldrp(c, &c.e, c.hl())
}

// 0x5F
func ldEA(c *CPU) {
	fetchCycle(c, func(c *CPU) { c.e = c.a })
}

// 0x60
func ldHB(c *CPU) {
	fetchCycle(c, func(c *CPU) { c.h = c.b })
}

// 0x61
func ldHC(c *CPU) {
	fetchCycle(c, func(c *CPU) { c.h = c.c })
}

// 0x62 NOP
func ldHD(c *CPU) {
	fetchCycle(c, func(c *CPU) { c.h = c.d })
}

// 0x63
func ldHE(c *CPU) {
	fetchCycle(c, func(c *CPU) { c.h = c.e })
}

// 0x64 NOP

// 0x65
func ldHL(c *CPU) {
	fetchCycle(c, func(c *CPU) { c.h = c.l })
}

// 0x66
func ldHHLp(c *CPU) {
	ldrp(c, &c.h, c.hl())
}

// 0x67
func ldHA(c *CPU) {
	fetchCycle(c, func(c *CPU) { c.h = c.a })
}

// 0x68
func ldLB(c *CPU) {
	fetchCycle(c, func(c *CPU) { c.l = c.b })
}

// 0x69
func ldLC(c *CPU) {
	fetchCycle(c, func(c *CPU) { c.l = c.c })
}

// 0x6A
func ldLD(c *CPU) {
	fetchCycle(c, func(c *CPU) { c.l = c.d })
}

// 0x6B
func ldLE(c *CPU) {
	fetchCycle(c, func(c *CPU) { c.l = c.e })
}

// 0x6C
func ldLH(c *CPU) {
	fetchCycle(c, func(c *CPU) { c.l = c.h })
}

// 0x6D NOP

// 0x6E
func ldLHLp(c *CPU) {
	ldrp(c, &c.l, c.hl())
}

// 0x6F
func ldLA(c *CPU) {
	fetchCycle(c, func(c *CPU) { c.l = c.a })
}

// 0x70
func ldHLpB(c *CPU) {
	ldpr(c, c.hl(), c.b)
}

// 0x71
func ldHLpC(c *CPU) {
	ldpr(c, c.hl(), c.c)
}

// 0x72
func ldHLpD(c *CPU) {
	ldpr(c, c.hl(), c.d)
}

// 0x73
func ldHLpE(c *CPU) {
	ldpr(c, c.hl(), c.e)
}

// 0x74
func ldHLpH(c *CPU) {
	ldpr(c, c.hl(), c.h)
}

// 0x75
func ldHLpL(c *CPU) {
	ldpr(c, c.hl(), c.l)
}

// 0x76
func halt(c *CPU) {

	if !c.interrupts.MasterEnabled() && c.interrupts.InterruptsPending() {
		c.ops.Push(func(c *CPU) {
			opcode := c.mmu.Read(c.pc)
			c.ir = instructions[opcode]
			// Do not increment PC - HALT BUG
		})
	} else {
		c.ops.Push(func(c *CPU) {
			c.state = halted
		})
	}
}

// 0x77
func ldHLpA(c *CPU) {
	ldpr(c, c.hl(), c.a)
}

// 0x78
func ldAB(c *CPU) {
	fetchCycle(c, func(c *CPU) { c.a = c.b })
}

// 0x79
func ldAC(c *CPU) {
	fetchCycle(c, func(c *CPU) { c.a = c.c })
}

// 0x7A
func ldAD(c *CPU) {
	fetchCycle(c, func(c *CPU) { c.a = c.d })
}

// 0x7B
func ldAE(c *CPU) {
	fetchCycle(c, func(c *CPU) { c.a = c.e })
}

// 0x7C
func ldAH(c *CPU) {
	fetchCycle(c, func(c *CPU) { c.a = c.h })
}

// 0x7D
func ldAL(c *CPU) {
	fetchCycle(c, func(c *CPU) { c.a = c.l })
}

// 0x7E
func ldAHLp(c *CPU) {
	ldrp(c, &c.a, c.hl())
}

// 0x7F NOP

// 0x80
func addB(c *CPU) {
	addAd8(c, &c.b, false)
}

// 0x81
func addC(c *CPU) {
	addAd8(c, &c.c, false)
}

// 0x82
func addD(c *CPU) {
	addAd8(c, &c.d, false)
}

// 0x83
func addE(c *CPU) {
	addAd8(c, &c.e, false)
}

// 0x84
func addH(c *CPU) {
	addAd8(c, &c.h, false)
}

// 0x85
func addL(c *CPU) {
	addAd8(c, &c.l, false)
}

// 0x86
func addHLp(c *CPU) {
	c.ops.Push(func(c *CPU) {
		c.z = c.mmu.Read(c.hl())
	})

	addAd8(c, &c.z, false)
}

// 0x87
func addA(c *CPU) {
	addAd8(c, &c.a, false)
}

// 0x88
func adcB(c *CPU) {
	addAd8(c, &c.b, c.f.isSet(carry))
}

// 0x89
func adcC(c *CPU) {
	addAd8(c, &c.c, c.f.isSet(carry))
}

// 0x8A
func adcD(c *CPU) {
	addAd8(c, &c.d, c.f.isSet(carry))
}

// 0x8B
func adcE(c *CPU) {
	addAd8(c, &c.e, c.f.isSet(carry))
}

// 0x8C
func adcH(c *CPU) {
	addAd8(c, &c.h, c.f.isSet(carry))
}

// 0x8D
func adcL(c *CPU) {
	addAd8(c, &c.l, c.f.isSet(carry))
}

// 0x8E
func adcHLp(c *CPU) {
	c.ops.Push(func(c *CPU) {
		c.z = c.mmu.Read(c.hl())
	})
	addAd8(c, &c.z, c.f.isSet(carry))
}

// 0x8F
func adcA(c *CPU) {
	addAd8(c, &c.a, c.f.isSet(carry))
}

// 0x90
func subB(c *CPU) {
	subAd8(c, &c.b, false)
}

// 0x91
func subC(c *CPU) {
	subAd8(c, &c.c, false)
}

// 0x92
func subD(c *CPU) {
	subAd8(c, &c.d, false)
}

// 0x93
func subE(c *CPU) {
	subAd8(c, &c.e, false)
}

// 0x94
func subH(c *CPU) {
	subAd8(c, &c.h, false)
}

// 0x95
func subL(c *CPU) {
	subAd8(c, &c.l, false)
}

// 0x96
func subHLp(c *CPU) {
	c.ops.Push(func(c *CPU) {
		c.z = c.mmu.Read(c.hl())
	})
	subAd8(c, &c.z, false)
}

// 0x97
func subA(c *CPU) {
	subAd8(c, &c.a, false)
}

// 0x98
func sbcB(c *CPU) {
	subAd8(c, &c.b, c.f.isSet(carry))
}

// 0x99
func sbcC(c *CPU) {
	subAd8(c, &c.c, c.f.isSet(carry))
}

// 0x9A
func sbcD(c *CPU) {
	subAd8(c, &c.d, c.f.isSet(carry))
}

// 0x9B
func sbcE(c *CPU) {
	subAd8(c, &c.e, c.f.isSet(carry))
}

// 0x9C
func sbcH(c *CPU) {
	subAd8(c, &c.h, c.f.isSet(carry))
}

// 0x9D
func sbcL(c *CPU) {
	subAd8(c, &c.l, c.f.isSet(carry))
}

// 0x9E
func sbcHLp(c *CPU) {
	c.ops.Push(func(c *CPU) {
		c.z = c.mmu.Read(c.hl())
	})
	subAd8(c, &c.z, c.f.isSet(carry))
}

// 0x9F
func sbcA(c *CPU) {
	subAd8(c, &c.a, c.f.isSet(carry))
}

// 0xA0
func andB(c *CPU) {
	andAd8(c, &c.b)
}

// 0xA1
func andC(c *CPU) {
	andAd8(c, &c.c)
}

// 0xA2
func andD(c *CPU) {
	andAd8(c, &c.d)
}

// 0xA3
func andE(c *CPU) {
	andAd8(c, &c.e)
}

// 0xA4
func andH(c *CPU) {
	andAd8(c, &c.h)
}

// 0xA5
func andL(c *CPU) {
	andAd8(c, &c.l)
}

// 0xA6
func andHLp(c *CPU) {
	c.ops.Push(func(c *CPU) {
		c.z = c.mmu.Read(c.hl())
	})
	andAd8(c, &c.z)
}

// 0xA7
func andA(c *CPU) {
	andAd8(c, &c.a)
}

// 0xA8
func xorB(c *CPU) {
	xorAd8(c, &c.b)
}

// 0xA9
func xorC(c *CPU) {
	xorAd8(c, &c.c)
}

// 0xAA
func xorD(c *CPU) {
	xorAd8(c, &c.d)
}

// 0xAB
func xorE(c *CPU) {
	xorAd8(c, &c.e)
}

// 0xAC
func xorH(c *CPU) {
	xorAd8(c, &c.h)
}

// 0xAD
func xorL(c *CPU) {
	xorAd8(c, &c.l)
}

// 0x9E
func xorHLp(c *CPU) {
	c.ops.Push(func(c *CPU) {
		c.z = c.mmu.Read(c.hl())
	})
	xorAd8(c, &c.z)
}

// 0xAF
func xorA(c *CPU) {
	xorAd8(c, &c.a)
}

// 0xB0
func orB(c *CPU) {
	orAd8(c, &c.b)
}

// 0xB1
func orC(c *CPU) {
	orAd8(c, &c.c)
}

// 0xB2
func orD(c *CPU) {
	orAd8(c, &c.d)
}

// 0xB3
func orE(c *CPU) {
	orAd8(c, &c.e)
}

// 0xB4
func orH(c *CPU) {
	orAd8(c, &c.h)
}

// 0xB5
func orL(c *CPU) {
	orAd8(c, &c.l)
}

// 0xA6
func orHLp(c *CPU) {
	c.ops.Push(func(c *CPU) {
		c.z = c.mmu.Read(c.hl())
	})
	orAd8(c, &c.z)
}

// 0xB7
func orA(c *CPU) {
	orAd8(c, &c.a)
}

// 0xB8
func cpB(c *CPU) {
	cpAd8(c, &c.b, false)
}

// 0xB9
func cpC(c *CPU) {
	cpAd8(c, &c.c, false)
}

// 0xBA
func cpD(c *CPU) {
	cpAd8(c, &c.d, false)
}

// 0xBB
func cpE(c *CPU) {
	cpAd8(c, &c.e, false)
}

// 0xBC
func cpH(c *CPU) {
	cpAd8(c, &c.h, false)
}

// 0xBD
func cpL(c *CPU) {
	cpAd8(c, &c.l, false)
}

// 0x9E
func cpHLp(c *CPU) {
	c.ops.Push(func(c *CPU) {
		c.z = c.mmu.Read(c.hl())
	})
	cpAd8(c, &c.z, false)
}

// 0xBF
func cpA(c *CPU) {
	cpAd8(c, &c.a, false)
}

// 0xC0
func retnz(c *CPU) {
	retcc(c, !c.f.isSet(zero))
}

// 0xC1
func popBC(c *CPU) {
	poprr(c, &c.b, &c.c)
}

// 0xC2
func jpnzd16(c *CPU) {
	jpccd16(c, !c.f.isSet(zero))
}

// 0xC3
func jpd16(c *CPU) {
	jpccd16(c, true)
}

// 0xC4
func callnzd16(c *CPU) {
	callccd16(c, !c.f.isSet(zero))
}

// 0xC5
func pushBC(c *CPU) {
	pushrr(c, c.b, c.c)
}

// 0xC6
func addd8(c *CPU) {
	readd8(c)
	addAd8(c, &c.z, false)
}

// 0xC7
func rst00(c *CPU) {
	pushrr(c, byte(c.pc>>8), byte(c.pc), func(c *CPU) { c.pc = 0x0000 })
}

// 0xC8
func retz(c *CPU) {
	retcc(c, c.f.isSet(zero))
}

// 0xC9
func ret(c *CPU) {
	doRet(c)
}

// 0xCA
func jpzd16(c *CPU) {
	jpccd16(c, c.f.isSet(zero))
}

// 0xCB
func cb(c *CPU) {
	c.ops.Push(func(c *CPU) {
		opCode := c.mmu.Read(c.pc)
		c.pcOfInstruction = c.pc
		c.ir = extendedInstructions[opCode]
		c.pc++
	})
}

// 0xCC
func callzd16(c *CPU) {
	callccd16(c, c.f.isSet(zero))
}

// 0xCD
func calld16(c *CPU) {
	callccd16(c, true)
}

// 0xCE
func adcd8(c *CPU) {
	readd8(c)
	addAd8(c, &c.z, c.f.isSet(carry))
}

// 0xCF
func rst08(c *CPU) {
	pushrr(c, byte(c.pc>>8), byte(c.pc), func(c *CPU) { c.pc = 0x0008 })
}

// 0xD0
func retnc(c *CPU) {
	retcc(c, !c.f.isSet(carry))
}

// 0xD1
func popDE(c *CPU) {
	poprr(c, &c.d, &c.e)
}

// 0xD2
func jpncd16(c *CPU) {
	jpccd16(c, !c.f.isSet(carry))
}

// 0xD3 UNDEFINED

// 0xD4
func callncd16(c *CPU) {
	callccd16(c, !c.f.isSet(carry))
}

// 0xD5
func pushDE(c *CPU) {
	pushrr(c, c.d, c.e)
}

// 0xD6
func subd8(c *CPU) {
	readd8(c)
	subAd8(c, &c.z, false)
}

// 0xD7
func rst10(c *CPU) {
	pushrr(c, byte(c.pc>>8), byte(c.pc), func(c *CPU) { c.pc = 0x0010 })
}

// 0xD8
func retc(c *CPU) {
	retcc(c, c.f.isSet(carry))
}

// 0xD9
func reti(c *CPU) {
	doRet(c, func(c *CPU) { c.interrupts.SetMasterEnable(true) })
}

// 0xDA
func jpcd16(c *CPU) {
	jpccd16(c, c.f.isSet(carry))
}

// 0xDB UNDEFINED

// 0xDC
func callcd16(c *CPU) {
	callccd16(c, c.f.isSet(carry))
}

// 0xDD UNDEFINED

// 0xDE
func sbcd8(c *CPU) {
	readd8(c)
	subAd8(c, &c.z, c.f.isSet(carry))
}

// 0xDF
func rst18(c *CPU) {
	pushrr(c, byte(c.pc>>8), byte(c.pc), func(c *CPU) { c.pc = 0x0018 })
}

// 0xE0
func ldff00d8A(c *CPU) {
	readd8(c)

	c.ops.Push(func(c *CPU) {
		c.mmu.Write(uint16(0xFF00)|uint16(c.z), c.a)
	})

	fetchCycle(c)
}

// 0xE1
func popHL(c *CPU) {
	poprr(c, &c.h, &c.l)
}

// 0xE2
func ldff00CA(c *CPU) {
	ldpr(c, uint16(0xFF00)|uint16(c.c), c.a)
}

// 0xE3 UNDEFINED

// 0xE4 UNDEFINED

// 0xE5
func pushHL(c *CPU) {
	pushrr(c, c.h, c.l)
}

// 0xE6
func andd8(c *CPU) {
	readd8(c)
	andAd8(c, &c.z)
}

// 0xE7
func rst20(c *CPU) {
	pushrr(c, byte(c.pc>>8), byte(c.pc), func(c *CPU) { c.pc = 0x0020 })
}

// 0xE8
func addSPr8(c *CPU) {
	readd8(c)

	// Lo byte addition
	// Done after this post https://www.reddit.com/r/EmuDev/comments/y51i1c/game_boy_dealing_with_carry_flags_when_handling/
	c.ops.Push(func(c *CPU) {
		c.f.unsetFlag(zero)
		c.f.unsetFlag(negative)

		lsb := byte(c.sp)
		readValue := c.z

		// saving read value also in w to have it on access in next step
		c.w = c.z
		result := lsb + readValue

		if util.BitIsSet8(lsb&0xF+readValue&0xF, 4) {
			c.f.setFlag(halfCarry)
		} else {
			c.f.unsetFlag(halfCarry)
		}

		if result < lsb {
			c.f.setFlag(carry)
		} else {
			c.f.unsetFlag(carry)
		}

		c.z = result
	})

	// Hi byte adjustment
	c.ops.Push(func(c *CPU) {
		msb := byte(c.sp >> 8)
		originalByte := c.w // originally read number still in temporary register w (copied in last step)

		var adj byte
		if util.BitIsSet8(originalByte, 7) {
			adj = 0xFF
		}

		var car byte
		if c.f.isSet(carry) {
			car = 0x01
		}

		c.w = msb + adj + car
	})

	fetchCycle(c, func(c *CPU) {
		c.sp = c.wz()
	})
}

// 0xE9
func jpHL(c *CPU) {
	fetchCycle(c, func(c *CPU) { c.pc = c.hl() })
}

// 0xEA
func ldd16pA(c *CPU) {
	readd16(c)

	c.ops.Push(func(c *CPU) {
		c.mmu.Write(c.wz(), c.a)
	})

	fetchCycle(c)
}

// 0xEB UNDEFINED

// 0xEC UNDEFINED

// 0xED UNDEFINED

// 0xEE
func xord8(c *CPU) {
	readd8(c)
	xorAd8(c, &c.z)
}

// 0xEF
func rst28(c *CPU) {
	pushrr(c, byte(c.pc>>8), byte(c.pc), func(c *CPU) { c.pc = 0x0028 })
}

// 0xF0
func ldAff00d8(c *CPU) {
	readd8(c)

	c.ops.Push(func(c *CPU) {
		c.z = c.mmu.Read(uint16(0xFF00) | uint16(c.z))
	})

	fetchCycle(c, func(c *CPU) {
		c.a = c.z
	})
}

// 0xF1
func popAF(c *CPU) {
	c.ops.Push(func(c *CPU) {
		c.z = c.mmu.Read(c.sp)
		c.sp++
	})

	c.ops.Push(func(c *CPU) {
		c.w = c.mmu.Read(c.sp)
		c.sp++
	})

	fetchCycle(c, func(c *CPU) {
		c.a = c.w
		c.f = flags(c.z & 0xF0)
	})
}

// 0xF2
func ldAff00C(c *CPU) {
	ldrp(c, &c.a, uint16(0xFF00)|uint16(c.c))
}

// 0xF3
func di(c *CPU) {
	fetchCycle(c, func(c *CPU) {
		c.interrupts.SetMasterEnable(false)
	})
}

// 0xF4 UNDEFINED

// 0xF5
func pushAF(c *CPU) {
	pushrr(c, c.a, byte(c.f)&0xF0)
}

// 0xF6
func ord8(c *CPU) {
	readd8(c)
	orAd8(c, &c.z)
}

// 0xF7
func rst30(c *CPU) {
	pushrr(c, byte(c.pc>>8), byte(c.pc), func(c *CPU) { c.pc = 0x0030 })
}

// 0xF8
func ldHLSPr8(c *CPU) {

	readd8(c)

	// Lo byte addition
	// Done after this post https://www.reddit.com/r/EmuDev/comments/y51i1c/game_boy_dealing_with_carry_flags_when_handling/
	c.ops.Push(func(c *CPU) {

		c.f.unsetFlag(zero)
		c.f.unsetFlag(negative)

		lsb := byte(c.sp)
		result := lsb + c.z

		if util.BitIsSet8(lsb&0xF+c.z&0xF, 4) {
			c.f.setFlag(halfCarry)
		} else {
			c.f.unsetFlag(halfCarry)
		}

		if result < lsb {
			c.f.setFlag(carry)
		} else {
			c.f.unsetFlag(carry)
		}

		c.l = result
	})

	// Hi byte adjustment
	fetchCycle(c, func(c *CPU) {
		msb := byte(c.sp >> 8)

		var adj byte
		if util.BitIsSet8(c.z, 7) {
			adj = 0xFF
		}

		var car byte
		if c.f.isSet(carry) {
			car = 0x01
		}

		c.h = msb + adj + car
	})
}

// 0xF9
func ldSPHL(c *CPU) {
	c.ops.Push(func(c *CPU) {
		c.sp = c.hl()
	})

	fetchCycle(c)
}

// 0xFA
func ldAd16p(c *CPU) {
	readd16(c)

	c.ops.Push(func(c *CPU) {
		c.z = c.mmu.Read(c.wz())
	})

	fetchCycle(c, func(c *CPU) { c.a = c.z })
}

// 0xFB
func ei(c *CPU) {
	fetchCycle(c,
		func(_ *CPU) { /* nothing to do before instr fetch*/ },
		func(c *CPU) { c.interrupts.SetMasterEnable(true) },
	)
}

// 0xFC UNDEFINED

// 0xFD UNDEFINED

// 0xFE
func cpd8(c *CPU) {
	readd8(c)
	cpAd8(c, &c.z, false)
}

// 0xFF
func rst38(c *CPU) {
	pushrr(c, byte(c.pc>>8), byte(c.pc), func(c *CPU) { c.pc = 0x0038 })
}

// Generic instructions used to regroup common code. The following placeholders
// are used in names and descriptions:
//
// r 	single (8-bit) register (A, F, B, C, D, E, H or L)
// rr	double (16-bit) register (AF, BC, DE or HL)
// d8	8-bit (unsigned) parameter (1 byte) after opcode
// d16	16-bit (unsigned) parameter (2 bytes, little-endian) after opcode
// r8	8-bit (signed) parameter (1 byte) after opcode
// p    Pointer (reading value from register and using it as memory address)

// ldrrd16 loads 16bit number into 16bit register
// Cycles: 12
func ldrrd16(c *CPU, high, low *byte) {
	readd16(c)
	fetchCycle(c, func(c *CPU) {
		*low = c.z
		*high = c.w
	})
}

// ldrd8 loads 8bit number into 8bit register
// Cycles: 8
func ldrd8(c *CPU, reg *byte) {
	readd8(c)
	fetchCycle(c, func(c *CPU) {
		*reg = c.z
	})
}

// ldrp loads contents of the given memory address to given 8bit register
// Cycles: 8
func ldrp(c *CPU, reg *byte, address uint16) {
	c.ops.Push(func(c *CPU) {
		c.z = c.mmu.Read(address)
	})

	fetchCycle(c, func(c *CPU) {
		*reg = c.z
	})
}

// ldpr loads the value of 8bit register to given memory address.
// Cycles: 8
func ldpr(c *CPU, address uint16, reg byte) {
	c.ops.Push(func(c *CPU) {
		c.mmu.Write(address, reg)
	})

	fetchCycle(c)
}

// incr increments given 8bit register by one and setting all according flags.
// Cycles: 4
func incr(c *CPU, reg *byte) {
	fetchCycle(c, func(c *CPU) {
		c.f.unsetFlag(negative)

		// overflow from bit 3 to 4
		if *reg&0x0F == 0x0F {
			c.f.setFlag(halfCarry)
		} else {
			c.f.unsetFlag(halfCarry)
		}

		*reg++

		if *reg == 0x00 {
			c.f.setFlag(zero)
		} else {
			c.f.unsetFlag(zero)
		}
	})
}

// decr decrements given 8bit register by one and setting all according flags.
// Cycles: 4
func decr(c *CPU, reg *byte) {
	fetchCycle(c, func(c *CPU) {
		c.f.setFlag(negative)

		// decrement will wrap around lower nibble
		if *reg&0xF == 0x00 {
			c.f.setFlag(halfCarry)
		} else {
			c.f.unsetFlag(halfCarry)
		}

		*reg--

		if *reg == 0x00 {
			c.f.setFlag(zero)
		} else {
			c.f.unsetFlag(zero)
		}
	})
}

// addHLrr adds two 16 bit registers together. Only the result of the high byte addition is relevant for flags.
// Cycles: 8
func addHLrr(c *CPU, high, low byte) {

	c.ops.Push(func(c *CPU) {
		c.f.unsetFlag(negative)

		oldValue := c.l
		c.l += low

		if c.l < oldValue {
			c.f.setFlag(carry)
		} else {
			c.f.unsetFlag(carry)
		}
	})

	fetchCycle(c, func(c *CPU) {
		oldValue := c.h

		car := byte(0)
		if c.f.isSet(carry) {
			car = 1
		}

		if util.BitIsSet8(c.h&0xF+high&0xF+car, 4) {
			c.f.setFlag(halfCarry)
		} else {
			c.f.unsetFlag(halfCarry)
		}

		c.h += high + car

		if c.h < oldValue || (c.f.isSet(carry) && c.h == oldValue) {
			c.f.setFlag(carry)
		} else {
			c.f.unsetFlag(carry)
		}
	})
}

// addAd8 adds value to A and writes the value back.
// Cycles: 4
func addAd8(c *CPU, val *byte, carryIn bool) {
	fetchCycle(c, func(c *CPU) {
		var carryBit byte
		if carryIn {
			carryBit = 0x1
		}

		c.f.unsetFlag(negative)

		if util.BitIsSet8(c.a&0xF+*val&0xF+carryBit, 4) {
			c.f.setFlag(halfCarry)
		} else {
			c.f.unsetFlag(halfCarry)
		}

		oldValue := c.a
		c.a += *val + carryBit

		if c.a < oldValue || (carryIn && c.a == oldValue) {
			c.f.setFlag(carry)
		} else {
			c.f.unsetFlag(carry)
		}

		if c.a == 0x00 {
			c.f.setFlag(zero)
		} else {
			c.f.unsetFlag(zero)
		}
	})
}

// subAd8 subtracts value from given 8bit register and writes the value back.
// Cycles: 4
func subAd8(c *CPU, val *byte, carryIn bool) {
	doSubCpAd8(c, val, carryIn, true)
}

// cpAd8 compares value with given 8bit register. Sets the flags but doesn't write the value
// Cycles: 4
func cpAd8(c *CPU, val *byte, carryIn bool) {
	doSubCpAd8(c, val, carryIn, false)
}

func doSubCpAd8(c *CPU, val *byte, carryIn bool, writeBack bool) {
	fetchCycle(c, func(c *CPU) {
		var carryBit byte
		if carryIn {
			carryBit = 0x1
		}

		c.f.setFlag(negative)

		if util.BitIsSet8(c.a&0xF-*val&0xF-carryBit, 4) {
			c.f.setFlag(halfCarry)
		} else {
			c.f.unsetFlag(halfCarry)
		}

		result := c.a - *val - carryBit

		if result > c.a || (carryIn && result == c.a) {
			c.f.setFlag(carry)
		} else {
			c.f.unsetFlag(carry)
		}

		if result == 0x00 {
			c.f.setFlag(zero)
		} else {
			c.f.unsetFlag(zero)
		}

		if writeBack {
			c.a = result
		}
	})
}

func andAd8(c *CPU, val *byte) {
	fetchCycle(c, func(c *CPU) {
		c.a &= *val
		if c.a == 0x00 {
			c.f.setFlag(zero)
		} else {
			c.f.unsetFlag(zero)
		}
		c.f.unsetFlag(negative)
		c.f.setFlag(halfCarry)
		c.f.unsetFlag(carry)
	})
}

func xorAd8(c *CPU, val *byte) {
	fetchCycle(c, func(c *CPU) {
		c.a ^= *val
		if c.a == 0x00 {
			c.f.setFlag(zero)
		} else {
			c.f.unsetFlag(zero)
		}
		c.f.unsetFlag(negative)
		c.f.unsetFlag(halfCarry)
		c.f.unsetFlag(carry)
	})
}

func orAd8(c *CPU, val *byte) {
	fetchCycle(c, func(c *CPU) {
		c.a |= *val
		if c.a == 0x00 {
			c.f.setFlag(zero)
		} else {
			c.f.unsetFlag(zero)
		}
		c.f.unsetFlag(negative)
		c.f.unsetFlag(halfCarry)
		c.f.unsetFlag(carry)
	})
}

func jrcc(c *CPU, condition bool) {
	readd8(c)

	if !condition {
		fetchCycle(c)
		return
	}

	c.ops.Push(func(c *CPU) {
		// lsb
		sign := util.BitIsSet8(c.z, 7)
		result := c.z + byte(c.pc)
		carryOver := result < c.z
		c.z = result

		var adj byte
		if carryOver && !sign {
			adj = 0x01
		} else if !carryOver && sign {
			adj = 0xFF
		}

		// msb
		c.w = byte(c.pc>>8) + adj
	})

	fetchCycle(c, func(c *CPU) { c.pc = c.wz() })
}

func jpccd16(c *CPU, condition bool) {
	readd16(c)

	if condition {
		// Set the PC to the read number
		c.ops.Push(func(c *CPU) {
			c.pc = c.wz()
		})
	}

	fetchCycle(c)
}

func pushrr(c *CPU, high, low byte, doAfterPush ...func(*CPU)) {
	c.ops.Push(func(c *CPU) {
		c.sp--
	})

	// Push high register to stack.
	c.ops.Push(func(c *CPU) {
		c.mmu.Write(c.sp, high)
		c.sp--
	})

	// Push low register to stack.
	c.ops.Push(func(c *CPU) {
		c.mmu.Write(c.sp, low)
		if len(doAfterPush) > 0 {
			doAfterPush[0](c)
		}
	})

	fetchCycle(c)
}

// poprr stores the 16-bit value at the memory address in SP in the given
// 16-bit register.
// Cycles: 12
func poprr(c *CPU, high, low *uint8) {
	c.ops.Push(func(c *CPU) {
		c.z = c.mmu.Read(c.sp)
		c.sp++
	})

	c.ops.Push(func(c *CPU) {
		c.w = c.mmu.Read(c.sp)
		c.sp++
	})

	fetchCycle(c, func(c *CPU) {
		*low = c.z
		*high = c.w
	})
}

// doRet returns from subroutine
// Cycles: 16
func doRet(c *CPU, doAfterReturn ...func(*CPU)) {
	c.ops.Push(func(c *CPU) {
		c.z = c.mmu.Read(c.sp)
		c.sp++
	})

	c.ops.Push(func(c *CPU) {
		c.w = c.mmu.Read(c.sp)
		c.sp++
	})

	c.ops.Push(func(c *CPU) {
		c.pc = c.wz()
		if len(doAfterReturn) > 0 {
			doAfterReturn[0](c)
		}
	})

	fetchCycle(c)
}

// retcc returns from subroutine based on given condition
// Cycles: 8 (condition = false) or 20 (condition = true)
func retcc(c *CPU, condition bool) {
	c.ops.Push(func(c *CPU) {
		// Does nothing
	})

	if !condition {
		fetchCycle(c)
		return
	}

	doRet(c)
}

// callccd16 calls subroutine if condition is met
// Cycles: 12 (condition = false) or 24 (condition = true)
func callccd16(c *CPU, condition bool) {
	readd16(c)

	if condition {
		c.ops.Push(func(c *CPU) {
			c.sp--
		})

		// Push high byte to stack.
		c.ops.Push(func(c *CPU) {
			c.mmu.Write(c.sp, byte(c.pc>>8))
			c.sp--
		})

		// Push low byte to stack and update PC in the same operation.
		c.ops.Push(func(c *CPU) {
			c.mmu.Write(c.sp, byte(c.pc))
			c.pc = c.wz()
		})
	}

	fetchCycle(c)
}

// readd8 reads a 16bit number into both temporary registers. After this operation Z contains value.
// Cycles: 4
func readd8(c *CPU) {
	c.ops.Push(func(c *CPU) {
		c.z = c.mmu.Read(c.pc)
		c.pc++
	})
}

// readd16 reads a 16bit number into both temporary registers. After this operation Z contains the lsb and W the msb.
// Cycles: 8
func readd16(c *CPU) {
	// read lsb
	readd8(c)

	// read msb
	c.ops.Push(func(c *CPU) {
		c.w = c.mmu.Read(c.pc)
		c.pc++
	})
}

func fetchCycle(c *CPU, doInFetchCycle ...func(*CPU)) {
	beforeInstrFetch := func(c *CPU) { /* Do nothing */ }
	if len(doInFetchCycle) >= 1 {
		beforeInstrFetch = doInFetchCycle[0]
	}

	afterInstrFetch := func(c *CPU) { /* Do nothing */ }
	if len(doInFetchCycle) >= 2 {
		afterInstrFetch = doInFetchCycle[1]
	}

	c.ops.Push(func(c *CPU) {
		beforeInstrFetch(c)
		c.pcOfInstruction = c.pc
		opCode := c.mmu.Read(c.pc)
		c.ir = instructions[opCode]
		c.pc++

		if c.interrupts.MustHandleInterrupt() {
			c.interrupts.HandleInterrupt(func(t interrupts.InterruptType) {
				enqueueInterruptRoutine(c, t, afterInstrFetch)
			})
		} else {
			afterInstrFetch(c)
		}
	})
}

func enqueueInterruptRoutine(c *CPU, t interrupts.InterruptType, afterInstrFetch func(*CPU)) {
	c.ops.Push(func(c *CPU) {
		c.pc--
	})

	c.ops.Push(func(c *CPU) {
		c.sp--
	})

	// Write high Byte of PC
	c.ops.Push(func(c *CPU) {
		c.mmu.Write(c.sp, byte(c.pc>>8))
		c.sp--
	})

	// Write low Byte of PC
	c.ops.Push(func(c *CPU) {
		c.mmu.Write(c.sp, byte(c.pc))

		switch t {
		case interrupts.VBlank:
			c.pc = 0x40
		case interrupts.LcdStat:
			c.pc = 0x48
		case interrupts.Timer:
			c.pc = 0x50
		case interrupts.Serial:
			c.pc = 0x58
		case interrupts.Joypad:
			c.pc = 0x60
		}
	})

	c.ops.Push(func(c *CPU) {
		c.pcOfInstruction = c.pc
		opCode := c.mmu.Read(c.pc)
		c.ir = instructions[opCode]
		c.pc++
		afterInstrFetch(c)
	})
}
