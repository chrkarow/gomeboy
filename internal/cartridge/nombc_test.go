package cartridge

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNoMBCReadRom_outOfBounds(t *testing.T) {
	// GIVEN
	rom := make([]byte, 0x8000)
	cartridge := newNoMBC(&rom)

	// WHEN + THEN
	assert.PanicsWithValue(t,
		"Invalid ROM read",
		func() { cartridge.ReadROM(0x8000) },
	)
}

func TestNoMBCReadRom(t *testing.T) {
	// GIVEN
	rom := make([]byte, 0x8000)
	cartridge := newNoMBC(&rom)

	address := uint16(0x5670)
	expectedValue := uint8(0xDE)

	rom[address] = expectedValue

	// WHEN
	result := cartridge.ReadROM(address)

	// THEN
	assert.Equal(t, expectedValue, result)
}
