package cartridge

import (
	"gameboy-emulator/internal/util"
	log "go.uber.org/zap"
	"os"
)

// MBC2 supports ROM sizes up to 2 Mbit (16 banks of 0x4000 bytes) and includes an internal
// 512x4 bit RAM array, which is its unique feature.
//
// Source: docs/gbctr.pdf page 142 ff
type mbc2 struct {
	rom            *[]byte
	ram            [0x200]byte
	currentROMBank byte
	ramEnabled     bool
	name           string
}

func newMBC2(rom *[]byte) Cartridge {
	return &mbc2{
		rom:            rom,
		currentROMBank: 1,
		ramEnabled:     false,
		name:           getCartridgeName(rom),
	}
}

func (mbc *mbc2) ReadROM(address uint16) byte {
	switch {
	case address <= 0x3FFF: // hardcoded ROM Bank 0
		return (*mbc.rom)[address]

	case address <= 0x7FFF: // other ROM Banks
		physicalAddress := (uint32(mbc.currentROMBank)<<14 | uint32(address&0x3FFF)) & uint32(len(*mbc.rom)-1)
		return (*mbc.rom)[physicalAddress]

	default:
		log.L().Panic("Invalid ROM read", log.Uint16("address", address))
	}
	return 0
}

func (mbc *mbc2) HandleBanking(address uint16, data byte) {
	if address >= 0x4000 {
		return
	}

	if util.BitIsSet16(address, 8) { // if least significant bit of upper address byte is one ...
		mbc.currentROMBank = data & 0x0F // ...lower 4 bits of written value encode the ROM romBank

		if mbc.currentROMBank == 0x00 { // if ROM romBank should ever be set to 0 it is treated as 1
			mbc.currentROMBank++
		}

	} else { // if least significant bit of upper address byte is zero, data controls RAM enabling
		mbc.ramEnabled = data&0x0F == 0x0A // RAM only enabled, if lower nibble of data all 1s.
	}
}

func (mbc *mbc2) WriteRAM(address uint16, data byte) {
	if address >= 0x2000 {
		log.L().Panic("Invalid RAM write attempt", log.Uint16("address", address))
	}

	// If RAM is not enabled writes are simply ignored
	if !mbc.ramEnabled {
		return
	}

	// MBC2 only has 0x200 byte of RAM, that's why only the lower 9 bits of the address are used
	mbc.ram[address&0x1FF] = data | 0xF0 // only lower for bits are stored and upper for bits are set to 1 (open bus)
}

func (mbc *mbc2) ReadRAM(address uint16) byte {
	if address >= 0x2000 {
		log.L().Panic("Invalid RAM read", log.Uint16("address", address))
	}

	// If RAM is not enabled reads return 0xFF
	if !mbc.ramEnabled {
		return 0xFF
	}

	// MBC2 only has 0x200 byte of RAM, that's why only the lower 9 bits of the address are used
	return mbc.ram[address&0x1FF]
}

func (mbc *mbc2) Save() {
	err := os.WriteFile(mbc.name+".sgo", mbc.ram[:], 0644)
	if err != nil {
		log.L().Error("Error writing save file", log.Error(err))
		return
	}
}

func (mbc *mbc2) load() {
	data, err := os.ReadFile(mbc.name + ".sgo")
	if err != nil {
		log.L().Error("Error reading save file", log.Error(err))
		return
	}
	mbc.ram = [512]byte(data)
}
