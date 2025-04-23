package cartridge

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMbc2ReadRAM_notEnabled(t *testing.T) {
	// GIVEN
	rom := make([]byte, 0)
	cartridge := newMBC2(&rom)

	// WHEN
	result := cartridge.ReadRAM(0x00)

	// THEN
	assert.Equal(t, byte(0xFF), result)
}

func TestMbc2ReadRAM_outOfBounds(t *testing.T) {
	// GIVEN
	rom := make([]byte, 0)
	cartridge := newMBC2(&rom)

	// WHEN + THEN
	assert.PanicsWithValue(t,
		"Invalid RAM read",
		func() { cartridge.ReadRAM(0x2000) },
	)
}

func TestMbc2WriteRAM_notEnabled(t *testing.T) {
	// GIVEN
	rom := make([]byte, 0)
	cartridge := newMBC2(&rom)

	address := uint16(0x0000)

	// WHEN
	cartridge.WriteRAM(address, 0x3A)

	// WHEN + THEN
	assert.Equal(t, uint8(0x00), cartridge.(*mbc2).ram[address])

}

func TestMbc2WriteRAM_outOfBounds(t *testing.T) {
	// GIVEN
	rom := make([]byte, 0)
	cartridge := newMBC2(&rom)

	// WHEN + THEN
	assert.PanicsWithValue(t,
		"Invalid RAM write attempt",
		func() { cartridge.WriteRAM(0x2000, 0xAA) },
	)
}

func TestMbc2WriteRAM_ReadRAM_HandleBanking(t *testing.T) {

	// GIVEN
	rom := make([]byte, 0)
	cartridge := newMBC2(&rom)

	// WHEN
	cartridge.HandleBanking(0x3000, 0xFA) // Enable RAM
	cartridge.WriteRAM(0x01FF, 0xAB)
	result := cartridge.ReadRAM(0x01FF)

	// RAM reads should roll over
	result2 := cartridge.ReadRAM(0x03FF)
	result3 := cartridge.ReadRAM(0x11FF)

	// THEN
	expectedValue := uint8(0xFB) // upper nibble set to 1
	assert.Equal(t, expectedValue, result)
	assert.Equal(t, expectedValue, result2)
	assert.Equal(t, expectedValue, result3)
}

func TestMbc2ReadRom_bank0(t *testing.T) {
	// GIVEN
	rom := make([]byte, 0x4000)
	cartridge := newMBC2(&rom)

	address := uint16(0x3FFF)
	expectedValue := uint8(0xAB)

	rom[address] = expectedValue

	// WHEN
	result := cartridge.ReadROM(address)

	// THEN
	assert.Equal(t, expectedValue, result)
}

func TestMbc2ReadRom_outOfBounds(t *testing.T) {
	// GIVEN
	rom := make([]byte, 0x8000)
	cartridge := newMBC2(&rom)

	// WHEN + THEN
	assert.PanicsWithValue(t,
		"Invalid ROM read",
		func() { cartridge.ReadROM(0x8000) },
	)
}

func TestMbc2ReadRom_HandleBanking(t *testing.T) {
	// GIVEN
	rom := make([]byte, 0xC000)
	cartridge := newMBC2(&rom)

	readAddress := uint16(0x7FFF)

	// has to be an address < 0x4000 in which the least
	// significant bit of upper address byte is set to 1
	writeAddress := uint16(0x3100)

	expectedValue1 := uint8(0xAB)
	expectedValue2 := uint8(0xBA)

	rom[0x3FFF] = expectedValue1 // victim of "and" by len of memory
	rom[0xBFFF] = expectedValue2

	// WHEN
	result := cartridge.ReadROM(readAddress)
	cartridge.HandleBanking(writeAddress, 0x02)
	result2 := cartridge.ReadROM(readAddress)

	// THEN
	assert.Equal(t, expectedValue1, result)
	assert.Equal(t, expectedValue2, result2)
}

func TestMbc2HandleBanking_romBankNever0(t *testing.T) {
	// GIVEN
	rom := make([]byte, 0x8000)
	cartridge := newMBC2(&rom)

	// has to be an address < 0x4000 in which the least
	// significant bit of upper address byte is set to 1
	writeAddress := uint16(0x3100)

	// WHEN
	cartridge.HandleBanking(writeAddress, 0x00)

	// THEN
	assert.Equal(t, byte(1), cartridge.(*mbc2).currentROMBank)

}
