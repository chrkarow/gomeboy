package internal

import (
	"gameboy-emulator/internal/cartridge"
	"gameboy-emulator/internal/cpu"
	"gameboy-emulator/internal/gpu"
	"gameboy-emulator/internal/interrupts"
	"gameboy-emulator/internal/joypad"
	"gameboy-emulator/internal/memory"
	"gameboy-emulator/internal/timer"
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

		updateHandler FrameUpdateHandler
	}

	FrameUpdateHandler interface {
		UpdateFrame()
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

func (e *Emulator) RegisterFrameUpdateHandler(handler FrameUpdateHandler) {
	e.updateHandler = handler
}

func (e *Emulator) InsertCartridgeAndRun(pathToCartridgeImage string) {
	var refreshCounter int

	e.memory.InsertGameCartridge(cartridge.LoadCartridgeImage(pathToCartridgeImage))

	for {
		startTime := time.Now().UnixNano()

		stepCycles := e.cpu.Step()
		e.timer.UpdateTimer(stepCycles)
		e.gpu.UpdateDisplay(stepCycles)
		e.interrupts.HandleInterrupt()

		emulatedNanos := int64(stepCycles) * 238

		// Update UI circa 60 times per second
		refreshCounter += stepCycles
		if refreshCounter >= 69905 {
			refreshCounter %= 69905
			e.updateHandler.UpdateFrame()
		}

		// Ty to synchronize to real time
		realNanos := time.Now().UnixNano() - startTime
		if emulatedNanos > realNanos {
			time.Sleep(time.Duration(emulatedNanos - realNanos))
		}
	}
}

func (e *Emulator) KeyPressed(index byte) {
	e.joypad.KeyPressed(index)
}

func (e *Emulator) KeyReleased(index byte) {
	e.joypad.KeyReleased(index)
}

func (e *Emulator) GetScreen() [144][160]byte {
	return e.gpu.GetScreen()
}
