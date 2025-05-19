package main

import (
	"flag"
	"gameboy-emulator/internal/cycle/apu"
	"gameboy-emulator/internal/cycle/cpu"
	"gameboy-emulator/internal/cycle/emulation"
	"gameboy-emulator/internal/cycle/gpu"
	"gameboy-emulator/internal/cycle/interrupts"
	"gameboy-emulator/internal/cycle/joypad"
	"gameboy-emulator/internal/cycle/memory"
	"gameboy-emulator/internal/cycle/timer"
	"github.com/ebitengine/oto/v3"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

func main() {
	// Get Application path
	appPath, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		panic(err)
	}

	biosPath := flag.String("bios", filepath.Join(appPath, "dmg_boot.bin"), "Path to the boot image")
	logConfigPath := flag.String("log", filepath.Join(appPath, "zap_config.yaml"), "Path to zap logging config")
	flag.Parse()

	// Setup Logger
	logger := createLogger(*logConfigPath)
	defer logger.Sync()

	undo := zap.ReplaceGlobals(logger)
	defer undo()

	// Load BIOS
	bios, err := os.ReadFile(*biosPath)
	if err != nil {
		panic(err)
	}

	// Wire dependencies
	a := apu.New()
	i := interrupts.New()
	j := joypad.New(i)
	t := timer.New(i)
	p := gpu.NewPPU(i)
	m := memory.New(i, t, p, j, a, (*[0x100]byte)(bios))
	c := cpu.New(m, i)

	emulatorCore := emulation.NewCore(i, j, t, p, m, c, a)
	defer emulatorCore.SaveGame()

	// Setup sound
	op := &oto.NewContextOptions{}
	op.SampleRate = apu.SamplingRate
	op.ChannelCount = 2
	op.Format = oto.FormatUnsignedInt8
	op.BufferSize = 4096

	ctx, ready, err := oto.NewContext(op)
	if err != nil {
		panic(err)
	}
	<-ready

	// Create Sound d
	driver := NewSoundDriver(ctx, emulatorCore)
	ui := NewUserInterface(driver)
	ui.ShowAndRun()
}

func createLogger(logConfigPath string) *zap.Logger {
	configFile, err := os.ReadFile(logConfigPath)
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
