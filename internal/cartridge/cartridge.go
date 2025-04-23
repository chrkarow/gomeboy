package cartridge

import (
	"fmt"
	log "go.uber.org/zap"
	"os"
)

var ramSizes = [6]int{
	0,       // No RAM
	0,       // Unused
	0x2000,  // 1 bank of 8 KiB
	0x8000,  // 4 banks of 8 KiB = 32 KiB
	0x20000, // 16 banks of 8 KiB = 128 KiB
	0x10000, // 8 banks of 8 KiB = 64 KiB
}

type Cartridge interface {

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
}

func LoadCartridgeImage(fileName string) Cartridge {
	data, err := os.ReadFile(fileName)

	if err != nil {
		panic(err)
	}

	return createCartridge(&data)
}

func createCartridge(data *[]byte) Cartridge {

	ramSize := ramSizes[(*data)[0x149]]

	switch (*data)[0x147] {
	case 0x00, 0x08, 0x09:
		return newNoMBC(data)
	case 0x01, 0x02, 0x03:
		return newMBC1(data, ramSize)
	case 0x05, 0x06:
		return newMBC2(data)
	default:
		log.L().Panic("Required MBC not implemented", log.String("value", fmt.Sprintf("0x%02X", (*data)[0x147])))
	}
	return nil
}
