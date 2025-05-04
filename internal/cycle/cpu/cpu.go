package cpu

import (
	"fmt"
	"gameboy-emulator/internal/cycle/interrupts"
	"gameboy-emulator/internal/cycle/memory"
	"gameboy-emulator/internal/util"
	log "go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	executing cpuState = iota
	halted
	stopped
)

type (
	CPU struct {

		// 8bit registers
		a byte
		f flags

		b, c byte
		d, e byte
		h, l byte

		// Temporary registers
		z byte
		w byte

		// 16bit registers
		pc uint16 // Program counter
		sp uint16 // Stack pointer

		ir              instruction // instruction register
		pcOfInstruction uint16

		ops *util.Queue[func(*CPU)]

		mmu        *memory.Memory
		interrupts *interrupts.Interrupts

		ticks  byte
		state  cpuState
		Cycles uint64
	}

	instruction struct {
		disassembly string
		execute     func(*CPU)
	}

	cpuState byte
)

func New(memory *memory.Memory, interrupts *interrupts.Interrupts) *CPU {
	cpu := CPU{
		mmu:        memory,
		interrupts: interrupts,
		ops:        &util.Queue[func(*CPU)]{},
	}
	cpu.Reset()
	return &cpu
}

func (c *CPU) Reset() {
	c.a = 0x01
	c.f = 0xb0 // flags zero, halfCarry and carry are set

	c.b = 0x00
	c.c = 0x13
	c.d = 0x00
	c.e = 0xD8
	c.h = 0x01
	c.l = 0x4D

	c.w = 0x00
	c.z = 0x00

	c.sp = 0xfffe
	c.pc = 0x0000

	c.ir = instructions[0x00] // On startup CPU has NOP loaded

	c.ticks = 0
	c.state = executing
	c.pcOfInstruction = 0
	c.ops.Clear()
}

func (c *CPU) Tick() {
	switch c.state {
	case executing:
		// Only do something every 4 Cycles
		c.ticks++
		if c.ticks < 4 {
			return
		}
		c.ticks = 0

		if c.ops.Size() == 0 {
			if c.ir.execute == nil {
				log.L().Panic("Undefined instruction", log.String("pc", fmt.Sprintf("0x%04x", c.pcOfInstruction)))
			}

			if log.L().Level() <= zapcore.InfoLevel {
				log.L().Info(c.ir.disassembly, log.String("dump",
					fmt.Sprintf("A:%02X F:%02X B:%02X C:%02X D:%02X E:%02X H:%02X L:%02X SP:%04X PC:%04X PCMEM:%02X,%02X,%02X,%02X",
						c.a, c.f, c.b, c.c, c.d, c.e, c.h, c.l, c.sp, c.pc-1, c.mmu.Read(c.pc-1), c.mmu.Read(c.pc), c.mmu.Read(c.pc+1), c.mmu.Read(c.pc+2))))
			}
			c.ir.execute(c)
		}
	case halted:
		if !c.interrupts.InterruptsPending() {
			return
		}

		c.state = executing
		fetchCycle(c)

	case stopped:
		log.L().Warn("CPU is stopped!")
		return
	}

	opItem, _ := c.ops.Pop()
	opItem(c)
	c.Cycles++
}

func (c *CPU) bc() uint16 {
	return uint16(c.b)<<8 | uint16(c.c)
}

func (c *CPU) writeBC(data uint16) {
	c.b = byte(data >> 8)
	c.c = byte(data)
}

func (c *CPU) de() uint16 {
	return uint16(c.d)<<8 | uint16(c.e)
}

func (c *CPU) writeDE(data uint16) {
	c.d = byte(data >> 8)
	c.e = byte(data)
}

func (c *CPU) hl() uint16 {
	return uint16(c.h)<<8 | uint16(c.l)
}

func (c *CPU) writeHL(data uint16) {
	c.h = byte(data >> 8)
	c.l = byte(data)
}

func (c *CPU) wz() uint16 {
	return uint16(c.w)<<8 | uint16(c.z)
}
