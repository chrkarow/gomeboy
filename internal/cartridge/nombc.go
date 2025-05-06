package cartridge

import (
	log "go.uber.org/zap"
)

// Small games of not more than 32 KiB ROM do not require
// a MBC chip for ROM banking. The ROM is directly mapped
// to memory at $0000-7FFF. Optionally up to 8 KiB of RAM could
// be connected at $A000-BFFF, using a discrete logic decoder
// in place of a full MBC chip.
type noMBC struct {
	rom *[]byte
	ram [0x2000]byte
}

func newNoMBC(rom *[]byte) Cartridge {
	return &noMBC{rom: rom}
}

func (mbc *noMBC) ReadROM(address uint16) byte {
	if address >= 0x8000 {
		log.L().Panic("Invalid ROM read", log.Uint16("address", address))
	}
	return (*mbc.rom)[address&uint16(len(*mbc.rom)-1)]
}

func (mbc *noMBC) HandleBanking(_ uint16, _ byte) {
	// no banking
}

func (mbc *noMBC) WriteRAM(address uint16, data byte) {
	mbc.ram[address] = data
}

func (mbc *noMBC) ReadRAM(address uint16) byte {
	return mbc.ram[address]
}

func (mbc *noMBC) Save() {
	// Nothing to do here
}

func (mbc *noMBC) load() {
	// Nothing to do here
}
