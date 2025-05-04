package main

import (
	"gameboy-emulator/internal/cycle"
	"gameboy-emulator/internal/cycle/apu"
	"gameboy-emulator/internal/cycle/cpu"
	"gameboy-emulator/internal/cycle/gpu"
	"gameboy-emulator/internal/cycle/interrupts"
	"gameboy-emulator/internal/cycle/joypad"
	"gameboy-emulator/internal/cycle/memory"
	"gameboy-emulator/internal/cycle/timer"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"os"
)

func main() {
	// Setup Logger
	logger := createLogger()
	defer logger.Sync()

	undo := zap.ReplaceGlobals(logger)
	defer undo()

	// Load BIOS
	bios, err := os.ReadFile("roms/dmg_boot.bin")
	if err != nil {
		panic(err)
	}

	// Wire dependencies
	apuMock := apu.New()
	inter := interrupts.New()
	joyp := joypad.New(inter)
	tim := timer.New(inter)
	lcd := gpu.NewPPU(inter)
	mem := memory.New(inter, tim, lcd, joyp, apuMock, (*[0x100]byte)(bios))
	processor := cpu.New(mem, inter)

	emulator := cycle.NewEmulator(inter, joyp, tim, lcd, mem, processor, apuMock)

	ui := NewUserInterface(emulator)
	ui.ShowAndRun()
}

func createLogger() *zap.Logger {
	configFile, err := os.ReadFile("configs/zap_config.yaml")
	if err != nil {
		panic(err)
	}

	config := zap.Config{}
	if err = yaml.Unmarshal(configFile, &config); err != nil {
		panic(err)
	}
	logger, err := config.Build()
	if err != nil {
		panic(err)
	}
	return logger
}
