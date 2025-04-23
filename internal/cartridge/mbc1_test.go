package cartridge

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMbc1ReadRAM_notEnabled(t *testing.T) {
	// GIVEN
	rom := make([]byte, 0)
	cartridge := newMBC1(&rom, 0x2000)

	// WHEN
	result := cartridge.ReadRAM(0x00)

	// THEN
	assert.Equal(t, byte(0xFF), result)
}

func TestMbc1ReadRAM_outOfBounds(t *testing.T) {
	// GIVEN
	rom := make([]byte, 0)
	cartridge := newMBC1(&rom, 0x2000)

	// WHEN + THEN
	assert.PanicsWithValue(t,
		"Invalid RAM read",
		func() { cartridge.ReadRAM(0x2000) },
	)
}

func TestMbc1WriteRAM_notEnabled(t *testing.T) {
	// GIVEN
	rom := make([]byte, 0)
	cartridge := newMBC1(&rom, 0x2000)

	address := uint16(0x0000)

	// WHEN
	cartridge.WriteRAM(address, 0x3A)

	// WHEN + THEN
	assert.Equal(t, uint8(0x00), cartridge.(*mbc1).ram[address])

}

func TestMbc1WriteRAM_outOfBounds(t *testing.T) {
	// GIVEN
	rom := make([]byte, 0x2000)
	cartridge := newMBC1(&rom, 0x2000)

	// WHEN + THEN
	assert.PanicsWithValue(t,
		"Invalid RAM write attempt",
		func() { cartridge.WriteRAM(0x2000, 0xAA) },
	)
}

func TestMbc1ReadROM_lowerRangeMode0(t *testing.T) {

	// GIVEN
	rom := make([]byte, 0xF000)
	cartridge := newMBC1(&rom, 0)

	address := uint16(0x3FFF)
	expectedValue := byte(0xAB)

	rom[0x2FFF] = expectedValue // Scaling because of ROM size smaller than address space with banking

	// WHEN
	cartridge.HandleBanking(0x6000, 0x0) // Set Mode to zero
	result := cartridge.ReadROM(address)

	// THEN
	assert.Equal(t, expectedValue, result)
}

func TestMbc1ReadROM_lowerRangeMode1(t *testing.T) {

	// GIVEN
	rom := make([]byte, 0x100000)
	cartridge := newMBC1(&rom, 0)

	address := uint16(0x3FFF)
	expectedValue := byte(0xAB)

	rom[0x83FFF] = expectedValue

	// WHEN
	cartridge.HandleBanking(0x6000, 0x1) // Set Mode to one
	cartridge.HandleBanking(0x4000, 0x1) // sets bit 5 (from 0) to 1
	result := cartridge.ReadROM(address)

	// THEN
	assert.Equal(t, expectedValue, result)
}

func TestMbc1ReadROM_higherRange(t *testing.T) {

	// GIVEN
	rom := make([]byte, 0x120000)
	cartridge := newMBC1(&rom, 0)

	address := uint16(0x72A7)
	expectedValue := byte(0xAB)

	rom[0x1132A7] = expectedValue

	// WHEN
	cartridge.HandleBanking(0x6000, 0x0) // Set Mode to zero
	cartridge.HandleBanking(0x2000, 0x4) // Write 0b00100 to lower 5 bits of romBank register
	cartridge.HandleBanking(0x4000, 0x2) // sets bit 6 (from 0) to 1
	result := cartridge.ReadROM(address)

	// THEN
	assert.Equal(t, expectedValue, result)
}

// TODO write more MBC1 tests
