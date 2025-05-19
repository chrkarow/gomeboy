package cartridge

import (
	log "go.uber.org/zap"
)

type mbc5 struct {
	*cartridgeCore

	romb0      byte
	romb1      byte
	ramb       byte
	ramEnabled bool
}

func newMBC5(core *cartridgeCore) Cartridge {
	return &mbc5{
		cartridgeCore: core,
		romb0:         1,
	}
}

func (mbc *mbc5) ReadROM(address uint16) byte {

	switch {
	case address < 0x4000:
		return (*mbc.rom)[address&uint16(len(*mbc.rom)-1)]
	case address < 0x8000:
		bankAddress := uint32(uint16(mbc.romb1)<<8|uint16(mbc.romb0)) << 14
		physicalAddress := (bankAddress | uint32(address&0x3FFF)) & uint32(len(*mbc.rom)-1)
		return (*mbc.rom)[physicalAddress]
	default:
		log.L().Panic("Invalid ROM read", log.Uint16("address", address))
	}
	return 0xFF
}

func (mbc *mbc5) HandleBanking(address uint16, data byte) {
	switch {
	case address < 0x2000: // Enable/Disable RAM (all bits count)
		mbc.ramEnabled = data == 0x0A

	case address < 0x3000: // write to register BANK1
		mbc.romb0 = data

	case address < 0x4000: // write to register BANK2
		mbc.romb1 = data & 0x01 // only lowest  bit is relevant

	case address < 0x6000: // change mode (uses only last bit)
		mbc.ramb = data & 0x0F // only bit 0-3 are relevant
	}
}

func (mbc *mbc5) WriteRAM(address uint16, data byte) {
	if address >= 0x2000 {
		log.L().Panic("Invalid RAM write attempt", log.Uint16("address", address))
	}

	// If RAM is not enabled writes are simply ignored
	if !mbc.ramEnabled {
		return
	}

	physicalAddress := (uint16(mbc.ramb)<<13 | address&0x1FFF) & uint16(len(mbc.ram)-1)

	mbc.ram[physicalAddress] = data
}

func (mbc *mbc5) ReadRAM(address uint16) byte {
	if address >= 0x2000 {
		log.L().Panic("Invalid RAM read", log.Uint16("address", address))
	}

	// If RAM is not enabled reads return 0xFF
	if !mbc.ramEnabled {
		return 0xFF
	}

	physicalAddress := (uint16(mbc.ramb)<<13 | address&0x1FFF) & uint16(len(mbc.ram)-1)

	return mbc.ram[physicalAddress]
}
