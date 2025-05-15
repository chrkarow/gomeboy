package cartridge

import (
	"gameboy-emulator/internal/util"
	log "go.uber.org/zap"
	"os"
	"time"
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
type (
	nowProvider func() time.Time

	mbc3 struct {
		rom           *[]byte
		ram           []byte
		romb          byte
		rambRtc       byte
		rtcLatchHigh  bool
		ramRtcEnabled bool

		rtcS  byte
		rtcM  byte
		rtcH  byte
		rtcDL byte
		rtcDH byte

		shadowRtcS  byte
		shadowRtcM  byte
		shadowRtcH  byte
		shadowRtcDL byte
		shadowRtcDH byte

		lastUpdate time.Time
		getNow     nowProvider
		ticker     *time.Ticker
		stopTicker chan bool
		name       string
	}
)

func newMBC3(rom *[]byte, ramSize int, getNow nowProvider) Cartridge {
	mbc := &mbc3{
		rom:        rom,
		ram:        make([]byte, ramSize),
		romb:       1,
		lastUpdate: getNow().UTC(),
		getNow:     getNow,
		stopTicker: make(chan bool),
		name:       getCartridgeName(rom),
	}
	mbc.startRTC()
	return mbc
}

func (mbc *mbc3) ReadROM(address uint16) byte {
	if address >= 0x8000 {
		log.L().Panic("Invalid ROM read", log.Uint16("address", address))
	}
	switch {
	case address < 0x4000:
		return (*mbc.rom)[address&uint16(len(*mbc.rom)-1)]
	case address < 0x8000:
		bankOffset := uint32(mbc.romb) << 14
		physicalAddress := bankOffset | uint32(address&0x3FFF)
		return (*mbc.rom)[physicalAddress]
	default:
		log.L().Panic("Invalid ROM read", log.Uint16("address", address))
	}

	return 0xFF
}

func (mbc *mbc3) HandleBanking(address uint16, data byte) {
	switch {
	case address < 0x2000: // Enable/Disable RAM
		mbc.ramRtcEnabled = data&0x0F == 0x0A

	case address < 0x4000: // ROM bank
		mbc.romb = data

		// lower 7 bits must never be all zeroes
		if mbc.romb == 0 {
			mbc.romb++
		}

	case address < 0x6000: // RAM bank or RTC Register
		mbc.rambRtc = data

	case address < 0x8000: // Latch clock data
		switch data {
		case 0x00:
			mbc.rtcLatchHigh = false
		case 0x01:
			if !mbc.rtcLatchHigh {
				mbc.latch()
			}
			mbc.rtcLatchHigh = true
		}
	}
}

func (mbc *mbc3) WriteRAM(address uint16, data byte) {
	if address >= 0x2000 {
		log.L().Panic("Invalid RAM write attempt", log.Uint16("address", address))
	}

	// If RAM is not enabled writes are simply ignored
	if !mbc.ramRtcEnabled {
		return
	}

	if mbc.rambRtc <= 0x07 {
		physicalAddress := address & 0x1FFF
		physicalAddress |= uint16(mbc.rambRtc) << 13
		physicalAddress &= uint16(len(mbc.ram)) - 1
		mbc.ram[physicalAddress] = data
	}

	switch mbc.rambRtc {
	case 0x08:
		mbc.rtcS = data & 0x3F
		mbc.shadowRtcS = data & 0x3F
	case 0x09:
		mbc.rtcM = data & 0x3F
		mbc.shadowRtcM = data & 0x3F
	case 0x0A:
		mbc.rtcH = data & 0x1F
		mbc.shadowRtcH = data & 0x1F
	case 0x0B:
		mbc.rtcDL = data
		mbc.shadowRtcDL = data
	case 0x0C:
		mbc.rtcDH = data & 0xC1
		mbc.shadowRtcDH = data & 0xC1
		if util.BitIsSet8(data, 6) {
			mbc.stopRTC()
		} else {
			mbc.startRTC()
		}
	}
}

func (mbc *mbc3) ReadRAM(address uint16) byte {
	if address >= 0x2000 {
		log.L().Panic("Invalid RAM read", log.Uint16("address", address))
	}

	// If RAM is not enabled reads return 0xFF
	if !mbc.ramRtcEnabled {
		return 0xFF
	}

	if mbc.rambRtc <= 0x07 {
		physicalAddress := address & 0x1FFF
		physicalAddress |= uint16(mbc.rambRtc) << 13
		physicalAddress &= uint16(len(mbc.ram)) - 1

		return mbc.ram[physicalAddress]
	}

	switch mbc.rambRtc {
	case 0x08:
		return mbc.rtcS
	case 0x09:
		return mbc.rtcM
	case 0x0A:
		return mbc.rtcH
	case 0x0B:
		return mbc.rtcDL
	case 0x0C:
		return mbc.rtcDH
	}

	return 0xFF
}

func (mbc *mbc3) latch() {
	mbc.rtcS = mbc.shadowRtcS
	mbc.rtcM = mbc.shadowRtcM
	mbc.rtcH = mbc.shadowRtcH
	mbc.rtcDL = mbc.shadowRtcDL
	mbc.rtcDH = mbc.shadowRtcDH
}

func (mbc *mbc3) startRTC() {
	if mbc.ticker != nil {
		return
	}

	mbc.ticker = time.NewTicker(time.Second)
	go func() {
		for {
			select {
			case <-mbc.ticker.C:
				mbc.tick()
			case <-mbc.stopTicker:
				return
			}
		}
	}()
}

func (mbc *mbc3) stopRTC() {
	if mbc.ticker == nil {
		return
	}

	mbc.ticker.Stop()
	mbc.stopTicker <- true
	mbc.ticker = nil
}

func (mbc *mbc3) tick() {
	mbc.lastUpdate = mbc.getNow()
	overflow := false

	mbc.shadowRtcS++
	switch mbc.shadowRtcS {
	case 0x3C:
		overflow = true
		mbc.shadowRtcS = 0
	case 0x40:
		mbc.shadowRtcS = 0
	}

	if !overflow {
		return
	}
	overflow = false

	mbc.shadowRtcM++
	switch mbc.shadowRtcM {
	case 0x3C:
		overflow = true
		mbc.shadowRtcM = 0
	case 0x40:
		mbc.shadowRtcM = 0
	}

	if !overflow {
		return
	}
	overflow = false

	mbc.shadowRtcH++
	switch mbc.shadowRtcH {
	case 0x18:
		overflow = true
		mbc.shadowRtcH = 0
	case 0x20:
		mbc.shadowRtcH = 0
	}

	if !overflow {
		return
	}
	overflow = false

	mbc.shadowRtcDL++
	if mbc.shadowRtcDL == 0x0 {
		overflow = true
	}

	if !overflow {
		return
	}

	if mbc.shadowRtcDH&0x01 == 0x00 {
		mbc.shadowRtcDH++
	} else {
		mbc.shadowRtcDH--
		util.SetBit(&mbc.shadowRtcDH, 7)
	}
}

func (mbc *mbc3) Save() {
	saveState := mbc.ram[:]
	saveState = append(saveState, mbc.shadowRtcS, mbc.shadowRtcM, mbc.shadowRtcH, mbc.shadowRtcDL, mbc.shadowRtcDH)
	saveState = append(saveState, []byte(mbc.lastUpdate.Format(time.DateTime))...)

	err := os.WriteFile(mbc.name+".sgo", mbc.ram, 0644)
	if err != nil {
		log.L().Error("Error writing save file", log.Error(err))
	}
}

func (mbc *mbc3) load() {
	data, err := os.ReadFile(mbc.name + ".sgo")
	if err != nil {
		if !os.IsNotExist(err) {
			log.L().Error("Error reading save file", log.Error(err))
		}
		return
	}

	ramSize := len(mbc.ram)
	mbc.ram = data[:ramSize]
	mbc.shadowRtcS = data[ramSize]
	mbc.shadowRtcM = data[ramSize+1]
	mbc.shadowRtcH = data[ramSize+2]
	mbc.shadowRtcDL = data[ramSize+3]
	mbc.shadowRtcDH = data[ramSize+4]
	mbc.lastUpdate, _ = time.Parse(time.DateTime, string(data[ramSize+5:]))
}
