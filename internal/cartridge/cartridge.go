package cartridge

import (
	"fmt"
	"gameboy-emulator/internal/util"
	log "go.uber.org/zap"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var ramSizes = [6]int{
	0,       // No RAM
	0,       // Unused
	0x2000,  // 1 bank of 8 KiB
	0x8000,  // 4 banks of 8 KiB = 32 KiB
	0x20000, // 16 banks of 8 KiB = 128 KiB
	0x10000, // 8 banks of 8 KiB = 64 KiB
}

type (
	Cartridge interface {

		// ReadROM returns the data stored under the given address in the cartridge ROM.
		//
		// Allowed values for address range from 0x0000 to 0x3FFF (both bounds inclusive)
		ReadROM(address uint16) byte

		// HandleBanking enables RAM banking and changing ROM and RAM banks.
		// This function is called when there is a write attempt to a ROM address (nice hack nintendo!)
		//
		// Allowed values for address range from 0x0000 to 0x3FFF (both bounds inclusive)
		//
		// Source: http://www.codeslinger.co.uk/pages/projects/gameboy/banking.html
		HandleBanking(address uint16, data byte)

		// ReadRAM returns the data stored under the given address in the cartridge RAM used for persistently
		// saving game data. Results in panic if RAM wasn't enabled before (see HandleBanking).
		//
		// Allowed values for address depend on the underlying cartridge type (which MBC is used).
		// RAM always starts at 0x0000. Don't simply use the register ranges to access.
		ReadRAM(address uint16) byte

		// WriteRAM stores the given data under the given address in the cartridge RAM used for persistently
		// saving game data. Results in panic if RAM wasn't enabled before (see HandleBanking).
		//
		// Allowed values for address depend on the underlying cartridge type (which MBC is used)
		// RAM always starts at 0x0000. Don't simply use the register ranges to access.
		WriteRAM(address uint16, data byte)

		// Save persists the RAM data to survive turning off the emulator
		Save()
	}

	cartridgeCore struct {
		rom          *[]byte
		ram          []byte
		imagePath    string
		saveGamePath string
	}
)

func LoadCartridgeImage(imagePath string) Cartridge {
	data, err := os.ReadFile(imagePath)

	if err != nil {
		panic(err)
	}

	core := &cartridgeCore{
		rom:          &data,
		ram:          make([]byte, ramSizes[data[0x149]]),
		imagePath:    imagePath,
		saveGamePath: strings.TrimSuffix(imagePath, filepath.Ext(imagePath)) + ".sgo",
	}

	return createCartridge(core)
}

func createCartridge(core *cartridgeCore) Cartridge {
	switch (*core.rom)[0x147] {
	case 0x00, 0x08, 0x09:
		return newNoMBC(core)
	case 0x01, 0x02, 0x03:
		return newMBC1(core)
	case 0x05, 0x06:
		return newMBC2(core)
	case 0x0F, 0x10, 0x11:
		return newMBC3(core, time.Now)
	case 0x12, 0x13:
		return newMBC3(core, time.Now)
	case 0x19, 0x1A, 0x1B:
		return newMBC5(core)
	case 0x1C, 0x1D, 0x1E:
		return newMBC5(core)
	default:
		log.L().Panic("Required MBC not implemented", log.String("value", fmt.Sprintf("0x%02X", (*core.rom)[0x147])))
	}
	return nil
}

func (c *cartridgeCore) Save() {
	// If RAM is completely empty (= all zeroes) don't save
	if util.IsEmpty(c.ram) {
		return
	}
	err := os.WriteFile(c.saveGamePath, c.ram, 0644)
	if err != nil {
		log.L().Error("Error writing save file", log.Error(err))
		return
	}
}

func (c *cartridgeCore) load() {
	data, err := os.ReadFile(c.saveGamePath)
	if err != nil {
		if !os.IsNotExist(err) {
			log.L().Error("Error reading save file", log.Error(err))
		}
		return
	}
	c.ram = data
}
