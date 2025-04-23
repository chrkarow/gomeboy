package cartridge

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoadCartridge_noMBC(t *testing.T) {
	// GIVEN
	pathToImage := "../../roms/Tetris.gb"

	// WHEN
	cartridge := LoadCartridgeImage(pathToImage)

	// THEN
	assert.IsType(t, &noMBC{
		rom: nil,
		ram: [8192]byte{},
	}, cartridge)
}

func TestLoadCartridge_mbc1(t *testing.T) {
	// GIVEN
	pathToImage := "../../roms/Super Mario Land.gb"

	// WHEN
	cartridge := LoadCartridgeImage(pathToImage)

	// THEN
	assert.IsType(t, &mbc1{
		rom: nil,
	}, cartridge)
}

func TestLoadCartridge_mbc2(t *testing.T) {
	// GIVEN
	pathToImage := "../../roms/Kirbys Pinball Land.gb"

	// WHEN
	cartridge := LoadCartridgeImage(pathToImage)

	// THEN
	assert.IsType(t, &mbc2{
		rom:            nil,
		ram:            [512]byte{},
		currentROMBank: 0,
		ramEnabled:     false,
	}, cartridge)
}
