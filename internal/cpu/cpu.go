// Source: https://gbdev.io/pandocs/CPU_Registers_and_Flags.html
package cpu

import (
	"fmt"
	"gameboy-emulator/internal/interrupts"
	"gameboy-emulator/internal/memory"
	log "go.uber.org/zap"
	"strings"
)

type (
	CPU struct {
		a  SubRegister
		f  flags
		bc Register
		de Register
		hl Register

		// 16 bit registers
		pc uint16 // Program counter
		sp uint16 // Stack pointer

		memory     *memory.Memory
		interrupts *interrupts.Interrupts

		stopped    bool
		stepCycles int
	}

	instruction struct {
		disassembly   string
		operandLength byte               // length of the operand in number of bytes (0 = no operand, 1 = 8 bit, 2 = 16 bit)
		execute       func(*CPU, uint16) // uint16 so that a 2 byte long operator may be accepted
		ticks         byte               // duration in number of t cycles
	}
)

func (i instruction) toString(operand uint16) string {
	result := fmt.Sprintf(i.disassembly, operand)

	if i := strings.Index(result, "%!"); i != -1 {
		result = result[:i]
	}

	return result
}

var instructions = [256]instruction{
	{"NOP", 0, nop, 4},                         // 0x00
	{"LD BC, 0x%04X", 2, ldBCnn, 12},           // 0x01
	{"LD (BC), A", 0, ldDBCA, 8},               // 0x02
	{"INC BC", 0, incBC, 8},                    // 0x03
	{"INC B", 0, incB, 4},                      // 0x04
	{"DEC B", 0, decB, 4},                      // 0x05
	{"LD B, 0x%02X", 1, ldBn, 8},               // 0x06
	{"RLCA", 0, rlca, 4},                       // 0x07
	{"LD (0x%04X), SP", 0, ldDnnSP, 20},        // 0x08
	{"ADD HL, BC", 0, addHLBC, 8},              // 0x09
	{"LD A, (BC)", 0, ldADBC, 8},               // 0x0A
	{"DEC BC", 0, decBC, 8},                    // 0x0B
	{"INC C", 0, incC, 4},                      // 0x0C
	{"DEC C", 0, decC, 4},                      // 0x0D
	{"LD C, 0x%02X", 1, ldCn, 8},               // 0x0E
	{"RRCA", 0, rrca, 4},                       // 0x0F
	{"STOP", 0, stop, 4},                       // 0x10
	{"LD DE, 0x%04X", 2, ldDEnn, 12},           // 0x11
	{"LD (DE), A", 0, ldDDEA, 8},               // 0x12
	{"INC DE", 0, incDE, 8},                    // 0x13
	{"INC D", 0, incD, 4},                      // 0x14
	{"DEC D", 0, decD, 4},                      // 0x15
	{"LD D, 0x%02X", 1, ldDn, 8},               // 0x16
	{"RLA", 0, rla, 4},                         // 0x17
	{"JR 0x%02X", 1, jrn, 8},                   // 0x18
	{"ADD HL, DE", 0, addHLDE, 8},              // 0x19
	{"LD A, (DE)", 0, ldADDE, 8},               // 0x1A
	{"DEC DE", 0, decDE, 8},                    // 0x1B
	{"INC E", 0, incE, 4},                      // 0x1C
	{"DEC E", 0, decE, 4},                      // 0x1D
	{"LD E, 0x%02X", 1, ldEn, 8},               // 0x1E
	{"RRA", 0, rra, 4},                         // 0x1F
	{"JR NZ, 0x%02X", 1, jrnzn, 8},             // 0x20
	{"LD HL, 0x%04X", 2, ldHLnn, 12},           // 0x21
	{"LD (HL+), A", 0, ldiHLA, 8},              // 0x22
	{"INC HL", 0, incHL, 8},                    // 0x23
	{"INC H", 0, incH, 4},                      // 0x24
	{"DEC H", 0, decH, 4},                      // 0x25
	{"LD H, 0x%02X", 1, ldHn, 8},               // 0x26
	{"DAA", 0, daa, 4},                         // 0x27
	{"JR Z, 0x%02X", 1, jrzn, 8},               // 0x28
	{"ADD HL, HL", 0, addHLHL, 8},              // 0x29
	{"LD A, (HL+)", 0, ldiAHL, 8},              // 0x2A
	{"DEC HL", 0, decHL, 8},                    // 0x2B
	{"INC L", 0, incL, 4},                      // 0x2C
	{"DEC L", 0, decL, 4},                      // 0x2D
	{"LD L, 0x%02X", 1, ldLn, 8},               // 0x2E
	{"CPL", 0, cpl, 4},                         // 0x2F
	{"JR NC, 0x%02X", 1, jrncn, 8},             // 0x30
	{"LD SP, 0x%04X", 2, ldSPnn, 12},           // 0x31
	{"LD (HL-), A", 0, lddHLA, 8},              // 0x32
	{"INC SP", 0, incSP, 8},                    // 0x33
	{"INC (HL)", 0, incDHL, 12},                // 0x34
	{"DEC (HL)", 0, decDHL, 12},                // 0x35
	{"LD (HL), 0x%02X", 1, ldDHLn, 12},         // 0x36
	{"SCF", 0, scf, 4},                         // 0x37
	{"JR C, 0x%02X", 1, jrcn, 8},               // 0x38
	{"ADD HL, SP", 0, addHLSP, 8},              // 0x39
	{"LD A, (HL-)", 0, lddAHL, 8},              // 0x3A
	{"DEC SP", 0, decSP, 8},                    // 0x3B
	{"INC A", 0, incA, 4},                      // 0x3C
	{"DEC A", 0, decA, 4},                      // 0x3D
	{"LD A, 0x%02X", 1, ldAn, 8},               // 0x3E
	{"CCF", 0, ccf, 4},                         // 0x3F
	{"LD B, B", 0, nop, 4},                     // 0x40
	{"LD B, C", 0, ldBC, 4},                    // 0x41
	{"LD B, D", 0, ldBD, 4},                    // 0x42
	{"LD B, E", 0, ldBE, 4},                    // 0x43
	{"LD B, H", 0, ldBH, 4},                    // 0x44
	{"LD B, L", 0, ldBL, 4},                    // 0x45
	{"LD B, (HL)", 0, ldBDHL, 8},               // 0x46
	{"LD B, A", 0, ldBA, 4},                    // 0x47
	{"LD C, B", 0, ldCB, 4},                    // 0x48
	{"LD C, C", 0, nop, 4},                     // 0x49
	{"LD C, D", 0, ldCD, 4},                    // 0x4A
	{"LD C, E", 0, ldCE, 4},                    // 0x4B
	{"LD C, H", 0, ldCH, 4},                    // 0x4C
	{"LD C, L", 0, ldCL, 4},                    // 0x4D
	{"LD C, (HL)", 0, ldCDHL, 8},               // 0x4E
	{"LD C, A", 0, ldCA, 4},                    // 0x4F
	{"LD D, B", 0, ldDB, 4},                    // 0x50
	{"LD D, C", 0, ldDC, 4},                    // 0x51
	{"LD D, D", 0, nop, 4},                     // 0x52
	{"LD D, E", 0, ldDE, 4},                    // 0x53
	{"LD D, H", 0, ldDH, 4},                    // 0x54
	{"LD D, L", 0, ldDL, 4},                    // 0x55
	{"LD D, (HL)", 0, ldDDHL, 8},               // 0x56
	{"LD D, A", 0, ldDA, 4},                    // 0x57
	{"LD E, B", 0, ldEB, 4},                    // 0x58
	{"LD E, C", 0, ldEC, 4},                    // 0x59
	{"LD E, D", 0, ldED, 4},                    // 0x5A
	{"LD E, E", 0, nop, 4},                     // 0x5B
	{"LD E, H", 0, ldEH, 4},                    // 0x5C
	{"LD E, L", 0, ldEL, 4},                    // 0x5D
	{"LD E, (HL)", 0, ldEDHL, 8},               // 0x5E
	{"LD E, A", 0, ldEA, 4},                    // 0x5F
	{"LD H, B", 0, ldHB, 4},                    // 0x60
	{"LD H, C", 0, ldHC, 4},                    // 0x61
	{"LD H, D", 0, ldHD, 4},                    // 0x62
	{"LD H, E", 0, ldHE, 4},                    // 0x63
	{"LD H, H", 0, nop, 4},                     // 0x64
	{"LD H, L", 0, ldHL, 4},                    // 0x65
	{"LD H, (HL)", 0, ldHDHL, 8},               // 0x66
	{"LD H, A", 0, ldHA, 4},                    // 0x67
	{"LD L, B", 0, ldLB, 4},                    // 0x68
	{"LD L, C", 0, ldLC, 4},                    // 0x69
	{"LD L, D", 0, ldLD, 4},                    // 0x6A
	{"LD L, E", 0, ldLE, 4},                    // 0x6B
	{"LD L, H", 0, ldLH, 4},                    // 0x6C
	{"LD L, L", 0, nop, 4},                     // 0x6D
	{"LD L, (HL)", 0, ldLDHL, 8},               // 0x6E
	{"LD L, A", 0, ldLA, 4},                    // 0x6F
	{"LD (HL), B", 0, ldDHLB, 4},               // 0x70
	{"LD (HL), C", 0, ldDHLC, 4},               // 0x71
	{"LD (HL), D", 0, ldDHLD, 4},               // 0x72
	{"LD (HL), E", 0, ldDHLE, 4},               // 0x73
	{"LD (HL), H", 0, ldDHLH, 4},               // 0x74
	{"LD (HL), L", 0, ldDHLL, 4},               // 0x75
	{"HALT", 0, halt, 4},                       // 0x76
	{"LD (HL), A", 0, ldDHLA, 8},               // 0x77
	{"LD A, B", 0, ldAB, 4},                    // 0x78
	{"LD A, C", 0, ldAC, 4},                    // 0x79
	{"LD A, D", 0, ldAD, 4},                    // 0x7A
	{"LD A, E", 0, ldAE, 4},                    // 0x7B
	{"LD A, H", 0, ldAH, 4},                    // 0x7C
	{"LD A, L", 0, ldAL, 4},                    // 0x7D
	{"LD A, (HL)", 0, ldADHL, 8},               // 0x7E
	{"LD A, A", 0, nop, 4},                     // 0x7F
	{"ADD B", 0, addB, 4},                      // 0x80
	{"ADD C", 0, addC, 4},                      // 0x81
	{"ADD D", 0, addD, 4},                      // 0x82
	{"ADD E", 0, addE, 4},                      // 0x83
	{"ADD H", 0, addH, 4},                      // 0x84
	{"ADD L", 0, addL, 4},                      // 0x85
	{"ADD (HL)", 0, addDHL, 8},                 // 0x86
	{"ADD A", 0, addA, 4},                      // 0x87
	{"ADC B", 0, adcB, 4},                      // 0x88
	{"ADC C", 0, adcC, 4},                      // 0x89
	{"ADC D", 0, adcD, 4},                      // 0x8A
	{"ADC E", 0, adcE, 4},                      // 0x8B
	{"ADC H", 0, adcH, 4},                      // 0x8C
	{"ADC L", 0, adcL, 4},                      // 0x8D
	{"ADC (HL)", 0, adcDHL, 8},                 // 0x8E
	{"ADC A", 0, adcA, 4},                      // 0x8F
	{"SUB B", 0, subB, 4},                      // 0x90
	{"SUB C", 0, subC, 4},                      // 0x91
	{"SUB D", 0, subD, 4},                      // 0x92
	{"SUB E", 0, subE, 4},                      // 0x93
	{"SUB H", 0, subH, 4},                      // 0x94
	{"SUB L", 0, subL, 4},                      // 0x95
	{"SUB (HL)", 0, subDHL, 8},                 // 0x96
	{"SUB A", 0, subA, 4},                      // 0x97
	{"SBC B", 0, sbcB, 4},                      // 0x98
	{"SBC C", 0, sbcC, 4},                      // 0x99
	{"SBC D", 0, sbcD, 4},                      // 0x9A
	{"SBC E", 0, sbcE, 4},                      // 0x9B
	{"SBC H", 0, sbcH, 4},                      // 0x9C
	{"SBC L", 0, sbcL, 4},                      // 0x9D
	{"SBC (HL)", 0, sbcDHL, 8},                 // 0x9E
	{"SBC A", 0, sbcA, 4},                      // 0x9F
	{"AND B", 0, andB, 4},                      // 0xA0
	{"AND C", 0, andC, 4},                      // 0xA1
	{"AND D", 0, andD, 4},                      // 0xA2
	{"AND E", 0, andE, 4},                      // 0xA3
	{"AND H", 0, andH, 4},                      // 0xA4
	{"AND L", 0, andL, 4},                      // 0xA5
	{"AND (HL)", 0, andDHL, 8},                 // 0xA6
	{"AND A", 0, andA, 4},                      // 0xA7
	{"XOR B", 0, xorB, 4},                      // 0xA8
	{"XOR C", 0, xorC, 4},                      // 0xA9
	{"XOR D", 0, xorD, 4},                      // 0xAA
	{"XOR E", 0, xorE, 4},                      // 0xAB
	{"XOR H", 0, xorH, 4},                      // 0xAC
	{"XOR L", 0, xorL, 4},                      // 0xAD
	{"XOR (HL)", 0, xorDHL, 8},                 // 0xAE
	{"XOR A", 0, xorA, 4},                      // 0xAF
	{"OR B", 0, orB, 4},                        // 0xB0
	{"OR C", 0, orC, 4},                        // 0xB1
	{"OR D", 0, orD, 4},                        // 0xB2
	{"OR E", 0, orE, 4},                        // 0xB3
	{"OR H", 0, orH, 4},                        // 0xB4
	{"OR L", 0, orL, 4},                        // 0xB5
	{"OR (HL)", 0, orDHL, 8},                   // 0xB6
	{"OR A", 0, orA, 4},                        // 0xB7
	{"CP B", 0, cpB, 4},                        // 0xB8
	{"CP C", 0, cpC, 4},                        // 0xB9
	{"CP D", 0, cpD, 4},                        // 0xBA
	{"CP E", 0, cpE, 4},                        // 0xBB
	{"CP H", 0, cpH, 4},                        // 0xBC
	{"CP L", 0, cpL, 4},                        // 0xBD
	{"CP (HL)", 0, cpDHL, 8},                   // 0xBE
	{"CP A", 0, cpA, 4},                        // 0xBF
	{"RET NZ", 0, retnz, 8},                    // 0xC0
	{"POP BC", 0, popBC, 12},                   // 0xC1
	{"JP NZ, 0x%04X", 2, jpnznn, 12},           // 0xC2
	{"JP 0x%04X", 2, jpnn, 12},                 // 0xC3
	{"CALL NZ, 0x%04X", 2, callnznn, 12},       // 0xC4
	{"PUSH BC", 0, pushBC, 16},                 // 0xC5
	{"ADD 0x%02X", 1, addn, 8},                 // 0xC6
	{"RST 0x00", 0, rst00, 16},                 // 0xC7
	{"RET Z", 0, retz, 8},                      // 0xC8
	{"RET", 0, ret, 4},                         // 0xC9
	{"JP Z, 0x%04X", 2, jpznn, 12},             // 0xCA
	{"CB ", 1, cbn, 4},                         // 0xCB
	{"CALL Z, 0x%04X", 2, callznn, 12},         // 0xCC
	{"CALL 0x%04X", 2, callnn, 12},             // 0xCD
	{"ADC 0x%02X", 1, adcn, 8},                 // 0xCE
	{"RST 0x08", 0, rst08, 16},                 // 0xCF
	{"RET NC", 0, retnc, 8},                    // 0xD0
	{"POP DE", 0, popDE, 12},                   // 0xD1
	{"JP NC, 0x%04X", 2, jpncnn, 12},           // 0xD2
	{"UNDEFINED", 0, undefined, 0},             // 0xD3
	{"CALL NC, 0x%04X", 2, callncnn, 12},       // 0xD4
	{"PUSH DE", 0, pushDE, 16},                 // 0xD5
	{"SUB 0x%02X", 1, subn, 8},                 // 0xD6
	{"RST 0x10", 0, rst10, 16},                 // 0xD7
	{"RET C", 0, retc, 8},                      // 0xD8
	{"RETI", 0, reti, 16},                      // 0xD9
	{"JP C, 0x%04X", 2, jpcnn, 12},             // 0xDA
	{"UNDEFINED", 0, undefined, 0},             // 0xDB
	{"CALL C, 0x%04X", 2, callcnn, 12},         // 0xDC
	{"UNDEFINED", 0, undefined, 0},             // 0xDD
	{"SBC 0x%02X", 1, sbcn, 8},                 // 0xDE
	{"RST 0x18", 0, rst18, 16},                 // 0xDF
	{"LD (0xFF00+0x%02X), A", 1, ldff00nA, 12}, // 0xE0
	{"POP HL", 0, popHL, 12},                   // 0xE1
	{"LD (0xFF00+C), A", 0, ldff00CA, 8},       // 0xE2
	{"UNDEFINED", 0, undefined, 0},             // 0xE3
	{"UNDEFINED", 0, undefined, 0},             // 0xE4
	{"PUSH HL", 0, pushHL, 16},                 // 0xE5
	{"AND 0x%02X", 1, andn, 8},                 // 0xE6
	{"RST 0x20", 0, rst20, 16},                 // 0xE7
	{"ADD SP, 0x%02X", 1, addSPn, 16},          // 0xE8
	{"JP HL", 0, jpHL, 0},                      // 0xE9
	{"LD (0x%04X), A", 2, ldnnA, 16},           // 0xEA
	{"UNDEFINED", 0, undefined, 0},             // 0xEB
	{"UNDEFINED", 0, undefined, 0},             // 0xEC
	{"UNDEFINED", 0, undefined, 0},             // 0xED
	{"XOR 0x%02X", 1, xorn, 8},                 // 0xEE
	{"RST 0x28", 0, rst28, 16},                 // 0xEF
	{"LD A, (0xFF00+0x%02X)", 1, ldAff00n, 12}, // 0xF0
	{"POP AF", 0, popAF, 12},                   // 0xF1
	{"LD A, (0xFF00+C)", 0, ldAff00C, 8},       // 0xF2
	{"DI", 0, di, 4},                           // 0xF3
	{"UNDEFINED", 0, undefined, 0},             // 0xF4
	{"PUSH AF", 0, pushAF, 16},                 // 0xF5
	{"OR 0x%02X", 1, orn, 8},                   // 0xF6
	{"RST 0x30", 0, rst30, 16},                 // 0xF7
	{"LD HL, SP+0x%02X", 1, ldHLSPn, 12},       // 0xF8
	{"LD SP, HL", 0, ldSPHL, 8},                // 0xF9
	{"LD A, (0x%04X)", 2, ldAnn, 16},           // 0xFA
	{"EI", 0, ei, 4},                           // 0xFB
	{"UNDEFINED", 0, undefined, 0},             // 0xFC
	{"UNDEFINED", 0, undefined, 0},             // 0xFD
	{"CP 0x%02X", 1, cpn, 8},                   // 0xFE
	{"RST 0x38", 0, rst38, 16},                 // 0xFF
}

func New(memory *memory.Memory, interrupts *interrupts.Interrupts) *CPU {
	cpu := CPU{
		a:          SubRegister{},
		bc:         *newRegister(),
		de:         *newRegister(),
		hl:         *newRegister(),
		memory:     memory,
		interrupts: interrupts,
	}
	cpu.Reset()
	interrupts.RegisterHandlers(&cpu)
	return &cpu
}

func (cpu *CPU) Reset() {
	cpu.a.SetValue(0x01)
	cpu.f.setValue(0xb0) // flags z, h and c are set

	cpu.bc.SetValue(0x0013)
	cpu.de.SetValue(0x00d8)
	cpu.hl.SetValue(0x014d)

	cpu.sp = 0xfffe
	cpu.pc = 0x0000
}

// Step executes one instruction and returns the CPU cycles needed for the execution.
func (cpu *CPU) Step() int {

	if cpu.stopped {
		return 0
	}

	stepLogger := log.L().WithLazy(log.String("pc", fmt.Sprintf("0x%04X", cpu.pc)))

	opCode := cpu.memory.Read8BitValue(cpu.pc)
	instr := instructions[opCode]

	cpu.pc++ // advance program counter to next position

	var operand uint16
	switch instr.operandLength {
	case 0:
		break

	case 1:
		operand = uint16(cpu.memory.Read8BitValue(cpu.pc))
		cpu.pc++

	case 2:
		operand = cpu.memory.Read16BitValue(cpu.pc)
		cpu.pc += 2 // advance program counter two steps because we read 2 bytes
	}

	stepLogger.Info(instr.toString(operand))
	instr.execute(cpu, operand)
	cpu.stepCycles += int(instr.ticks)

	defer func() {
		cpu.stepCycles = 0
	}()

	return cpu.stepCycles
}

func (cpu *CPU) HandleVblankInterrupt() {
	cpu.call(0x40, &cpu.stepCycles, always())
}

func (cpu *CPU) HandleLcdStatInterrupt() {
	cpu.call(0x48, &cpu.stepCycles, always())
}

func (cpu *CPU) HandleTimerInterrupt() {
	cpu.call(0x50, &cpu.stepCycles, always())
}

func (cpu *CPU) HandleSerialInterrupt() {
	cpu.call(0x58, &cpu.stepCycles, always())
}

func (cpu *CPU) HandleJoypadInterrupt() {
	cpu.call(0x60, &cpu.stepCycles, always())
}

// Glossary:
// n 	8-bit number
// nn	16-bit number
// X	8-bit register (e.g. A, B)
// XX	16-bit register (e.g. BC, DE)
// DXX	dereferenced 16-bit register (in documentation referenced either as "(XX)" or "[XX]")

// 0x00
func nop(_ *CPU, _ uint16) {
	// Does nothing
}

// 0x01
func ldBCnn(cpu *CPU, operand uint16) { loadValueToRegister(&cpu.bc, operand) }

// 0x02
func ldDBCA(cpu *CPU, _ uint16) { cpu.load8BitValueToMemory(cpu.bc.GetValue(), cpu.a.GetValue()) }

// 0x03
func incBC(cpu *CPU, _ uint16) { incrementRegister(&cpu.bc) }

// 0x04
func incB(cpu *CPU, _ uint16) { incrementSubRegister(&cpu.bc.Hi, &cpu.f) }

// 0x05
func decB(cpu *CPU, _ uint16) { decrementSubRegister(&cpu.bc.Hi, &cpu.f) }

// 0x06
func ldBn(cpu *CPU, operand uint16) { loadValueToSubRegister(&cpu.bc.Hi, uint8(operand)) }

// 0x07
func rlca(cpu *CPU, _ uint16) {
	value := cpu.a.GetValue()

	// get the highest bit of value by masking it with 0b10000000 (aka 0x80 or 0d128) and shifting
	// it 7 bits down
	carry := (value & 0x80) >> 7

	if carry == 1 {
		cpu.f.setFlag(c)
	} else {
		cpu.f.unsetFlag(c)
	}

	// shift value 1bit up
	value <<= 1
	// append carry at its end
	value += carry

	cpu.a.SetValue(value)
	cpu.f.unsetFlag(z)
	cpu.f.unsetFlag(n)
	cpu.f.unsetFlag(h)
}

// 0x08
func ldDnnSP(cpu *CPU, operand uint16) { cpu.load16BitValueToMemory(operand, cpu.sp) }

// 0x09
func addHLBC(cpu *CPU, _ uint16) { add16BitValue(&cpu.hl, cpu.bc.GetValue(), &cpu.f) }

// 0x0A
func ldADBC(cpu *CPU, _ uint16) { cpu.loadMemoryToSubRegister(&cpu.a, cpu.bc.GetValue()) }

// 0x0B
func decBC(cpu *CPU, _ uint16) { decrementRegister(&cpu.bc) }

// 0x0C
func incC(cpu *CPU, _ uint16) { incrementSubRegister(&cpu.bc.Lo, &cpu.f) }

// 0x0D
func decC(cpu *CPU, _ uint16) { decrementSubRegister(&cpu.bc.Lo, &cpu.f) }

// 0x0E
func ldCn(cpu *CPU, operand uint16) { loadValueToSubRegister(&cpu.bc.Lo, byte(operand)) }

// 0x0F
func rrca(cpu *CPU, _ uint16) {
	value := cpu.a.GetValue()

	carry := value & 0x01

	if carry == 1 {
		cpu.f.setFlag(c)
	} else {
		cpu.f.unsetFlag(c)
	}

	// shift value 1bit down
	value >>= 1
	value += carry << 7

	cpu.a.SetValue(value)
	cpu.f.unsetFlag(z)
	cpu.f.unsetFlag(n)
	cpu.f.unsetFlag(h)
}

// 0x10
func stop(cpu *CPU, _ uint16) { cpu.stopped = true }

// 0x11
func ldDEnn(cpu *CPU, operand uint16) { loadValueToRegister(&cpu.de, operand) }

// 0x12
func ldDDEA(cpu *CPU, _ uint16) { cpu.load8BitValueToMemory(cpu.de.GetValue(), cpu.a.GetValue()) }

// 0x13
func incDE(cpu *CPU, _ uint16) { incrementRegister(&cpu.de) }

// 0x14
func incD(cpu *CPU, _ uint16) { incrementSubRegister(&cpu.de.Hi, &cpu.f) }

// 0x15
func decD(cpu *CPU, _ uint16) { decrementSubRegister(&cpu.de.Hi, &cpu.f) }

// 0x16
func ldDn(cpu *CPU, operand uint16) {
	loadValueToSubRegister(&cpu.de.Hi, uint8(operand))
}

// 0x17
func rla(cpu *CPU, _ uint16) {
	value := cpu.a.GetValue()

	carryFlagWasSet := cpu.f.isSet(c)

	// get the highest bit of value by masking it with 0b10000000 (aka 0x80 or 0d128) and shifting
	// it 7 bits down
	carry := (value & 0x80) >> 7

	if carry == 1 {
		cpu.f.setFlag(c)
	} else {
		cpu.f.unsetFlag(c)
	}

	// shift value 1bit up
	value <<= 1
	if carryFlagWasSet {
		value++
	}

	cpu.a.SetValue(value)
	cpu.f.unsetFlag(z)
	cpu.f.unsetFlag(n)
	cpu.f.unsetFlag(h)
}

// 0x18
func jrn(cpu *CPU, operand uint16) {
	relativeJump(&cpu.pc, int8(operand), &cpu.stepCycles, always())
}

// 0x19
func addHLDE(cpu *CPU, _ uint16) { add16BitValue(&cpu.hl, cpu.de.GetValue(), &cpu.f) }

// 0x1A
func ldADDE(cpu *CPU, _ uint16) { cpu.loadMemoryToSubRegister(&cpu.a, cpu.de.GetValue()) }

// 0x1B
func decDE(cpu *CPU, _ uint16) { decrementRegister(&cpu.de) }

// 0x1C
func incE(cpu *CPU, _ uint16) { incrementSubRegister(&cpu.de.Lo, &cpu.f) }

// 0x1D
func decE(cpu *CPU, _ uint16) { decrementSubRegister(&cpu.de.Lo, &cpu.f) }

// 0x1E
func ldEn(cpu *CPU, operand uint16) { loadValueToSubRegister(&cpu.de.Lo, byte(operand)) }

// 0x1F
func rra(cpu *CPU, _ uint16) {
	value := cpu.a.GetValue()
	carryFlagWasSet := cpu.f.isSet(c)

	carry := value & 0x01

	if carry == 1 {
		cpu.f.setFlag(c)
	} else {
		cpu.f.unsetFlag(c)
	}

	// shift value 1bit down
	value >>= 1
	if carryFlagWasSet {
		value += 1 << 7
	}

	cpu.a.SetValue(value)
	cpu.f.unsetFlag(z)
	cpu.f.unsetFlag(n)
	cpu.f.unsetFlag(h)
}

// 0x20
func jrnzn(cpu *CPU, operand uint16) {
	relativeJump(&cpu.pc, int8(operand), &cpu.stepCycles, onZNotSet(cpu))
}

// 0x21
func ldHLnn(cpu *CPU, operand uint16) { loadValueToRegister(&cpu.hl, operand) }

// 0x22
func ldiHLA(cpu *CPU, _ uint16) {
	cpu.load8BitValueToMemory(cpu.hl.GetValue(), cpu.a.GetValue())
	incrementRegister(&cpu.hl)
}

// 0x23
func incHL(cpu *CPU, _ uint16) { incrementRegister(&cpu.hl) }

// 0x24
func incH(cpu *CPU, _ uint16) { incrementSubRegister(&cpu.hl.Hi, &cpu.f) }

// 0x25
func decH(cpu *CPU, _ uint16) { decrementSubRegister(&cpu.hl.Hi, &cpu.f) }

// 0x26
func ldHn(cpu *CPU, operand uint16) { loadValueToSubRegister(&cpu.hl.Hi, uint8(operand)) }

// 0x27
// Implemented according to https://ehaskins.com/2018-01-30%20Z80%20DAA/
func daa(cpu *CPU, _ uint16) {
	var correction byte

	value := cpu.a.GetValue()

	if cpu.f.isSet(h) || (!cpu.f.isSet(n) && value&0xf > 0x9) {
		correction |= 0x6
	}

	if cpu.f.isSet(c) || (!cpu.f.isSet(n) && value > 0x99) {
		correction |= 0x60
		cpu.f.setFlag(c)
	}

	if cpu.f.isSet(n) {
		value -= correction
	} else {
		value += correction
	}

	if value == 0 {
		cpu.f.setFlag(z)
	} else {
		cpu.f.unsetFlag(z)
	}

	cpu.f.unsetFlag(h)

	cpu.a.SetValue(value)
}

// 0x28
func jrzn(cpu *CPU, operand uint16) {
	relativeJump(&cpu.pc, int8(operand), &cpu.stepCycles, onZSet(cpu))
}

// 0x29
func addHLHL(cpu *CPU, _ uint16) { add16BitValue(&cpu.hl, cpu.hl.GetValue(), &cpu.f) }

// 0x2A
func ldiAHL(cpu *CPU, _ uint16) {
	cpu.loadMemoryToSubRegister(&cpu.a, cpu.hl.GetValue())
	incrementRegister(&cpu.hl)
}

// 0x2B
func decHL(cpu *CPU, _ uint16) { decrementRegister(&cpu.hl) }

// 0x2C
func incL(cpu *CPU, _ uint16) { incrementSubRegister(&cpu.hl.Lo, &cpu.f) }

// 0x2D
func decL(cpu *CPU, _ uint16) { decrementSubRegister(&cpu.hl.Lo, &cpu.f) }

// 0x2E
func ldLn(cpu *CPU, operand uint16) { loadValueToSubRegister(&cpu.hl.Lo, byte(operand)) }

// 0x2F
func cpl(cpu *CPU, _ uint16) {
	cpu.a.SetValue(^cpu.a.GetValue())
	cpu.f.setFlag(n)
	cpu.f.setFlag(h)
}

// 0x30
func jrncn(cpu *CPU, operand uint16) {
	relativeJump(&cpu.pc, int8(operand), &cpu.stepCycles, onCNotSet(cpu))
}

// 0x31
func ldSPnn(cpu *CPU, operand uint16) { cpu.sp = operand }

// 0x32
func lddHLA(cpu *CPU, _ uint16) {
	cpu.load8BitValueToMemory(cpu.hl.GetValue(), cpu.a.GetValue())
	decrementRegister(&cpu.hl)
}

// 0x33
func incSP(cpu *CPU, _ uint16) { cpu.sp++ }

// 0x34
func incDHL(cpu *CPU, _ uint16) { cpu.incrementMemoryLocation(&cpu.hl, &cpu.f) }

// 0x35
func decDHL(cpu *CPU, _ uint16) { cpu.decrementMemoryLocation(&cpu.hl, &cpu.f) }

// 0x36
func ldDHLn(cpu *CPU, operand uint16) { cpu.load8BitValueToMemory(cpu.hl.GetValue(), byte(operand)) }

// 0x37
func scf(cpu *CPU, _ uint16) {
	cpu.f.unsetFlag(n)
	cpu.f.unsetFlag(h)
	cpu.f.setFlag(c)
}

// 0x38
func jrcn(cpu *CPU, operand uint16) {
	relativeJump(&cpu.pc, int8(operand), &cpu.stepCycles, onCSet(cpu))
}

// 0x39
func addHLSP(cpu *CPU, _ uint16) { add16BitValue(&cpu.hl, cpu.sp, &cpu.f) }

// 0x3A
func lddAHL(cpu *CPU, _ uint16) {
	cpu.loadMemoryToSubRegister(&cpu.a, cpu.hl.GetValue())
	decrementRegister(&cpu.hl)
}

// 0x3B
func decSP(cpu *CPU, _ uint16) { cpu.sp-- }

// 0x3C
func incA(cpu *CPU, _ uint16) { incrementSubRegister(&cpu.a, &cpu.f) }

// 0x3D
func decA(cpu *CPU, _ uint16) { decrementSubRegister(&cpu.a, &cpu.f) }

// 0x3E
func ldAn(cpu *CPU, operand uint16) { loadValueToSubRegister(&cpu.a, byte(operand)) }

// 0x3F
func ccf(cpu *CPU, _ uint16) {
	if cpu.f.isSet(c) {
		cpu.f.unsetFlag(c)
	} else {
		cpu.f.setFlag(c)
	}
	cpu.f.unsetFlag(n)
	cpu.f.unsetFlag(h)
}

// 0x40 NOP

// 0x41
func ldBC(cpu *CPU, _ uint16) { loadSubRegisterToSubRegister(&cpu.bc.Hi, &cpu.bc.Lo) }

// 0x42
func ldBD(cpu *CPU, _ uint16) { loadSubRegisterToSubRegister(&cpu.bc.Hi, &cpu.de.Hi) }

// 0x43
func ldBE(cpu *CPU, _ uint16) { loadSubRegisterToSubRegister(&cpu.bc.Hi, &cpu.de.Lo) }

// 0x44
func ldBH(cpu *CPU, _ uint16) { loadSubRegisterToSubRegister(&cpu.bc.Hi, &cpu.hl.Hi) }

// 0x45
func ldBL(cpu *CPU, _ uint16) { loadSubRegisterToSubRegister(&cpu.bc.Hi, &cpu.hl.Lo) }

// 0x46
func ldBDHL(cpu *CPU, _ uint16) { cpu.loadMemoryToSubRegister(&cpu.bc.Hi, cpu.hl.GetValue()) }

// 0x47
func ldBA(cpu *CPU, _ uint16) { loadSubRegisterToSubRegister(&cpu.bc.Hi, &cpu.a) }

// 0x48
func ldCB(cpu *CPU, _ uint16) { loadSubRegisterToSubRegister(&cpu.bc.Lo, &cpu.bc.Hi) }

// 0x49 NOP

// 0x4A
func ldCD(cpu *CPU, _ uint16) { loadSubRegisterToSubRegister(&cpu.bc.Lo, &cpu.de.Hi) }

// 0x4B
func ldCE(cpu *CPU, _ uint16) { loadSubRegisterToSubRegister(&cpu.bc.Lo, &cpu.de.Lo) }

// 0x4C
func ldCH(cpu *CPU, _ uint16) { loadSubRegisterToSubRegister(&cpu.bc.Lo, &cpu.hl.Hi) }

// 0x4D
func ldCL(cpu *CPU, _ uint16) { loadSubRegisterToSubRegister(&cpu.bc.Lo, &cpu.hl.Lo) }

// 0x4E
func ldCDHL(cpu *CPU, _ uint16) { cpu.loadMemoryToSubRegister(&cpu.bc.Lo, cpu.hl.GetValue()) }

// 0x4F
func ldCA(cpu *CPU, _ uint16) { loadSubRegisterToSubRegister(&cpu.bc.Lo, &cpu.a) }

// 0x50
func ldDB(cpu *CPU, _ uint16) { loadSubRegisterToSubRegister(&cpu.de.Hi, &cpu.bc.Hi) }

// 0x51
func ldDC(cpu *CPU, _ uint16) { loadSubRegisterToSubRegister(&cpu.de.Hi, &cpu.bc.Lo) }

// 0x52 NOP

// 0x53
func ldDE(cpu *CPU, _ uint16) { loadSubRegisterToSubRegister(&cpu.de.Hi, &cpu.de.Lo) }

// 0x54
func ldDH(cpu *CPU, _ uint16) { loadSubRegisterToSubRegister(&cpu.de.Hi, &cpu.hl.Hi) }

// 0x55
func ldDL(cpu *CPU, _ uint16) { loadSubRegisterToSubRegister(&cpu.de.Hi, &cpu.hl.Lo) }

// 0x56
func ldDDHL(cpu *CPU, _ uint16) { cpu.loadMemoryToSubRegister(&cpu.de.Hi, cpu.hl.GetValue()) }

// 0x57
func ldDA(cpu *CPU, _ uint16) { loadSubRegisterToSubRegister(&cpu.de.Hi, &cpu.a) }

// 0x58
func ldEB(cpu *CPU, _ uint16) { loadSubRegisterToSubRegister(&cpu.de.Lo, &cpu.bc.Hi) }

// 0x59
func ldEC(cpu *CPU, _ uint16) { loadSubRegisterToSubRegister(&cpu.de.Lo, &cpu.bc.Lo) }

// 0x5A
func ldED(cpu *CPU, _ uint16) { loadSubRegisterToSubRegister(&cpu.de.Lo, &cpu.de.Hi) }

// 0x5B NOP

// 0x5C
func ldEH(cpu *CPU, _ uint16) { loadSubRegisterToSubRegister(&cpu.de.Lo, &cpu.hl.Hi) }

// 0x5D
func ldEL(cpu *CPU, _ uint16) { loadSubRegisterToSubRegister(&cpu.de.Lo, &cpu.hl.Lo) }

// 0x5E
func ldEDHL(cpu *CPU, _ uint16) { cpu.loadMemoryToSubRegister(&cpu.de.Lo, cpu.hl.GetValue()) }

// 0x5F
func ldEA(cpu *CPU, _ uint16) { loadSubRegisterToSubRegister(&cpu.de.Lo, &cpu.a) }

// 0x60
func ldHB(cpu *CPU, _ uint16) { loadSubRegisterToSubRegister(&cpu.hl.Hi, &cpu.bc.Hi) }

// 0x61
func ldHC(cpu *CPU, _ uint16) { loadSubRegisterToSubRegister(&cpu.hl.Hi, &cpu.bc.Lo) }

// 0x62
func ldHD(cpu *CPU, _ uint16) { loadSubRegisterToSubRegister(&cpu.hl.Hi, &cpu.de.Hi) }

// 0x63
func ldHE(cpu *CPU, _ uint16) { loadSubRegisterToSubRegister(&cpu.hl.Hi, &cpu.de.Lo) }

// 0x64 NOP

// 0x65
func ldHL(cpu *CPU, _ uint16) { loadSubRegisterToSubRegister(&cpu.hl.Hi, &cpu.hl.Lo) }

// 0x66
func ldHDHL(cpu *CPU, _ uint16) { cpu.loadMemoryToSubRegister(&cpu.hl.Hi, cpu.hl.GetValue()) }

// 0x67
func ldHA(cpu *CPU, _ uint16) { loadSubRegisterToSubRegister(&cpu.hl.Hi, &cpu.a) }

// 0x68
func ldLB(cpu *CPU, _ uint16) { loadSubRegisterToSubRegister(&cpu.hl.Lo, &cpu.bc.Hi) }

// 0x69
func ldLC(cpu *CPU, _ uint16) { loadSubRegisterToSubRegister(&cpu.hl.Lo, &cpu.bc.Lo) }

// 0x6A
func ldLD(cpu *CPU, _ uint16) { loadSubRegisterToSubRegister(&cpu.hl.Lo, &cpu.de.Hi) }

// 0x6B
func ldLE(cpu *CPU, _ uint16) { loadSubRegisterToSubRegister(&cpu.hl.Lo, &cpu.de.Lo) }

// 0x6C
func ldLH(cpu *CPU, _ uint16) { loadSubRegisterToSubRegister(&cpu.hl.Lo, &cpu.hl.Hi) }

// 0x6D NOP

// 0x6E
func ldLDHL(cpu *CPU, _ uint16) { cpu.loadMemoryToSubRegister(&cpu.hl.Lo, cpu.hl.GetValue()) }

// 0x6F
func ldLA(cpu *CPU, _ uint16) { loadSubRegisterToSubRegister(&cpu.hl.Lo, &cpu.a) }

// 0x70
func ldDHLB(cpu *CPU, _ uint16) { cpu.load8BitValueToMemory(cpu.hl.GetValue(), cpu.bc.Hi.GetValue()) }

// 0x71
func ldDHLC(cpu *CPU, _ uint16) { cpu.load8BitValueToMemory(cpu.hl.GetValue(), cpu.bc.Lo.GetValue()) }

// 0x72
func ldDHLD(cpu *CPU, _ uint16) { cpu.load8BitValueToMemory(cpu.hl.GetValue(), cpu.de.Hi.GetValue()) }

// 0x73
func ldDHLE(cpu *CPU, _ uint16) { cpu.load8BitValueToMemory(cpu.hl.GetValue(), cpu.de.Lo.GetValue()) }

// 0x74
func ldDHLH(cpu *CPU, _ uint16) { cpu.load8BitValueToMemory(cpu.hl.GetValue(), cpu.hl.Hi.GetValue()) }

// 0x75
func ldDHLL(cpu *CPU, _ uint16) { cpu.load8BitValueToMemory(cpu.hl.GetValue(), cpu.hl.Lo.GetValue()) }

// 0x76
func halt(cpu *CPU, _ uint16) {
	if cpu.interrupts.IsMasterEnabled() {
		// If IME is set, do nothing and wait for interrupt handling
	} else {
		cpu.pc++
	}
}

// 0x77
func ldDHLA(cpu *CPU, _ uint16) { cpu.load8BitValueToMemory(cpu.hl.GetValue(), cpu.a.GetValue()) }

// 0x78
func ldAB(cpu *CPU, _ uint16) { loadSubRegisterToSubRegister(&cpu.a, &cpu.bc.Hi) }

// 0x79
func ldAC(cpu *CPU, _ uint16) { loadSubRegisterToSubRegister(&cpu.a, &cpu.bc.Lo) }

// 0x7A
func ldAD(cpu *CPU, _ uint16) { loadSubRegisterToSubRegister(&cpu.a, &cpu.de.Hi) }

// 0x7B
func ldAE(cpu *CPU, _ uint16) { loadSubRegisterToSubRegister(&cpu.a, &cpu.de.Lo) }

// 0x7C
func ldAH(cpu *CPU, _ uint16) { loadSubRegisterToSubRegister(&cpu.a, &cpu.hl.Hi) }

// 0x7D
func ldAL(cpu *CPU, _ uint16) { loadSubRegisterToSubRegister(&cpu.a, &cpu.hl.Lo) }

// 0x7E
func ldADHL(cpu *CPU, _ uint16) { cpu.loadMemoryToSubRegister(&cpu.a, cpu.hl.GetValue()) }

// 0x7F NOP

// 0x80
func addB(cpu *CPU, _ uint16) { add8BitValue(&cpu.a, cpu.bc.Hi.GetValue(), false, &cpu.f) }

// 0x81
func addC(cpu *CPU, _ uint16) { add8BitValue(&cpu.a, cpu.bc.Lo.GetValue(), false, &cpu.f) }

// 0x82
func addD(cpu *CPU, _ uint16) { add8BitValue(&cpu.a, cpu.de.Hi.GetValue(), false, &cpu.f) }

// 0x83
func addE(cpu *CPU, _ uint16) { add8BitValue(&cpu.a, cpu.de.Lo.GetValue(), false, &cpu.f) }

// 0x84
func addH(cpu *CPU, _ uint16) { add8BitValue(&cpu.a, cpu.hl.Hi.GetValue(), false, &cpu.f) }

// 0x85
func addL(cpu *CPU, _ uint16) { add8BitValue(&cpu.a, cpu.hl.Lo.GetValue(), false, &cpu.f) }

// 0x86
func addDHL(cpu *CPU, _ uint16) {
	add8BitValue(&cpu.a, cpu.memory.Read8BitValue(cpu.hl.GetValue()), false, &cpu.f)
}

// 0x87
func addA(cpu *CPU, _ uint16) { add8BitValue(&cpu.a, cpu.a.GetValue(), false, &cpu.f) }

// 0x88
func adcB(cpu *CPU, _ uint16) { add8BitValue(&cpu.a, cpu.bc.Hi.GetValue(), cpu.f.isSet(c), &cpu.f) }

// 0x89
func adcC(cpu *CPU, _ uint16) { add8BitValue(&cpu.a, cpu.bc.Lo.GetValue(), cpu.f.isSet(c), &cpu.f) }

// 0x8A
func adcD(cpu *CPU, _ uint16) { add8BitValue(&cpu.a, cpu.de.Hi.GetValue(), cpu.f.isSet(c), &cpu.f) }

// 0x8B
func adcE(cpu *CPU, _ uint16) { add8BitValue(&cpu.a, cpu.de.Lo.GetValue(), cpu.f.isSet(c), &cpu.f) }

// 0x8C
func adcH(cpu *CPU, _ uint16) { add8BitValue(&cpu.a, cpu.hl.Hi.GetValue(), cpu.f.isSet(c), &cpu.f) }

// 0x8D
func adcL(cpu *CPU, _ uint16) { add8BitValue(&cpu.a, cpu.hl.Lo.GetValue(), cpu.f.isSet(c), &cpu.f) }

// 0x8E
func adcDHL(cpu *CPU, _ uint16) {
	add8BitValue(&cpu.a, cpu.memory.Read8BitValue(cpu.hl.GetValue()), cpu.f.isSet(c), &cpu.f)
}

// 0x8F
func adcA(cpu *CPU, _ uint16) { add8BitValue(&cpu.a, cpu.a.GetValue(), cpu.f.isSet(c), &cpu.f) }

// 0x90
func subB(cpu *CPU, _ uint16) { subtract8BitValue(&cpu.a, cpu.bc.Hi.GetValue(), false, false, &cpu.f) }

// 0x91
func subC(cpu *CPU, _ uint16) { subtract8BitValue(&cpu.a, cpu.bc.Lo.GetValue(), false, false, &cpu.f) }

// 0x92
func subD(cpu *CPU, _ uint16) { subtract8BitValue(&cpu.a, cpu.de.Hi.GetValue(), false, false, &cpu.f) }

// 0x93
func subE(cpu *CPU, _ uint16) { subtract8BitValue(&cpu.a, cpu.de.Lo.GetValue(), false, false, &cpu.f) }

// 0x94
func subH(cpu *CPU, _ uint16) { subtract8BitValue(&cpu.a, cpu.hl.Hi.GetValue(), false, false, &cpu.f) }

// 0x95
func subL(cpu *CPU, _ uint16) { subtract8BitValue(&cpu.a, cpu.hl.Lo.GetValue(), false, false, &cpu.f) }

// 0x96
func subDHL(cpu *CPU, _ uint16) {
	subtract8BitValue(&cpu.a, cpu.memory.Read8BitValue(cpu.hl.GetValue()), false, false, &cpu.f)
}

// 0x97
func subA(cpu *CPU, _ uint16) { subtract8BitValue(&cpu.a, cpu.a.GetValue(), false, false, &cpu.f) }

// 0x98
func sbcB(cpu *CPU, _ uint16) {
	subtract8BitValue(&cpu.a, cpu.bc.Hi.GetValue(), cpu.f.isSet(c), false, &cpu.f)
}

// 0x99
func sbcC(cpu *CPU, _ uint16) {
	subtract8BitValue(&cpu.a, cpu.bc.Lo.GetValue(), cpu.f.isSet(c), false, &cpu.f)
}

// 0x9A
func sbcD(cpu *CPU, _ uint16) {
	subtract8BitValue(&cpu.a, cpu.de.Hi.GetValue(), cpu.f.isSet(c), false, &cpu.f)
}

// 0x9B
func sbcE(cpu *CPU, _ uint16) {
	subtract8BitValue(&cpu.a, cpu.de.Lo.GetValue(), cpu.f.isSet(c), false, &cpu.f)
}

// 0x9C
func sbcH(cpu *CPU, _ uint16) {
	subtract8BitValue(&cpu.a, cpu.hl.Hi.GetValue(), cpu.f.isSet(c), false, &cpu.f)
}

// 0x9D
func sbcL(cpu *CPU, _ uint16) {
	subtract8BitValue(&cpu.a, cpu.hl.Lo.GetValue(), cpu.f.isSet(c), false, &cpu.f)
}

// 0x9E
func sbcDHL(cpu *CPU, _ uint16) {
	subtract8BitValue(&cpu.a, cpu.memory.Read8BitValue(cpu.hl.GetValue()), cpu.f.isSet(c), false, &cpu.f)
}

// 0x9F
func sbcA(cpu *CPU, _ uint16) {
	subtract8BitValue(&cpu.a, cpu.a.GetValue(), cpu.f.isSet(c), false, &cpu.f)
}

// 0xA0
func andB(cpu *CPU, _ uint16) { and(&cpu.a, cpu.bc.Hi.GetValue(), &cpu.f) }

// 0xA1
func andC(cpu *CPU, _ uint16) { and(&cpu.a, cpu.bc.Lo.GetValue(), &cpu.f) }

// 0xA2
func andD(cpu *CPU, _ uint16) { and(&cpu.a, cpu.de.Hi.GetValue(), &cpu.f) }

// 0xA3
func andE(cpu *CPU, _ uint16) { and(&cpu.a, cpu.de.Lo.GetValue(), &cpu.f) }

// 0xA4
func andH(cpu *CPU, _ uint16) { and(&cpu.a, cpu.hl.Hi.GetValue(), &cpu.f) }

// 0xA5
func andL(cpu *CPU, _ uint16) { and(&cpu.a, cpu.hl.Lo.GetValue(), &cpu.f) }

// 0xA6
func andDHL(cpu *CPU, _ uint16) { and(&cpu.a, cpu.memory.Read8BitValue(cpu.hl.GetValue()), &cpu.f) }

// 0xA7
func andA(cpu *CPU, _ uint16) { and(&cpu.a, cpu.a.GetValue(), &cpu.f) }

// 0xA8
func xorB(cpu *CPU, _ uint16) { xor(&cpu.a, cpu.bc.Hi.GetValue(), &cpu.f) }

// 0xA9
func xorC(cpu *CPU, _ uint16) { xor(&cpu.a, cpu.bc.Lo.GetValue(), &cpu.f) }

// 0xAA
func xorD(cpu *CPU, _ uint16) { xor(&cpu.a, cpu.de.Hi.GetValue(), &cpu.f) }

// 0xAB
func xorE(cpu *CPU, _ uint16) { xor(&cpu.a, cpu.de.Lo.GetValue(), &cpu.f) }

// 0xAC
func xorH(cpu *CPU, _ uint16) { xor(&cpu.a, cpu.hl.Hi.GetValue(), &cpu.f) }

// 0xAD
func xorL(cpu *CPU, _ uint16) { xor(&cpu.a, cpu.hl.Lo.GetValue(), &cpu.f) }

// 0xAD
func xorDHL(cpu *CPU, _ uint16) { xor(&cpu.a, cpu.memory.Read8BitValue(cpu.hl.GetValue()), &cpu.f) }

// 0xAF
func xorA(cpu *CPU, _ uint16) { xor(&cpu.a, cpu.a.GetValue(), &cpu.f) }

// 0xB0
func orB(cpu *CPU, _ uint16) { or(&cpu.a, cpu.bc.Hi.GetValue(), &cpu.f) }

// 0xB1
func orC(cpu *CPU, _ uint16) { or(&cpu.a, cpu.bc.Lo.GetValue(), &cpu.f) }

// 0xB2
func orD(cpu *CPU, _ uint16) { or(&cpu.a, cpu.de.Hi.GetValue(), &cpu.f) }

// 0xB3
func orE(cpu *CPU, _ uint16) { or(&cpu.a, cpu.de.Lo.GetValue(), &cpu.f) }

// 0xB4
func orH(cpu *CPU, _ uint16) { or(&cpu.a, cpu.hl.Hi.GetValue(), &cpu.f) }

// 0xB5
func orL(cpu *CPU, _ uint16) { or(&cpu.a, cpu.hl.Lo.GetValue(), &cpu.f) }

// 0xB6
func orDHL(cpu *CPU, _ uint16) { or(&cpu.a, cpu.memory.Read8BitValue(cpu.hl.GetValue()), &cpu.f) }

// 0xB7
func orA(cpu *CPU, _ uint16) { or(&cpu.a, cpu.a.GetValue(), &cpu.f) }

// 0xB8
func cpB(cpu *CPU, _ uint16) { subtract8BitValue(&cpu.a, cpu.bc.Hi.GetValue(), false, true, &cpu.f) }

// 0xB9
func cpC(cpu *CPU, _ uint16) { subtract8BitValue(&cpu.a, cpu.bc.Lo.GetValue(), false, true, &cpu.f) }

// 0xBA
func cpD(cpu *CPU, _ uint16) { subtract8BitValue(&cpu.a, cpu.de.Hi.GetValue(), false, true, &cpu.f) }

// 0xBB
func cpE(cpu *CPU, _ uint16) { subtract8BitValue(&cpu.a, cpu.de.Lo.GetValue(), false, true, &cpu.f) }

// 0xBC
func cpH(cpu *CPU, _ uint16) { subtract8BitValue(&cpu.a, cpu.hl.Hi.GetValue(), false, true, &cpu.f) }

// 0xBD
func cpL(cpu *CPU, _ uint16) { subtract8BitValue(&cpu.a, cpu.hl.Lo.GetValue(), false, true, &cpu.f) }

// 0xBE
func cpDHL(cpu *CPU, _ uint16) {
	subtract8BitValue(&cpu.a, cpu.memory.Read8BitValue(cpu.hl.GetValue()), false, true, &cpu.f)
}

// 0xBF
func cpA(cpu *CPU, _ uint16) { subtract8BitValue(&cpu.a, cpu.a.GetValue(), false, true, &cpu.f) }

// 0xC0
func retnz(cpu *CPU, _ uint16) {
	cpu.returnFromSubroutine(&cpu.stepCycles, onZNotSet(cpu))
}

// 0xC1
func popBC(cpu *CPU, _ uint16) { cpu.bc.SetValue(cpu.popFromStack()) }

// 0xC2
func jpnznn(cpu *CPU, operand uint16) {
	jump(&cpu.pc, operand, &cpu.stepCycles, onZNotSet(cpu))
}

// 0xC3
func jpnn(cpu *CPU, operand uint16) { jump(&cpu.pc, operand, &cpu.stepCycles, always()) }

// 0xC4
func callnznn(cpu *CPU, operand uint16) {
	cpu.call(operand, &cpu.stepCycles, onZNotSet(cpu))
}

// 0xC5
func pushBC(cpu *CPU, _ uint16) { cpu.pushToStack(cpu.bc.GetValue()) }

// 0xC6
func addn(cpu *CPU, operand uint16) { add8BitValue(&cpu.a, byte(operand), false, &cpu.f) }

// 0xC7
func rst00(cpu *CPU, _ uint16) {
	cpu.pushToStack(cpu.pc)
	cpu.pc = 0x00
}

// 0xC8
func retz(cpu *CPU, _ uint16) {
	cpu.returnFromSubroutine(&cpu.stepCycles, onZSet(cpu))
}

// 0xC9
func ret(cpu *CPU, _ uint16) { cpu.returnFromSubroutine(&cpu.stepCycles, always()) }

// 0xCA
func jpznn(cpu *CPU, operand uint16) {
	jump(&cpu.pc, operand, &cpu.stepCycles, onZSet(cpu))
}

// 0xCB
func cbn(cpu *CPU, operand uint16) { executeExtendedInstruction(cpu, &cpu.stepCycles, byte(operand)) }

// 0xCC
func callznn(cpu *CPU, operand uint16) {
	cpu.call(operand, &cpu.stepCycles, onZSet(cpu))
}

// 0xCD
func callnn(cpu *CPU, operand uint16) { cpu.call(operand, &cpu.stepCycles, always()) }

// 0xCE
func adcn(cpu *CPU, operand uint16) { add8BitValue(&cpu.a, byte(operand), cpu.f.isSet(c), &cpu.f) }

// 0xCF
func rst08(cpu *CPU, _ uint16) {
	cpu.pushToStack(cpu.pc)
	cpu.pc = 0x08
}

// 0xD0
func retnc(cpu *CPU, _ uint16) {
	cpu.returnFromSubroutine(&cpu.stepCycles, onCNotSet(cpu))
}

// 0xD1
func popDE(cpu *CPU, _ uint16) { cpu.de.SetValue(cpu.popFromStack()) }

// 0xD2
func jpncnn(cpu *CPU, operand uint16) {
	jump(&cpu.pc, operand, &cpu.stepCycles, onCNotSet(cpu))
}

// 0xD3 UNDEFINED

// 0xD4
func callncnn(cpu *CPU, operand uint16) {
	cpu.call(operand, &cpu.stepCycles, onCNotSet(cpu))
}

// 0xD5
func pushDE(cpu *CPU, _ uint16) { cpu.pushToStack(cpu.de.GetValue()) }

// 0xD6
func subn(cpu *CPU, operand uint16) { subtract8BitValue(&cpu.a, byte(operand), false, false, &cpu.f) }

// 0xD7
func rst10(cpu *CPU, _ uint16) {
	cpu.pushToStack(cpu.pc)
	cpu.pc = 0x10
}

// 0xD8
func retc(cpu *CPU, _ uint16) {
	cpu.returnFromSubroutine(&cpu.stepCycles, onCSet(cpu))
}

func reti(cpu *CPU, _ uint16) {
	ei(cpu, 0)
	ret(cpu, 0)
}

// 0xDA
func jpcnn(cpu *CPU, operand uint16) {
	jump(&cpu.pc, operand, &cpu.stepCycles, onCSet(cpu))
}

// 0xDB UNDEFINED

// 0xDC
func callcnn(cpu *CPU, operand uint16) {
	cpu.call(operand, &cpu.stepCycles, onCSet(cpu))
}

// 0xDD UNDEFINED

// 0xDE
func sbcn(cpu *CPU, operand uint16) {
	subtract8BitValue(&cpu.a, byte(operand), cpu.f.isSet(c), false, &cpu.f)
}

// 0xDF
func rst18(cpu *CPU, _ uint16) {
	cpu.pushToStack(cpu.pc)
	cpu.pc = 0x18
}

// 0xE0
func ldff00nA(cpu *CPU, operand uint16) {
	cpu.load8BitValueToMemory(0xff00+operand, cpu.a.GetValue())
}

// 0xE1
func popHL(cpu *CPU, _ uint16) { cpu.hl.SetValue(cpu.popFromStack()) }

// 0xE2
func ldff00CA(cpu *CPU, _ uint16) {
	cpu.load8BitValueToMemory(0xff00+uint16(cpu.bc.Lo.GetValue()), cpu.a.GetValue())
}

// 0xE3 UNDEFINED

// 0xE4 UNDEFINED

// 0xE5
func pushHL(cpu *CPU, _ uint16) { cpu.pushToStack(cpu.hl.GetValue()) }

// 0xE6
func andn(cpu *CPU, operand uint16) { and(&cpu.a, byte(operand), &cpu.f) }

// 0xE7
func rst20(cpu *CPU, _ uint16) {
	cpu.pushToStack(cpu.pc)
	cpu.pc = 0x20
}

// 0xE8
func addSPn(cpu *CPU, operand uint16) {
	signedValue := int8(byte(operand))

	var result uint16
	if signedValue < 0 {
		absVal := uint16(0x00 - signedValue)
		result = cpu.sp - absVal

		if absVal&0x00FF > cpu.sp&0x0FF {
			cpu.f.setFlag(h)
		} else {
			cpu.f.unsetFlag(h)
		}

		if absVal > cpu.sp {
			cpu.f.setFlag(c)
		} else {
			cpu.f.unsetFlag(c)
		}

	} else {
		result = cpu.sp + operand

		if (operand&0x0FFF)+(cpu.sp&0x0FFF) > 0x0FFF {
			cpu.f.setFlag(h)
		} else {
			cpu.f.unsetFlag(h)
		}

		if result < cpu.sp {
			cpu.f.setFlag(c)
		} else {
			cpu.f.unsetFlag(c)
		}
	}

	cpu.f.unsetFlag(z)
	cpu.f.unsetFlag(n)

	cpu.sp = result
}

// 0xE9
func jpHL(cpu *CPU, _ uint16) {
	jump(&cpu.pc, cpu.hl.GetValue(), &cpu.stepCycles, always())
}

// 0xEA
func ldnnA(cpu *CPU, operand uint16) { cpu.load8BitValueToMemory(operand, cpu.a.GetValue()) }

// 0xEB UNDEFINED

// 0xEC UNDEFINED

// 0xED UNDEFINED

// 0xEE
func xorn(cpu *CPU, operand uint16) { xor(&cpu.a, byte(operand), &cpu.f) }

// 0xEF
func rst28(cpu *CPU, _ uint16) {
	cpu.pushToStack(cpu.pc)
	cpu.pc = 0x28
}

// 0xF0
func ldAff00n(cpu *CPU, operand uint16) {
	cpu.loadMemoryToSubRegister(&cpu.a, 0xff00+operand)
}

// 0xF1
func popAF(cpu *CPU, _ uint16) {
	af := cpu.popFromStack()
	cpu.a.SetValue(byte(af >> 8))
	cpu.f = flags{byte(af)}
}

// 0xF2
func ldAff00C(cpu *CPU, _ uint16) {
	cpu.loadMemoryToSubRegister(&cpu.a, 0xff00+uint16(cpu.bc.Lo.GetValue()))
}

// 0xF3
func di(cpu *CPU, _ uint16) { cpu.interrupts.SetMasterEnable(false) }

// 0xF4 UNDEFINED

// 0xF5
func pushAF(cpu *CPU, _ uint16) {
	cpu.pushToStack(uint16(cpu.a.GetValue())<<8 + uint16(cpu.f.getValue()))
}

// 0xF6
func orn(cpu *CPU, operand uint16) { or(&cpu.a, byte(operand), &cpu.f) }

// 0xF7
func rst30(cpu *CPU, _ uint16) {
	cpu.pushToStack(cpu.pc)
	cpu.pc = 0x30
}

// 0xF8
func ldHLSPn(cpu *CPU, operand uint16) {
	signedValue := int8(byte(operand))

	var result uint16
	if signedValue < 0 {
		absVal := uint16(0x00 - signedValue)
		result = cpu.sp - absVal

		if absVal&0x00FF > cpu.sp&0x0FF {
			cpu.f.setFlag(h)
		} else {
			cpu.f.unsetFlag(h)
		}

		if absVal > cpu.sp {
			cpu.f.setFlag(c)
		} else {
			cpu.f.unsetFlag(c)
		}

	} else {
		result = cpu.sp + operand

		if (operand&0x0FFF)+(cpu.sp&0x0FFF) > 0x0FFF {
			cpu.f.setFlag(h)
		} else {
			cpu.f.unsetFlag(h)
		}

		if result < cpu.sp {
			cpu.f.setFlag(c)
		} else {
			cpu.f.unsetFlag(c)
		}
	}

	cpu.f.unsetFlag(z)
	cpu.f.unsetFlag(n)

	loadValueToRegister(&cpu.hl, result)
}

// 0xF9
func ldSPHL(cpu *CPU, _ uint16) { cpu.sp = cpu.hl.GetValue() }

// 0xFA
func ldAnn(cpu *CPU, operand uint16) { cpu.loadMemoryToSubRegister(&cpu.a, operand) }

// 0xFB
func ei(cpu *CPU, _ uint16) { cpu.interrupts.SetMasterEnable(true) }

// 0xFC UNDEFINED

// 0xFD UNDEFINED

// 0xFE
func cpn(cpu *CPU, operand uint16) { subtract8BitValue(&cpu.a, byte(operand), false, true, &cpu.f) }

// 0xFF
func rst38(cpu *CPU, _ uint16) {
	cpu.pushToStack(cpu.pc)
	cpu.pc = 0x38
}

func undefined(cpu *CPU, _ uint16) {
	cpu.pc-- // step PC one down because this is an undefined instruction and must not have any effect
}

// loadSubRegisterToSubRegister copies (aka loads) the value in sub-register op2 into the sub-register op1.
func loadSubRegisterToSubRegister(op1, op2 *SubRegister) {
	op1.SetValue(op2.GetValue())
}

// loadValueToSubRegister copies the value op2 into sub-register op1.
func loadValueToSubRegister(sr *SubRegister, value byte) {
	sr.SetValue(value)
}

// loadValueToRegister copies the value op2 into register op1.
func loadValueToRegister(r *Register, value uint16) {
	r.SetValue(value)
}

// load16BitValueToMemory copies the given value to memory at given address
func (cpu *CPU) load16BitValueToMemory(address uint16, value uint16) {
	cpu.memory.Write16BitValue(address, value)
}

// load8BitValueToMemory copies the given value to memory at given address
func (cpu *CPU) load8BitValueToMemory(address uint16, value byte) {
	cpu.memory.Write8BitValue(address, value)
}

// loadMemoryToSubRegister copies from memory addressed by value of register addressRegister to the given sub-register sr
func (cpu *CPU) loadMemoryToSubRegister(sr *SubRegister, address uint16) {
	loadValueToSubRegister(sr, cpu.memory.Read8BitValue(address))
}

func incrementRegister(op1 *Register) {
	op1.Increment()
}

func decrementRegister(op1 *Register) {
	op1.Decrement()
}

func incrementSubRegister(op1 *SubRegister, flags *flags) {

	if op1.GetValue() == 0x0F {
		flags.setFlag(h)
	} else {
		flags.unsetFlag(h)
	}

	op1.Increment()

	if op1.GetValue() == 0 {
		flags.setFlag(z)
	} else {
		flags.unsetFlag(z)
	}

	flags.unsetFlag(n)
}

func decrementSubRegister(op1 *SubRegister, flags *flags) {

	if op1.GetValue() == 0x10 {
		flags.setFlag(h)
	} else {
		flags.unsetFlag(h)
	}

	op1.Decrement()

	if op1.GetValue() == 0 {
		flags.setFlag(z)
	} else {
		flags.unsetFlag(z)
	}

	flags.setFlag(n)
}

func (cpu *CPU) incrementMemoryLocation(addressRegister *Register, flags *flags) {
	value := cpu.memory.Read8BitValue(addressRegister.GetValue())
	if value == 0x0F {
		flags.setFlag(h)
	} else {
		flags.unsetFlag(h)
	}

	value++

	cpu.memory.Write8BitValue(addressRegister.GetValue(), value)
	if value == 0 {
		flags.setFlag(z)
	} else {
		flags.unsetFlag(z)
	}

	flags.unsetFlag(n)
}

func (cpu *CPU) decrementMemoryLocation(addressRegister *Register, flags *flags) {
	value := cpu.memory.Read8BitValue(addressRegister.GetValue())
	if value == 0x10 {
		flags.setFlag(h)
	} else {
		flags.unsetFlag(h)
	}

	value--
	cpu.memory.Write8BitValue(addressRegister.GetValue(), value)
	if value == 0 {
		flags.setFlag(z)
	} else {
		flags.unsetFlag(z)
	}

	flags.setFlag(n)
}

func relativeJump(pc *uint16, offset int8, cycles *int, predicate func() bool) {
	// if condition is not met don't jump
	if !predicate() {
		return
	}

	if offset < 0 {
		*pc -= uint16(0 - offset)
	} else {
		*pc += uint16(offset)
	}

	*cycles += 4 // jumping takes additional 4 stepCycles

	log.L().Debug(fmt.Sprintf(" <<PC set to 0x%04X>>", *pc))
}

func jump(pc *uint16, target uint16, cycles *int, predicate func() bool) {
	// if condition is not met don't jump
	if !predicate() {
		return
	}

	*pc = target
	*cycles += 4 // jumping takes additional 4 stepCycles

	log.L().Debug(fmt.Sprintf(" <<PC set to 0x%04X>>", *pc))
}

func add16BitValue(op1 *Register, value uint16, flags *flags) {
	oldValue := op1.GetValue()

	newValue := oldValue + value

	if newValue < oldValue {
		flags.setFlag(c)
	} else {
		flags.unsetFlag(c)
	}

	// Check for half-carry flag: mask the highest of the four nibbles (16 bits = 2 bytes = 4 nibbles) on both target and
	// operand and check if the result is bigger than 0b0000111111111111 (aka 0x0FFF)
	if (oldValue&0x0FFF)+(value&0x0FFF) > 0x0FFF {
		flags.setFlag(h)
	} else {
		flags.unsetFlag(h)
	}

	op1.SetValue(newValue)

	flags.unsetFlag(n)
}

func add8BitValue(op1 *SubRegister, value byte, carryIn bool, flags *flags) {
	oldValue := op1.GetValue()

	newValue := oldValue + value
	if carryIn {
		newValue++
	}

	if (newValue < oldValue) || (carryIn && newValue == oldValue) {
		flags.setFlag(c)
	} else {
		flags.unsetFlag(c)
	}

	// Check for half-carry flag: mask the upper nibble (1 byte = 2 nibbles) on both target and
	// operand and check if the result is bigger than 0b00001111 (aka 0x0F)
	if ((oldValue&0x0F)+(value&0x0F) > 0x0F) || (carryIn && (oldValue&0x0F)+(value&0x0F)+1 > 0x0F) {
		flags.setFlag(h)
	} else {
		flags.unsetFlag(h)
	}

	// 8bit addition sets z flag (in contrast to 16bit addition)
	if newValue == 0x00 {
		flags.setFlag(z)
	} else {
		flags.unsetFlag(z)
	}

	op1.SetValue(newValue)

	flags.unsetFlag(n)
}

func subtract8BitValue(sr *SubRegister, value byte, carryIn bool, compare bool, flags *flags) {
	oldValue := sr.GetValue()

	newValue := oldValue - value

	if (value > oldValue) || (carryIn && value+1 > oldValue) {
		flags.setFlag(c)
	} else {
		flags.unsetFlag(c)
	}

	// Check for half-carry flag: mask the upper nibble (1 byte = 2 nibbles) on both target and
	// operand and check if the masked operand is bigger than the masked old value (target)
	if (value&0x0F > oldValue&0x0f) || (carryIn && (value&0x0F)+1 > oldValue&0x0f) {
		flags.setFlag(h)
	} else {
		flags.unsetFlag(h)
	}

	if newValue == 0x00 {
		flags.setFlag(z)
	} else {
		flags.unsetFlag(z)
	}

	flags.setFlag(n)
	if !compare {
		sr.SetValue(newValue)
	}
}

func and(sr *SubRegister, value byte, flags *flags) {
	result := sr.GetValue() & value
	sr.SetValue(result)

	if result == 0x00 {
		flags.setFlag(z)
	} else {
		flags.unsetFlag(z)
	}

	flags.unsetFlag(n)
	flags.setFlag(h)
	flags.unsetFlag(c)
}

func xor(sr *SubRegister, value byte, flags *flags) {
	result := sr.GetValue() ^ value
	sr.SetValue(result)

	if result == 0x00 {
		flags.setFlag(z)
	} else {
		flags.unsetFlag(z)
	}

	flags.unsetFlag(n)
	flags.unsetFlag(h)
	flags.unsetFlag(c)
}

func or(sr *SubRegister, value byte, flags *flags) {
	result := sr.GetValue() | value
	sr.SetValue(result)

	if result == 0x00 {
		flags.setFlag(z)
	} else {
		flags.unsetFlag(z)
	}

	flags.unsetFlag(n)
	flags.unsetFlag(h)
	flags.unsetFlag(c)
}

func (cpu *CPU) pushToStack(value uint16) {
	hiByte := uint8(value >> 8)
	loByte := uint8(value)

	cpu.sp--
	cpu.memory.Write8BitValue(cpu.sp, hiByte)
	cpu.sp--
	cpu.memory.Write8BitValue(cpu.sp, loByte)
}

func (cpu *CPU) popFromStack() uint16 {
	loByte := cpu.memory.Read8BitValue(cpu.sp)
	cpu.sp++
	hiByte := cpu.memory.Read8BitValue(cpu.sp)
	cpu.sp++

	value := (uint16(hiByte) << 8) + uint16(loByte)
	return value
}

func (cpu *CPU) returnFromSubroutine(cycles *int, predicate func() bool) {
	if predicate() {
		cpu.pc = cpu.popFromStack()
		*cycles += 12 // Stack access takes 12 additional stepCycles
	}
}

func (cpu *CPU) call(operand uint16, cycles *int, predicate func() bool) {
	if predicate() {
		cpu.pushToStack(cpu.pc)
		jump(&cpu.pc, operand, cycles, always()) // adds 4 stepCycles
		*cycles += 8                             // add 8 more stepCycles to get to 12 additional stepCycles in case of branching
	}
}
