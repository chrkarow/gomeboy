package step

import (
	"gameboy-emulator/internal/cartridge"
	"gameboy-emulator/internal/step/cpu"
	"gameboy-emulator/internal/step/gpu"
	"gameboy-emulator/internal/step/interrupts"
	"gameboy-emulator/internal/step/joypad"
	"gameboy-emulator/internal/step/memory"
	"gameboy-emulator/internal/step/timer"
	"time"
)

type (
	Emulator struct {
		interrupts *interrupts.Interrupts
		joypad     *joypad.Joypad
		timer      *timer.Timer
		gpu        *gpu.GPU
		memory     *memory.Memory
		cpu        *cpu.CPU

		paused  bool
		stopped bool
	}
)

func NewEmulator(
	interrupts *interrupts.Interrupts,
	joypad *joypad.Joypad,
	timer *timer.Timer,
	gpu *gpu.GPU,
	memory *memory.Memory,
	cpu *cpu.CPU,
) *Emulator {
	return &Emulator{
		interrupts: interrupts,
		joypad:     joypad,
		timer:      timer,
		gpu:        gpu,
		memory:     memory,
		cpu:        cpu,
	}
}

func (e *Emulator) Reset() {
	e.interrupts.Reset()
	e.joypad.Reset()
	e.timer.Reset()
	e.gpu.Reset()
	e.memory.Reset()
	e.cpu.Reset()
}

func (e *Emulator) SetScreenHandler(handler func([144][160]byte)) {
	e.gpu.SetScreenHandler(handler)
}

func (e *Emulator) InsertCartridge(pathToCartridgeImage string) {
	e.memory.InsertGameCartridge(cartridge.LoadCartridgeImage(pathToCartridgeImage))
}

func (e *Emulator) Run() {
	go func() {

		// save cartridge RAM when emulator loop ends
		defer e.memory.GetGameCartridge().Save()

		for !e.stopped {

			if e.paused {
				time.Sleep(time.Second)
				continue
			}

			stepCycles := e.cpu.Step()
			e.timer.UpdateTimer(stepCycles)
			e.gpu.UpdateDisplay(stepCycles)
			e.interrupts.HandleInterrupt()
			time.Sleep(time.Duration(1000))
		}

		e.Reset()
		e.stopped = false
		e.paused = false
	}()
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
