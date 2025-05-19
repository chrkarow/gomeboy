package cartridge

import (
	log "go.uber.org/zap"
)

// This is the first MBC chip for the Game Boy. Any newer MBC chips work similarly,
// so it is relatively easy to upgrade a program from one MBC chip to another —
// or to make it compatible with several types of MBCs.
//
// In its default configuration, MBC1 supports up to 512 KiB ROM with up to 32 KiB of banked RAM.
// Some cartridges wire the MBC differently, where the 2-bit RAM banking register is wired as an
// extension of the ROM banking register (instead of to RAM) in order to support up to 2 MiB ROM,
// at the cost of only supporting a fixed 8 KiB of cartridge RAM. All MBC1 cartridges with 1 MiB
// of ROM or more use this alternate wiring. Also see the note on MBC1M multi-game compilation carts
// below.
//
// Note that the memory in range 0000–7FFF is used both for reading from ROM and writing to the
// MBCs Control Registers.
//
// Source: docs/gbctr.pdf page 136 ff
type mbc1 struct {
	*cartridgeCore

	bank1          byte
	bank2          byte
	currentRAMBank byte
	mode           byte
	ramEnabled     bool
}

func newMBC1(core *cartridgeCore) Cartridge {
	return &mbc1{
		cartridgeCore: core,
		bank1:         1,
	}
}

func (mbc *mbc1) ReadROM(address uint16) byte {

	if address > 0x7FFF {
		log.L().Panic("Invalid ROM read", log.Uint16("address", address))
	}

	physicalAddress := (mbc.getROMBank(address) | uint32(address&0x3FFF)) & uint32(len(*mbc.rom)-1)
	return (*mbc.rom)[physicalAddress]
}

func (mbc *mbc1) HandleBanking(address uint16, data byte) {
	switch {
	case address < 0x2000: // Enable/Disable RAM
		wasEnabled := mbc.ramEnabled
		mbc.ramEnabled = data&0x0F == 0x0A
		if wasEnabled && !mbc.ramEnabled {
			mbc.Save()
		} else if !wasEnabled && mbc.ramEnabled {
			mbc.load()
		}

	case address < 0x4000: // write to register BANK1
		mbc.bank1 = data & 0x1F // masking with 0x31 (aka. 0b00011111) to get the lower 5 bits

		// lower 5 bits must never be all zeroes
		if mbc.bank1 == 0 {
			mbc.bank1++
		}

	case address < 0x6000: // write to register BANK2
		mbc.bank2 = data & 0x03 // only lower 2 bits are relevant

	case address < 0x8000: // change mode (uses only last bit)
		mbc.mode = data & 0x01
	}
}

func (mbc *mbc1) WriteRAM(address uint16, data byte) {
	if address >= 0x2000 {
		log.L().Panic("Invalid RAM write attempt", log.Uint16("address", address))
	}

	// If RAM is not enabled writes are simply ignored
	if !mbc.ramEnabled {
		return
	}

	physicalAddress := address & 0x1FFF
	if mbc.mode == 1 { // in MBC mode 1 the value of register BANK2 represents the RAM Bank
		physicalAddress |= uint16(mbc.bank2) << 13
	}
	physicalAddress &= uint16(len(mbc.ram)) - 1

	mbc.ram[physicalAddress] = data
}

func (mbc *mbc1) ReadRAM(address uint16) byte {
	if address >= 0x2000 {
		log.L().Panic("Invalid RAM read", log.Uint16("address", address))
	}

	// If RAM is not enabled reads return 0xFF
	if !mbc.ramEnabled {
		return 0xFF
	}

	physicalAddress := address & 0x1FFF
	if mbc.mode == 1 { // in MBC mode 1 the value of register BANK2 represents the RAM Bank
		physicalAddress |= uint16(mbc.bank2) << 13
	}
	physicalAddress &= uint16(len(mbc.ram)) - 1

	return mbc.ram[physicalAddress]
}

func (mbc *mbc1) getROMBank(address uint16) uint32 {
	switch {
	case mbc.mode == 0 && address <= 0x3FFF:
		return 0
	case mbc.mode == 1 && address <= 0x3FFF:
		return uint32(mbc.bank2) << 19
	case address <= 0x7FFF:
		return uint32(mbc.bank2)<<19 | uint32(mbc.bank1)<<14
	}
	return 0
}
