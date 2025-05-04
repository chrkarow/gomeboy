package memory

import (
	"fmt"
	"gameboy-emulator/internal/cartridge"
	"gameboy-emulator/internal/step/gpu"
	"gameboy-emulator/internal/step/interrupts"
	"gameboy-emulator/internal/step/joypad"
	"gameboy-emulator/internal/step/timer"
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

		bootFlag byte // Set to non-zero to disable boot ROM

		interrupts *interrupts.Interrupts
		gpu        *gpu.GPU
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
	gpu *gpu.GPU,
	joypad *joypad.Joypad,
	bootRom *[0x100]byte,
) *Memory {
	m := &Memory{
		interrupts: interrupts,
		gpu:        gpu,
		bootRom:    *bootRom,
	}
	m.initializeIOAddressSpace(
		timer,
		gpu,
		interrupts,
		joypad,
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
}

func (mem *Memory) Read8BitValue(address uint16) byte {
	return mem.read(address)
}

func (mem *Memory) Read16BitValue(address uint16) uint16 {
	return uint16(mem.read(address+1))<<8 | uint16(mem.read(address))
}

func (mem *Memory) Write8BitValue(address uint16, data byte) {
	mem.write(address, data)
}

func (mem *Memory) Write16BitValue(address uint16, data uint16) {
	loByte := uint8(data)
	hiByte := uint8(data >> 8)
	mem.write(address, loByte)
	mem.write(address+0x0001, hiByte)
}

func (mem *Memory) InsertGameCartridge(cart cartridge.Cartridge) {
	mem.cartridge = cart
}

func (mem *Memory) initializeIOAddressSpace(
	timer *timer.Timer,
	gpu *gpu.GPU,
	interrupts *interrupts.Interrupts,
	joypad *joypad.Joypad,
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
	mem.io[0x10] = ioRegister{"NR10", missingWrite, missingRead}
	mem.io[0x11] = ioRegister{"NR11", missingWrite, missingRead}
	mem.io[0x12] = ioRegister{"NR12", missingWrite, missingRead}
	mem.io[0x13] = ioRegister{"NR13", missingWrite, missingRead}
	mem.io[0x14] = ioRegister{"NR14", missingWrite, missingRead}
	mem.io[0x16] = ioRegister{"NR21", missingWrite, missingRead}
	mem.io[0x17] = ioRegister{"NR22", missingWrite, missingRead}
	mem.io[0x18] = ioRegister{"NR23", missingWrite, missingRead}
	mem.io[0x19] = ioRegister{"NR24", missingWrite, missingRead}
	mem.io[0x1A] = ioRegister{"NR30", missingWrite, missingRead}
	mem.io[0x1B] = ioRegister{"NR31", missingWrite, missingRead}
	mem.io[0x1C] = ioRegister{"NR32", missingWrite, missingRead}
	mem.io[0x1D] = ioRegister{"NR33", missingWrite, missingRead}
	mem.io[0x1E] = ioRegister{"NR34", missingWrite, missingRead}
	mem.io[0x20] = ioRegister{"NR41", missingWrite, missingRead}
	mem.io[0x21] = ioRegister{"NR42", missingWrite, missingRead}
	mem.io[0x22] = ioRegister{"NR43", missingWrite, missingRead}
	mem.io[0x23] = ioRegister{"NR44", missingWrite, missingRead}
	mem.io[0x24] = ioRegister{"NR50", missingWrite, missingRead}
	mem.io[0x25] = ioRegister{"NR51", missingWrite, missingRead}
	mem.io[0x26] = ioRegister{"NR52", missingWrite, missingRead}

	// Wave RAM

	// LCD Control, Status, Position, Scrolling and Palettes
	mem.io[0x40] = ioRegister{"LCDC", gpu.SetControl, gpu.GetControl}
	mem.io[0x41] = ioRegister{"STAT", gpu.SetStatus, gpu.GetStatus}
	mem.io[0x42] = ioRegister{"SCY", gpu.SetScrollY, gpu.GetScrollY}
	mem.io[0x43] = ioRegister{"SCX", gpu.SetScrollX, gpu.GetScrollX}
	mem.io[0x44] = ioRegister{"LY", func(_ byte) { gpu.ResetCurrentLine() }, gpu.GetCurrentLine}
	mem.io[0x45] = ioRegister{"LYC", gpu.SetCurrentLineCompare, gpu.GetCurrentLineCompare}
	mem.io[0x46] = ioRegister{"DMA", mem.doDMATransfer, noRead("Tried to read DMA transfer address 0xFF46")}
	mem.io[0x47] = ioRegister{"BGP", gpu.SetBackgroundPalette, gpu.GetBackgroundPalette}
	mem.io[0x48] = ioRegister{"OBP0", gpu.SetObjectPalette0, gpu.GetObjectPalette0}
	mem.io[0x49] = ioRegister{"OBP1", gpu.SetObjectPalette1, gpu.GetObjectPalette1}
	mem.io[0x4A] = ioRegister{"WY", gpu.SetWindowY, gpu.GetWindowY}
	mem.io[0x4B] = ioRegister{"WX", gpu.SetWindowX, gpu.GetWindowX}

	// Boot flag control
	mem.io[0x50] = ioRegister{"BOOT", mem.setBootFlag, noRead("Tried to read boot loader flag")}
}

func (mem *Memory) write(address uint16, data byte) {
	switch {
	case address < 0x8000: // Write to actually read only memory changes banking within cartridge
		if mem.cartridgePresent() {
			mem.cartridge.HandleBanking(address, data)
		}

	case address < 0xA000: // VRAM
		mem.gpu.WriteVRam(address-0x8000, data)

	case address < 0xC000: // External RAM
		if mem.cartridgePresent() {
			mem.cartridge.WriteRAM(address-0xA000, data)
		}

	case address < 0xE000: // WRAM
		mem.wram[address-0xc000] = data

	case address < 0xFE00: // Write to so-called ECHO ram is the same as writing to WRAM (0xc000-0xddff)
		mem.write(address-0x2000, data)

	case address < 0xFEA0: // OAM
		mem.gpu.WriteOAM(address-0xFE00, data)

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

func (mem *Memory) read(address uint16) byte {
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
		return mem.gpu.ReadVRam(address - 0x8000)

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
		return mem.read(address - 0x2000)

	case address < 0xFEA0: // OAM
		return mem.gpu.ReadOAM(address - 0xFE00)

	case address < 0xFF00: // not usable area
		log.L().Panic("Attempt to read from not usable memory area")

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

// doDMATransfer executes a DMA transfer from ROM or RAM to OAM.
// The given value specifies the transfer source address divided by 0x100.
func (mem *Memory) doDMATransfer(value byte) {
	address := uint16(value) << 8 // this is the same as dividing by 0x100
	for i := 0; i < gpu.OAMSize; i++ {
		mem.Write8BitValue(0xFE00+uint16(i), mem.Read8BitValue(address+uint16(i)))
	}
}

func (mem *Memory) setBootFlag(data byte) {
	mem.bootFlag = data
}

func (mem *Memory) readSc() byte {
	return mem.sc
}

func (mem *Memory) writeSc(data byte) {
	mem.sc = data
}

func (mem *Memory) readSb() byte {
	return mem.sb
}

func (mem *Memory) writeSb(data byte) {
	mem.sb = data
}

func noRead(message string) func() byte {
	return func() byte {
		log.L().Panic(message)
		return 0x00
	}
}

func missingRead() byte {
	log.L().Debug("Not implemented yet")
	return 0x00
}

func missingWrite(data byte) {
	log.L().Debug("Not implemented yet")
}
