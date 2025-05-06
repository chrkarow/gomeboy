package cycle

import (
	"gameboy-emulator/internal/cartridge"
	"gameboy-emulator/internal/cycle/apu"
	"gameboy-emulator/internal/cycle/cpu"
	"gameboy-emulator/internal/cycle/gpu"
	"gameboy-emulator/internal/cycle/interrupts"
	"gameboy-emulator/internal/cycle/joypad"
	"gameboy-emulator/internal/cycle/memory"
	"gameboy-emulator/internal/cycle/timer"
	"time"
)

type (
	Emulator struct {
		interrupts *interrupts.Interrupts
		joypad     *joypad.Joypad
		timer      *timer.Timer
		ppu        *gpu.PPU
		memory     *memory.Memory
		cpu        *cpu.CPU
		apu        *apu.APU

		paused  bool
		stopped bool
		turbo   bool
	}
)

func NewEmulator(
	interrupts *interrupts.Interrupts,
	joypad *joypad.Joypad,
	timer *timer.Timer,
	ppu *gpu.PPU,
	memory *memory.Memory,
	cpu *cpu.CPU,
	apu *apu.APU,
) *Emulator {
	return &Emulator{
		interrupts: interrupts,
		joypad:     joypad,
		timer:      timer,
		ppu:        ppu,
		memory:     memory,
		cpu:        cpu,
		apu:        apu,
	}
}

func (e *Emulator) Reset() {
	e.interrupts.Reset()
	e.joypad.Reset()
	e.timer.Reset()
	e.ppu.Reset()
	e.memory.Reset()
	e.cpu.Reset()
	e.apu.Reset()
}

func (e *Emulator) SetScreenHandler(handler func([144][160]byte)) {
	e.ppu.GetDisplay().RegisterFrameOutputHandler(handler)
}

func (e *Emulator) InsertCartridge(pathToCartridgeImage string) {
	e.memory.InsertGameCartridge(cartridge.LoadCartridgeImage(pathToCartridgeImage))
}

func (e *Emulator) Run() {
	go func() {

		// save cartridge RAM when emulator loop ends
		defer e.memory.GetGameCartridge().Save()

		var count int
		for !e.stopped {

			if e.paused {
				time.Sleep(time.Second)
				continue
			}

			if !e.turbo && count == 10000 {
				count = 0
				time.Sleep(1600 * time.Microsecond)
			}

			e.cpu.Tick()
			e.memory.Tick()
			e.timer.Tick()
			e.ppu.Tick()

			count++
		}

		e.Reset()
		e.stopped = false
		e.paused = false
	}()
}

func (e *Emulator) ToggleTurbo() {
	e.turbo = !e.turbo
}

func (e *Emulator) TogglePause() {
	e.paused = !e.paused
}

func (e *Emulator) IsPaused() bool {
	return e.paused
}

func (e *Emulator) Stop() {
	e.stopped = true
}

func (e *Emulator) KeyPressed(index byte) {
	go func() {
		e.joypad.KeyPressed(index)
	}()
}

func (e *Emulator) KeyReleased(index byte) {
	go func() {
		e.joypad.KeyReleased(index)
	}()
}
