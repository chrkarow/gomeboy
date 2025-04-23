package main

import (
	"gameboy-emulator/internal/cartridge"
	"gameboy-emulator/internal/cpu"
	"gameboy-emulator/internal/gpu"
	"gameboy-emulator/internal/interrupts"
	"gameboy-emulator/internal/memory"
	"gameboy-emulator/internal/timer"
	"github.com/stretchr/testify/assert/yaml"
	"go.uber.org/zap"
	"os"
)

//TIP <p>To run your code, right-click the code and select <b>Run</b>.</p> <p>Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.</p>

func main() {

	logger := createLogger()
	defer logger.Sync()

	undo := zap.ReplaceGlobals(logger)
	defer undo()

	bios, err := os.ReadFile("roms/dmg_boot.bin")
	if err != nil {
		panic(err)
	}

	inter := interrupts.New()
	tim := timer.New(inter)
	lcd := gpu.New(inter)
	mem := memory.New(inter, tim, lcd, (*[0x100]byte)(bios))
	processor := cpu.New(mem, inter)

	mem.InsertGameCartridge(cartridge.LoadCartridgeImage("roms/Tetris.gb"))

	ui := NewUserInterface(lcd)

	go func() {
		var lastUpdateCycles uint64
		var refreshCounter int

		for {
			currentCycles := processor.Step()
			tim.UpdateTimer(currentCycles)
			lcd.UpdateDisplay(currentCycles)
			inter.HandleInterrupt()

			refreshCounter += int(currentCycles - lastUpdateCycles)
			lastUpdateCycles = currentCycles

			if refreshCounter >= 69905 {
				refreshCounter %= 69905
				ui.UpdateFrame()
			}
		}
	}()

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
