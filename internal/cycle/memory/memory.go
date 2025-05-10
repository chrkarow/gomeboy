package memory

import (
	"fmt"
	"gameboy-emulator/internal/cartridge"
	"gameboy-emulator/internal/cycle/apu"
	"gameboy-emulator/internal/cycle/gpu"
	"gameboy-emulator/internal/cycle/interrupts"
	"gameboy-emulator/internal/cycle/joypad"
	"gameboy-emulator/internal/cycle/timer"
	log "go.uber.org/zap"
)

// Memory are considered all addressable storage locations within a GameBoy. In the first place, this
// is the internal storage and the storage of the currently inserted game cartridge.
//
// Memory addresses are divided into following ranges:
//
//	0000-3FFF 	16KB ROM Bank 00 (in cartridge, fixed at bank 00)
//	4000-7FFF 	16KB ROM Bank 01..NN (in cartridge, switchable bank number)
//	8000-9FFF 	8KB Video RAM (VRAM) (switchable bank 0-1 in CGB Mode)
//	A000-BFFF 	8KB External RAM (in cartridge, switchable bank, if any)
//	C000-CFFF 	4KB Work RAM Bank 0 (WRAM)
//	D000-DFFF 	4KB Work RAM Bank 1 (WRAM) (switchable bank 1-7 in CGB Mode)
//	E000-FDFF 	Same as C000-DDFF (ECHO) (typically not used)
//	FE00-FE9F 	Sprite Attribute Table (OAM)
//	FEA0-FEFF 	Not Usable
//	FF00-FF7F 	I/O Ports
//	FF80-FFFE 	High RAM (HRAM)
//	FFFF 		Interrupt Enable Register
//
// Source: https://gbdev.io/pandocs/Memory_Map.html
type (
	Memory struct {
		wram [0x2000]byte
		hram [0x80]byte

		io [0xA0]ioRegister

		// Dummy registers for Data Link Cable
		sb byte
		sc byte

		// Cache for last written byte to DMA transfer register to be able to return it when read
		// Satisfies Gekkio Test reg_read
		dmaSourceAddress          uint16
		dmaRequestedSourceAddress uint16
		dmaTransferRequested      bool
		dmaTransferInProgress     bool
		dmaTransferCount          int
		ticks                     int
		pendingWrite              func(m *Memory)

		bootFlag byte // Set to non-zero to disable boot ROM

		interrupts *interrupts.Interrupts
		ppu        *gpu.PPU
		cartridge  cartridge.Cartridge

		bootRom [0x100]byte
	}

	ioRegister struct {
		name  string
		write func(data byte)
		read  func() byte
	}
)

func New(
	interrupts *interrupts.Interrupts,
	timer *timer.Timer,
	ppu *gpu.PPU,
	joypad *joypad.Joypad,
	apu *apu.APU,
	bootRom *[0x100]byte,
) *Memory {
	m := &Memory{
		interrupts: interrupts,
		ppu:        ppu,
		bootRom:    *bootRom,
	}
	m.initializeIOAddressSpace(
		timer,
		ppu,
		interrupts,
		joypad,
		apu,
	)
	m.Reset()
	return m
}

// Reset the memory to initial state.
//
// Values taken from https://github.com/Gekkio/mooneye-test-suite/blob/main/acceptance/boot_hwio-dmgABCmgb.s
func (mem *Memory) Reset() {
	mem.sc = 0x7E
	mem.sb = 0x0
	mem.wram = [0x2000]byte{}
	mem.hram = [0x80]byte{}
	mem.bootFlag = 0x0
	mem.dmaRequestedSourceAddress = 0x0
	mem.dmaTransferRequested = false
	mem.dmaSourceAddress = 0x0
	mem.dmaTransferInProgress = false
	mem.dmaTransferCount = 0
	mem.ticks = 0
	mem.pendingWrite = nil
}

func (mem *Memory) InsertGameCartridge(cart cartridge.Cartridge) {
	mem.cartridge = cart
}

func (mem *Memory) GetGameCartridge() cartridge.Cartridge {
	return mem.cartridge
}

func (mem *Memory) Tick() {
	// If there is a write access pending, execute it after this method
	if mem.pendingWrite != nil {
		defer func() {
			mem.pendingWrite(mem)
			mem.pendingWrite = nil
		}()
	}

	if !mem.dmaTransferInProgress && !mem.dmaTransferRequested {
		return
	}

	if mem.ticks++; mem.ticks < 4 {
		return
	}
	mem.ticks = 0

	if mem.dmaTransferRequested {
		mem.dmaTransferRequested = false
		mem.dmaSourceAddress = mem.dmaRequestedSourceAddress
		mem.dmaTransferCount = 0
		mem.dmaTransferInProgress = true
	}

	if mem.dmaTransferInProgress {
		mem.doDMATransfer()
	}
}

func (mem *Memory) Write(address uint16, data byte) {

	// during DMA transfer OAM and source memory area are blocked
	if mem.dmaTransferInProgress &&
		((address >= 0xFE00 && address < 0xFEA0) ||
			(address >= mem.dmaSourceAddress && address < mem.dmaSourceAddress+uint16(gpu.OAMSize))) {
		return
	}

	// due to timing issues, we can't allow a write request to take effect
	// as soon as it is done. Only when the memory is live (during Tick()) the write access
	// may be executed
	mem.pendingWrite = func(m *Memory) { m.internalWrite(address, data) }
}

func (mem *Memory) Read(address uint16) byte {
	if mem.dmaTransferInProgress &&
		((address >= 0xFE00 && address < 0xFEA0) ||
			(address >= mem.dmaSourceAddress && address < mem.dmaSourceAddress+uint16(gpu.OAMSize))) {
		return 0xFF
	}

	return mem.internalRead(address)
}

func (mem *Memory) internalWrite(address uint16, data byte) {
	switch {
	case address < 0x8000: // Write to actually read only memory changes banking within cartridge
		if mem.cartridgePresent() {
			mem.cartridge.HandleBanking(address, data)
		}

	case address < 0xA000: // VRAM
		mem.ppu.WriteVRam(address-0x8000, data)

	case address < 0xC000: // External RAM
		if mem.cartridgePresent() {
			mem.cartridge.WriteRAM(address-0xA000, data)
		}

	case address < 0xE000: // WRAM
		mem.wram[address-0xc000] = data

	case address < 0xFE00: // Write to so-called ECHO ram is the same as writing to WRAM (0xc000-0xddff)
		mem.internalWrite(address-0x2000, data)

	case address < 0xFEA0: // OAM
		mem.ppu.WriteOAM(address-0xFE00, data)

	case address < 0xFF00: // not usable area
		log.L().Debug("Write attempt to not usable memory area")

	case address < 0xFF80: // I/O Ports
		mem.handleIOWrite(address, data)

	case address < 0xFFFF: // High RAM
		mem.hram[address-0xFF80] = data

	case address == 0xFFFF:
		mem.interrupts.SetEnable(data)
	}
}

func (mem *Memory) internalRead(address uint16) byte {
	switch {
	case address < 0x8000: // Game cartridge data

		// While the bootROM is mapped "overlay" cartridge data with bootRom
		if mem.bootRomMapped() && address < 0x100 {
			return mem.bootRom[address]
		}

		// if game cartridge is inserted, read from game cartridge otherwise return 0xFF
		// Source: https://gbdev.io/pandocs/Power_Up_Sequence.html#monochrome-models-dmg0-dmg-mgb
		if mem.cartridgePresent() {
			return mem.cartridge.ReadROM(address)
		} else {
			return 0xFF
		}

	case address < 0xA000: // VRAM
		return mem.ppu.ReadVRam(address - 0x8000)

	case address < 0xC000: // External RAM
		// if game cartridge is inserted, read from game cartridge otherwise return 0xFF
		// Source: https://gbdev.io/pandocs/Power_Up_Sequence.html#monochrome-models-dmg0-dmg-mgb
		if mem.cartridgePresent() {
			return mem.cartridge.ReadRAM(address - 0xA000)
		} else {
			return 0xFF
		}

	case address < 0xE000: // WRAM
		return mem.wram[address-0xc000]

	case address < 0xFE00: // Read from so-called ECHO ram is the same as reading from WRAM (0xc000-0xddff)
		return mem.internalRead(address - 0x2000)

	case address < 0xFEA0: // OAM
		return mem.ppu.ReadOAM(address - 0xFE00)

	case address < 0xFF00: // not usable area
		log.L().Debug("Reading from not usable memory area")
		return 0xFF // unused = 1

	case address < 0xFF80: // I/O Ports
		return mem.handleIORead(address)

	case address < 0xFFFF: // High RAM
		return mem.hram[address-0xFF80]

	case address == 0xFFFF: // Interrupts enable flag register
		return mem.interrupts.GetEnable()
	}

	return 0x00
}

func (mem *Memory) handleIOWrite(address uint16, data byte) {
	ioLogger := log.L().With(log.String("address", fmt.Sprintf("0x%04X", address)))

	register := mem.io[address-0xFF00]

	if register.name == "" {
		ioLogger.Debug("Write attempt to unmapped I/O address")
		return
	}

	ioLogger.Debug(fmt.Sprintf("Writing to I/O - %s", register.name))
	register.write(data)
}

func (mem *Memory) handleIORead(address uint16) byte {
	ioLogger := log.L().With(log.String("address", fmt.Sprintf("0x%04X", address)))

	register := mem.io[address-0xFF00]

	if register.name == "" {
		ioLogger.Debug("Read attempt from unmapped I/O address")
		return 0xFF
	}

	ioLogger.Debug(fmt.Sprintf("Reading from to I/O - %s", register.name))
	return register.read()
}

func (mem *Memory) cartridgePresent() bool {
	return mem.cartridge != nil
}

func (mem *Memory) bootRomMapped() bool {
	return mem.bootFlag == 0x00
}

// requestDMATransfer executes a DMA transfer from ROM or RAM to OAM.
// The given value specifies the transfer source address divided by 0x100.
func (mem *Memory) requestDMATransfer(value byte) {
	mem.dmaRequestedSourceAddress = uint16(value) << 8 // This is the same as multiplying by 0x100
	mem.dmaTransferRequested = true
}

func (mem *Memory) getDMAData() byte {
	log.L().Warn("Read DMA transfer address 0xFF46")
	return byte(mem.dmaRequestedSourceAddress >> 8)
}

func (mem *Memory) doDMATransfer() {
	if mem.dmaTransferCount == gpu.OAMSize {
		mem.dmaTransferInProgress = false
		return
	}

	mem.internalWrite(0xFE00+uint16(mem.dmaTransferCount), mem.internalRead(mem.dmaSourceAddress+uint16(mem.dmaTransferCount)))
	mem.dmaTransferCount++
}

func (mem *Memory) setBootFlag(data byte) {
	mem.bootFlag = data
}

func (mem *Memory) readSc() byte {
	return mem.sc
}

func (mem *Memory) writeSc(data byte) {
	if data == 0x81 {
		fmt.Print(string(mem.sb))
	}
	mem.sc = data | 0x7E
}

func (mem *Memory) readSb() byte {
	return mem.sb
}

func (mem *Memory) writeSb(data byte) {
	mem.sb = data
}

func (mem *Memory) initializeIOAddressSpace(
	timer *timer.Timer,
	ppu *gpu.PPU,
	interrupts *interrupts.Interrupts,
	joypad *joypad.Joypad,
	apu *apu.APU,
) {
	// Joypad
	mem.io[0x00] = ioRegister{"JOYP", joypad.WriteRegister, joypad.ReadRegister}

	// Serial Data Transfer
	mem.io[0x01] = ioRegister{"SB", mem.writeSb, mem.readSb}
	mem.io[0x02] = ioRegister{"SC", mem.writeSc, mem.readSc}

	// Timer and Divider
	mem.io[0x04] = ioRegister{"DIV", func(_ byte) { timer.ResetDiv() }, timer.GetDiv}
	mem.io[0x05] = ioRegister{"TIMA", timer.SetTima, timer.GetTima}
	mem.io[0x06] = ioRegister{"TMA", timer.SetTma, timer.GetTma}
	mem.io[0x07] = ioRegister{"TAC", timer.SetTac, timer.GetTac}

	// Interrupts
	mem.io[0x0F] = ioRegister{"IF", interrupts.SetFlags, interrupts.GetFlags}

	// Audio
	mem.io[0x10] = ioRegister{"NR10", apu.WriteNR10, apu.ReadNR10}
	mem.io[0x11] = ioRegister{"NR11", apu.WriteNR11, apu.ReadNR11}
	mem.io[0x12] = ioRegister{"NR12", apu.WriteNR12, apu.ReadNR12}
	mem.io[0x13] = ioRegister{"NR13", apu.WriteNR13, apu.ReadNR13}
	mem.io[0x14] = ioRegister{"NR14", apu.WriteNR14, apu.ReadNR14}
	mem.io[0x16] = ioRegister{"NR21", apu.WriteNR21, apu.ReadNR21}
	mem.io[0x17] = ioRegister{"NR22", apu.WriteNR22, apu.ReadNR22}
	mem.io[0x18] = ioRegister{"NR23", apu.WriteNR23, apu.ReadNR23}
	mem.io[0x19] = ioRegister{"NR24", apu.WriteNR24, apu.ReadNR24}
	mem.io[0x1A] = ioRegister{"NR30", apu.WriteNR30, apu.ReadNR30}
	mem.io[0x1B] = ioRegister{"NR31", apu.WriteNR31, apu.ReadNR31}
	mem.io[0x1C] = ioRegister{"NR32", apu.WriteNR32, apu.ReadNR32}
	mem.io[0x1D] = ioRegister{"NR33", apu.WriteNR33, apu.ReadNR33}
	mem.io[0x1E] = ioRegister{"NR34", apu.WriteNR34, apu.ReadNR34}
	mem.io[0x20] = ioRegister{"NR41", apu.WriteNR41, apu.ReadNR41}
	mem.io[0x21] = ioRegister{"NR42", apu.WriteNR42, apu.ReadNR42}
	mem.io[0x22] = ioRegister{"NR43", apu.WriteNR43, apu.ReadNR43}
	mem.io[0x23] = ioRegister{"NR44", apu.WriteNR44, apu.ReadNR44}
	mem.io[0x24] = ioRegister{"NR50", apu.WriteNR50, apu.ReadNR50}
	mem.io[0x25] = ioRegister{"NR51", apu.WriteNR51, apu.ReadNR51}
	mem.io[0x26] = ioRegister{"NR52", apu.WriteNR52, apu.ReadNR52}

	// Wave RAM - think about nicer mapping to APU :)
	mem.io[0x30] = ioRegister{"Wave RAM (0x0)", func(data byte) { apu.WriteWaveRAM(0x0, data) }, func() byte { return apu.ReadWaveRAM(0x0) }}
	mem.io[0x31] = ioRegister{"Wave RAM (0x1)", func(data byte) { apu.WriteWaveRAM(0x1, data) }, func() byte { return apu.ReadWaveRAM(0x1) }}
	mem.io[0x32] = ioRegister{"Wave RAM (0x2)", func(data byte) { apu.WriteWaveRAM(0x2, data) }, func() byte { return apu.ReadWaveRAM(0x2) }}
	mem.io[0x33] = ioRegister{"Wave RAM (0x3)", func(data byte) { apu.WriteWaveRAM(0x3, data) }, func() byte { return apu.ReadWaveRAM(0x3) }}
	mem.io[0x34] = ioRegister{"Wave RAM (0x4)", func(data byte) { apu.WriteWaveRAM(0x4, data) }, func() byte { return apu.ReadWaveRAM(0x4) }}
	mem.io[0x35] = ioRegister{"Wave RAM (0x5)", func(data byte) { apu.WriteWaveRAM(0x5, data) }, func() byte { return apu.ReadWaveRAM(0x5) }}
	mem.io[0x36] = ioRegister{"Wave RAM (0x6)", func(data byte) { apu.WriteWaveRAM(0x6, data) }, func() byte { return apu.ReadWaveRAM(0x6) }}
	mem.io[0x37] = ioRegister{"Wave RAM (0x7)", func(data byte) { apu.WriteWaveRAM(0x7, data) }, func() byte { return apu.ReadWaveRAM(0x7) }}
	mem.io[0x38] = ioRegister{"Wave RAM (0x8)", func(data byte) { apu.WriteWaveRAM(0x8, data) }, func() byte { return apu.ReadWaveRAM(0x8) }}
	mem.io[0x39] = ioRegister{"Wave RAM (0x9)", func(data byte) { apu.WriteWaveRAM(0x9, data) }, func() byte { return apu.ReadWaveRAM(0x9) }}
	mem.io[0x3A] = ioRegister{"Wave RAM (0xA)", func(data byte) { apu.WriteWaveRAM(0xA, data) }, func() byte { return apu.ReadWaveRAM(0xA) }}
	mem.io[0x3B] = ioRegister{"Wave RAM (0xB)", func(data byte) { apu.WriteWaveRAM(0xB, data) }, func() byte { return apu.ReadWaveRAM(0xB) }}
	mem.io[0x3C] = ioRegister{"Wave RAM (0xC)", func(data byte) { apu.WriteWaveRAM(0xC, data) }, func() byte { return apu.ReadWaveRAM(0xC) }}
	mem.io[0x3D] = ioRegister{"Wave RAM (0xD)", func(data byte) { apu.WriteWaveRAM(0xD, data) }, func() byte { return apu.ReadWaveRAM(0xD) }}
	mem.io[0x3E] = ioRegister{"Wave RAM (0xE)", func(data byte) { apu.WriteWaveRAM(0xE, data) }, func() byte { return apu.ReadWaveRAM(0xE) }}
	mem.io[0x3F] = ioRegister{"Wave RAM (0xF)", func(data byte) { apu.WriteWaveRAM(0xF, data) }, func() byte { return apu.ReadWaveRAM(0xF) }}

	// LCD Control, Status, Position, Scrolling and Palettes
	mem.io[0x40] = ioRegister{"LCDC", ppu.SetControl, ppu.GetControl}
	mem.io[0x41] = ioRegister{"STAT", ppu.SetStatus, ppu.GetStatus}
	mem.io[0x42] = ioRegister{"SCY", ppu.SetScrollY, ppu.GetScrollY}
	mem.io[0x43] = ioRegister{"SCX", ppu.SetScrollX, ppu.GetScrollX}
	mem.io[0x44] = ioRegister{"LY", func(_ byte) { /* ignore write */ }, ppu.GetCurrentLine}
	mem.io[0x45] = ioRegister{"LYC", ppu.SetCurrentLineCompare, ppu.GetCurrentLineCompare}
	mem.io[0x46] = ioRegister{"DMA", mem.requestDMATransfer, mem.getDMAData}
	mem.io[0x47] = ioRegister{"BGP", ppu.SetBackgroundPalette, ppu.GetBackgroundPalette}
	mem.io[0x48] = ioRegister{"OBP0", ppu.SetObjectPalette0, ppu.GetObjectPalette0}
	mem.io[0x49] = ioRegister{"OBP1", ppu.SetObjectPalette1, ppu.GetObjectPalette1}
	mem.io[0x4A] = ioRegister{"WY", ppu.SetWindowY, ppu.GetWindowY}
	mem.io[0x4B] = ioRegister{"WX", ppu.SetWindowX, ppu.GetWindowX}

	// Boot flag control
	mem.io[0x50] = ioRegister{"BOOT", mem.setBootFlag, func() byte { return 0xFF }}
}
