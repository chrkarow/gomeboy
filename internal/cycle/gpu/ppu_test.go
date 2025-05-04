package gpu

import (
	"gameboy-emulator/internal/cycle/interrupts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

// Control register
const (
	bgWindowEnable byte = 1 << iota
	objEnable
	objSize
	bgTileMap
	bgWindowTiles
	windowEnable
	windowTileMap
	lcdPPUEnable
)

type InterruptsMock struct {
	mock.Mock
}

func (m *InterruptsMock) RequestInterrupt(t interrupts.InterruptType) {
	m.Called(t)
}

func TestPPU_StateTransitions(t *testing.T) {
	// GIVEN
	interruptsMock := &InterruptsMock{}
	ppu := NewPPU(interruptsMock)
	interruptsMock.On("RequestInterrupt", interrupts.LcdStat)
	interruptsMock.On("RequestInterrupt", interrupts.VBlank)

	// WHEN
	ppu.SetStatus(0x38) // Turn all mode interrupts on (except LYC=LY)
	ppu.SetControl(lcdPPUEnable)
	ppu.Tick()

	// THEN
	assert.Equal(t, oamScan, ppuState(ppu.GetStatus()&0x3))
	assert.Equal(t, byte(0), ppu.currentLine)

	interruptsMock.AssertCalled(t, "RequestInterrupt", interrupts.LcdStat)
	interruptsMock.AssertNotCalled(t, "RequestInterrupt", interrupts.VBlank)
	interruptsMock.AssertNumberOfCalls(t, "RequestInterrupt", 1)

	// WHEN
	repeat(79, ppu.Tick)

	// THEN
	assert.Equal(t, pixelTransfer, ppuState(ppu.GetStatus()&0x3))
	assert.Equal(t, byte(0), ppu.currentLine)

	// WHEN
	repeat(172, ppu.Tick) // 172 = 160 (screen width) + 12 (initial load delay of backgroundFetcher)

	// THEN
	assert.Equal(t, hBlank, ppuState(ppu.GetStatus()&0x3))
	assert.Equal(t, byte(0), ppu.currentLine)

	interruptsMock.AssertNotCalled(t, "RequestInterrupt", interrupts.VBlank)
	interruptsMock.AssertNumberOfCalls(t, "RequestInterrupt", 2)

	// WHEN
	repeat(204, ppu.Tick) // to make it 456 ticks

	// THEN
	assert.Equal(t, oamScan, ppuState(ppu.GetStatus()&0x3))
	assert.Equal(t, byte(1), ppu.currentLine)

	interruptsMock.AssertNotCalled(t, "RequestInterrupt", interrupts.VBlank)
	interruptsMock.AssertNumberOfCalls(t, "RequestInterrupt", 3)

	// WHEN
	repeat(65208, ppu.Tick) // = 143 (lines to get to the end of the screen) * 456 (ticks per line)

	// THEN
	assert.Equal(t, vBlank, ppuState(ppu.GetStatus()&0x3))
	assert.Equal(t, byte(144), ppu.currentLine)
	interruptsMock.AssertCalled(t, "RequestInterrupt", interrupts.VBlank)

	// 290 = 2 (LcdStat - oamScan + hBlank) * 144 + 1 (lcdStat - Vblank) + 1 (Vblank)
	interruptsMock.AssertNumberOfCalls(t, "RequestInterrupt", 290)

	// WHEN
	repeat(4560, ppu.Tick) // = 143 (lines to get to the end of the screen) * 456 (ticks per line)

	// THEN
	assert.Equal(t, oamScan, ppuState(ppu.GetStatus()&0x3))
	assert.Equal(t, byte(0), ppu.currentLine)
	interruptsMock.AssertCalled(t, "RequestInterrupt", interrupts.VBlank)
	interruptsMock.AssertNumberOfCalls(t, "RequestInterrupt", 291)
}

func repeat(times int, toRepeat func()) {
	for i := 0; i < times; i++ {
		toRepeat()
	}
}
