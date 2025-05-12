package emulation

import (
	"gameboy-emulator/internal/cartridge"
	"gameboy-emulator/internal/cycle/apu"
	"gameboy-emulator/internal/cycle/cpu"
	"gameboy-emulator/internal/cycle/gpu"
	"gameboy-emulator/internal/cycle/interrupts"
	"gameboy-emulator/internal/cycle/joypad"
	"gameboy-emulator/internal/cycle/memory"
	"gameboy-emulator/internal/cycle/timer"
)

// Core of the Gameboy emulation. Holds all components and exposes
// functions for manipulating and getting feedback.
type Core struct {
	interrupts *interrupts.Interrupts
	joypad     *joypad.Joypad
	timer      *timer.Timer
	ppu        *gpu.PPU
	memory     *memory.Memory
	cpu        *cpu.CPU
	apu        *apu.APU
}

func NewCore(
	interrupts *interrupts.Interrupts,
	joypad *joypad.Joypad,
	timer *timer.Timer,
	ppu *gpu.PPU,
	memory *memory.Memory,
	cpu *cpu.CPU,
	apu *apu.APU,
) *Core {
	return &Core{
		interrupts: interrupts,
		joypad:     joypad,
		timer:      timer,
		ppu:        ppu,
		memory:     memory,
		cpu:        cpu,
		apu:        apu,
	}
}

func (e *Core) Reset() {
	e.interrupts.Reset()
	e.joypad.Reset()
	e.timer.Reset()
	e.ppu.Reset()
	e.memory.Reset()
	e.cpu.Reset()
	e.apu.Reset()
}

func (e *Core) SetScreenHandler(handler func([144][160]byte)) {
	e.ppu.GetDisplay().RegisterFrameOutputHandler(handler)
}

func (e *Core) InsertCartridge(pathToCartridgeImage string) {
	e.memory.InsertGameCartridge(cartridge.LoadCartridgeImage(pathToCartridgeImage))
}

func (e *Core) Tick() (left byte, right byte, play bool) {
	e.cpu.Tick()
	e.timer.Tick()
	e.ppu.Tick()
	left, right, play = e.apu.Tick()
	e.memory.Tick()
	return
}

func (e *Core) SaveGame() {
	e.memory.GetGameCartridge().Save()
}

func (e *Core) KeyPressed(index byte) {
	go func() {
		e.joypad.KeyPressed(index)
	}()
}

func (e *Core) KeyReleased(index byte) {
	go func() {
		e.joypad.KeyReleased(index)
	}()
}
