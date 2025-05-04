package main

import (
	"gameboy-emulator/internal/step"
	"gameboy-emulator/internal/step/cpu"
	"gameboy-emulator/internal/step/gpu"
	"gameboy-emulator/internal/step/interrupts"
	"gameboy-emulator/internal/step/joypad"
	"gameboy-emulator/internal/step/memory"
	"gameboy-emulator/internal/step/timer"
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
	inter := interrupts.New()
	joyp := joypad.New(inter)
	tim := timer.New(inter)
	lcd := gpu.New(inter)
	mem := memory.New(inter, tim, lcd, joyp, (*[0x100]byte)(bios))
	processor := cpu.New(mem, inter)

	emulator := step.NewEmulator(inter, joyp, tim, lcd, mem, processor)

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
