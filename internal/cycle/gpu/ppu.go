package gpu

import (
	"gameboy-emulator/internal/cycle/interrupts"
	"gameboy-emulator/internal/util"
)

const (
	hBlank ppuState = iota
	vBlank
	oamScan
	pixelTransfer
)

const OAMSize int = 0xA0

type (
	PPU struct {

		// Contains the 384 possible tiles which consist of 8x8 pixels. The contained values are the
		// color indices (possible values 0-3)
		tileSet [384]tile

		vram [0x2000]byte // Video RAM

		control            byte       // LCDC (0xFF40)
		status             byte       // STAT (0xFF41)
		currentLine        byte       // LY (0xFF44)
		currentLineCompare byte       // LYC (0xFF45)
		bgPalette          [4]byte    // BGP (0xFF47) as array for easier access
		objPalettes        [2][4]byte // 0 = OBP0 (0xFF48), 1 = OBP1 (0xFF48) in nested arrays for easier access

		state         ppuState
		display       *Display
		ticks         uint16
		xPos          byte
		pendingWrites util.Queue[func(*PPU)]

		interruptSink     interrupts.InterruptSink
		backgroundFetcher *BackgroundFetcher
		spriteFetcher     *SpriteFetcher

		PrintFrame bool
	}

	ppuState byte

	tile [8][8]byte
)

func NewPPU(interruptSink interrupts.InterruptSink) *PPU {
	p := &PPU{
		display:       NewDisplay(),
		interruptSink: interruptSink,
		pendingWrites: util.Queue[func(*PPU)]{},
	}
	p.Reset()
	return p
}

func (p *PPU) Reset() {

	p.status = 0x80
	p.currentLine = 0
	p.currentLineCompare = 0
	p.bgPalette = [4]byte{}
	p.objPalettes = [2][4]byte{}

	p.backgroundFetcher = NewBackgroundFetcher(&p.control, &p.currentLine, &p.vram, &p.tileSet)
	p.spriteFetcher = NewSpriteFetcher(&p.control, &p.currentLine, &p.xPos, &p.tileSet, p.backgroundFetcher)
	p.display.Reset()
	p.xPos = 0
	p.ticks = 0
	p.state = hBlank
	p.setPPUMode(hBlank)
	p.pendingWrites.Clear()
}

func (p *PPU) Tick() {

	if p.pendingWrites.Size() > 0 {
		defer func() {
			for p.pendingWrites.Size() > 0 {
				w, _ := p.pendingWrites.Pop()
				w(p)
			}
		}()
	}

	p.doCompareLYCAndLC()

	// Update display status based on LCDC bit 7
	if p.display.IsEnabled() {
		if !p.isEnabled() {
			p.display.Disable()
			p.state = hBlank // codeslinger.co.uk says vblankMode
			p.setPPUMode(hBlank)
			p.xPos = 0
			p.currentLine = 0
			p.ticks = 0
			return
		}
	} else {
		if p.isEnabled() {
			p.display.Enable()
			p.transitionToOamScan()

		} else {
			return // display stays turned off, PPU idling
		}
	}

	p.ticks++

	switch p.state {
	case oamScan:
		p.onOAMScan()
	case pixelTransfer:
		p.onPixelTransfer()
	case hBlank:
		p.onHBlank()
	case vBlank:
		p.onVBlank()
	}
}

func (p *PPU) GetDisplay() *Display {
	return p.display
}

// WriteVRam writes the given data to VRAM and at the same time updates the tile set.
//
// VRAM contains 0x2000 addressable bytes and contains the tile data (0x0000 - 0x17FF)
// and the tile maps (map 1: 0x1800 - 0x1BFF, map 2: 0x1C00 - 0x1FFF)
func (p *PPU) WriteVRam(address uint16, data byte) {
	p.pendingWrites.Push(func(ppu *PPU) {
		p.vram[address] = data

		if address < 0x1800 { // when we have written tile data, we have to update the tile set
			p.updateTileSet(address)
		}
	})
}

func (p *PPU) ReadVRam(address uint16) byte {
	return p.vram[address]
}

func (p *PPU) WriteOAM(address uint16, data byte) {
	p.pendingWrites.Push(func(ppu *PPU) {
		p.spriteFetcher.WriteOAM(address, data)
	})
}

func (p *PPU) ReadOAM(address uint16) byte {
	return p.spriteFetcher.ReadOAM(address)
}

func (p *PPU) GetControl() byte {
	return p.control
}

func (p *PPU) SetControl(value byte) {
	p.pendingWrites.Push(func(ppu *PPU) {
		p.control = value
	})
}

func (p *PPU) GetStatus() byte {
	return p.status
}

// SetStatus sets the value of the status register. Only bits 3-6 (starting from 0)
// are writable. Other bits will be ignored.
func (p *PPU) SetStatus(data byte) {
	p.pendingWrites.Push(func(ppu *PPU) {
		oldValue := p.status

		p.status &= 0x87        // turn bits 3-7 off (0x87 = 0b10000111)
		p.status |= data & 0x78 // only take bits 3-6 from data (0x78 = 0b01111000)

		if !p.isEnabled() {
			return
		}

		switch {
		case !util.BitIsSet8(oldValue, 3) && util.BitIsSet8(p.status, 3) && p.state == hBlank:
			p.interruptSink.RequestInterrupt(interrupts.LcdStat)
		case !util.BitIsSet8(oldValue, 4) && util.BitIsSet8(p.status, 4) && p.state == vBlank:
			p.interruptSink.RequestInterrupt(interrupts.LcdStat)
		case !util.BitIsSet8(oldValue, 5) && util.BitIsSet8(p.status, 5) && p.state == oamScan:
			p.interruptSink.RequestInterrupt(interrupts.LcdStat)
		case !util.BitIsSet8(oldValue, 6) && util.BitIsSet8(p.status, 6) && util.BitIsSet8(p.status, 2):
			p.interruptSink.RequestInterrupt(interrupts.LcdStat)
		}
	})
}

func (p *PPU) GetScrollY() byte {
	return p.backgroundFetcher.GetScrollY()
}

func (p *PPU) SetScrollY(data byte) {
	p.pendingWrites.Push(func(ppu *PPU) {
		p.backgroundFetcher.SetScrollY(data)
	})
}

func (p *PPU) GetScrollX() byte {
	return p.backgroundFetcher.GetScrollX()
}

func (p *PPU) SetScrollX(data byte) {
	p.pendingWrites.Push(func(ppu *PPU) {
		p.backgroundFetcher.SetScrollX(data)
	})
}

func (p *PPU) GetCurrentLine() byte {
	return p.currentLine
}

func (p *PPU) GetCurrentLineCompare() byte {
	return p.currentLineCompare
}

func (p *PPU) SetCurrentLineCompare(data byte) {
	p.pendingWrites.Push(func(ppu *PPU) {
		p.currentLineCompare = data
	})
}

// GetBackgroundPalette returns the value as byte by constructing it from the 4 indices of the palette array
func (p *PPU) GetBackgroundPalette() byte {
	return p.bgPalette[3]<<6 | p.bgPalette[2]<<4 | p.bgPalette[1]<<2 | p.bgPalette[0]
}

// SetBackgroundPalette splits the given value directly into the palette array to allow for easier access while rendering
func (p *PPU) SetBackgroundPalette(data byte) {
	p.pendingWrites.Push(func(ppu *PPU) {
		p.bgPalette[3] = data & 0xC0 >> 6
		p.bgPalette[2] = data & 0x30 >> 4
		p.bgPalette[1] = data & 0x0C >> 2
		p.bgPalette[0] = data & 0x03
	})
}

// GetObjectPalette0 returns the value as byte by constructing it from the 4 indices of the palette array
func (p *PPU) GetObjectPalette0() byte {
	return p.objPalettes[0][3]<<6 | p.objPalettes[0][2]<<4 | p.objPalettes[0][1]<<2 | p.objPalettes[0][0]
}

// SetObjectPalette0 splits the given value directly into the palette array to allow for easier access while rendering
func (p *PPU) SetObjectPalette0(data byte) {
	p.pendingWrites.Push(func(ppu *PPU) {
		p.objPalettes[0][3] = data & 0xC0 >> 6
		p.objPalettes[0][2] = data & 0x30 >> 4
		p.objPalettes[0][1] = data & 0x0C >> 2
		p.objPalettes[0][0] = data & 0x03
	})
}

// GetObjectPalette1 returns the value as byte by constructing it from the 4 indices of the palette array
func (p *PPU) GetObjectPalette1() byte {
	return p.objPalettes[1][3]<<6 | p.objPalettes[1][2]<<4 | p.objPalettes[1][1]<<2 | p.objPalettes[1][0]
}

// SetObjectPalette1 splits the given value directly into the palette array to allow for easier access while rendering
func (p *PPU) SetObjectPalette1(data byte) {
	p.pendingWrites.Push(func(ppu *PPU) {
		p.objPalettes[1][3] = data & 0xC0 >> 6
		p.objPalettes[1][2] = data & 0x30 >> 4
		p.objPalettes[1][1] = data & 0x0C >> 2
		p.objPalettes[1][0] = data & 0x03
	})
}

func (p *PPU) GetWindowY() byte {
	return p.backgroundFetcher.GetWindowY()
}

func (p *PPU) SetWindowY(data byte) {
	p.pendingWrites.Push(func(ppu *PPU) {
		p.backgroundFetcher.SetWindowY(data)
	})
}

func (p *PPU) GetWindowX() byte {
	return p.backgroundFetcher.GetWindowX()
}

func (p *PPU) SetWindowX(data byte) {
	p.pendingWrites.Push(func(ppu *PPU) {
		p.backgroundFetcher.SetWindowX(data)
	})
}

// doCompareLYCAndLC compares the LYC and LC register.
//
// The Game Boy constantly compares the value of the LYC and LY registers.
// When both values are identical, the “LYC=LY” flag in the STAT register is set,
// and (if enabled) a STAT interrupt is requested.
//
// Source: https://gbdev.io/pandocs/STAT.html#ff45--lyc-ly-compare
func (p *PPU) doCompareLYCAndLC() {
	p.status &= 0xFB // set LYC == LY flag to 0
	if p.currentLine == p.currentLineCompare {
		p.status |= 0x04

		// LYC == CY Interrupt only at start of scan line
		if p.ticks == 0 && util.BitIsSet8(p.status, 6) {
			p.interruptSink.RequestInterrupt(interrupts.LcdStat)
		}
	}
}

func (p *PPU) setPPUMode(state ppuState) {
	p.status &= 0xFC        // SetBit last two bits to zero
	p.status |= byte(state) // insert mode
}

func (p *PPU) isEnabled() bool {
	return util.BitIsSet8(p.control, 7)
}

func (p *PPU) onOAMScan() {
	// Source: https://hacktix.github.io/GBEDG/ppu/#oam-scan-mode-2
	if (p.ticks-1)%2 == 0 { // Every two ticks
		spriteIndex := byte(p.ticks / 2)
		p.spriteFetcher.OAMScan(spriteIndex)
	}

	if p.ticks == 80 {
		p.transitionToPixelTransfer()
	}
}

func (p *PPU) onPixelTransfer() {
	p.backgroundFetcher.Tick()
	p.spriteFetcher.Tick()
	bgPixel, skip := p.backgroundFetcher.OutputPixel()

	if skip { // No pixel retrieved
		return
	}

	spritePixel := p.spriteFetcher.OutputPixel()

	var color byte
	if util.BitIsSet8(p.control, 0) &&
		(spritePixel == nil || spritePixel.IsTransparent() || (spritePixel.bgPriority && bgPixel != 0x0) || !util.BitIsSet8(p.control, 1)) {
		color = p.bgPalette[bgPixel]

	} else if util.BitIsSet8(p.control, 1) && spritePixel != nil && !spritePixel.IsTransparent() {
		color = p.objPalettes[spritePixel.paletteIndex][spritePixel.colorId]
	}

	p.display.Write(color)

	p.xPos++
	if p.xPos == ScreenXResolution {
		p.transitionToHBlank()
	}
}

func (p *PPU) onHBlank() {
	// wait until the end of scan line and react to it
	if p.ticks == 456 {
		p.currentLine++
		p.ticks = 0

		if p.currentLine == 144 {
			p.transitionToVBlank()
		} else {
			p.transitionToOamScan()
		}
	}
}

func (p *PPU) onVBlank() {
	// Just chilling and let the CPU do some heavy lifting
	if p.ticks == 456 {
		p.currentLine++
		p.ticks = 0
		if p.currentLine == 154 {
			p.currentLine = 0
			p.transitionToOamScan()
		}
	}
}

func (p *PPU) transitionToOamScan() {

	p.setPPUMode(oamScan)
	p.state = oamScan

	if util.BitIsSet8(p.status, 5) {
		p.interruptSink.RequestInterrupt(interrupts.LcdStat)
	}
}

func (p *PPU) transitionToPixelTransfer() {
	p.backgroundFetcher.Start()
	p.spriteFetcher.Start()
	p.setPPUMode(pixelTransfer)
	p.state = pixelTransfer
}

func (p *PPU) transitionToHBlank() {
	p.xPos = 0
	p.setPPUMode(hBlank)
	p.state = hBlank
	p.display.HBlank()
	p.spriteFetcher.HBlank()

	if util.BitIsSet8(p.status, 3) {
		p.interruptSink.RequestInterrupt(interrupts.LcdStat)
	}
}

func (p *PPU) transitionToVBlank() {
	p.setPPUMode(vBlank)
	p.state = vBlank
	p.display.VBlank()
	p.backgroundFetcher.VBlank()

	p.interruptSink.RequestInterrupt(interrupts.VBlank)
	if util.BitIsSet8(p.status, 4) {
		p.interruptSink.RequestInterrupt(interrupts.LcdStat)
	}

	if p.PrintFrame {
		p.display.PrintFrame()
	}
}

// updateTileSet updates the tile set from the current state of VRAM
//
// Source: https://rylev.github.io/DMG-01/public/book/graphics/tile_ram.html
func (p *PPU) updateTileSet(address uint16) {

	// Tiles rows are encoded in two bytes with the first byte always
	// on an even address. Bitwise ANDing the address with 0xffe
	// gives us the address of the first byte.
	// For example: `12 & 0xFFFE == 12` and `13 & 0xFFFE == 12`
	normalizedAddress := address & 0xFFFE

	// First we need to get the two bytes that encode the affected tile row.
	byte1 := p.vram[normalizedAddress]   // least significant bits
	byte2 := p.vram[normalizedAddress+1] // most significant bits

	// A tile is 8 rows tall. Each row is encoded in two bytes. Therefore the index of the tile within the set is
	// its address divided by 16 (whole number division - rest is dropped)
	tileIndex := address / 16

	// Every two bytes is a new row within a tile
	rowIndex := (address % 16) / 2

	// looping over each pixel (column) in the line
	for colIndex := range p.tileSet[tileIndex][rowIndex] {

		mask := byte(1) << (7 - colIndex)
		lsb := byte1 & mask >> (7 - colIndex)
		msb := byte2 & mask >> (7 - colIndex)

		// Setting the color index at the specific pixel
		p.tileSet[tileIndex][rowIndex][colIndex] = (msb << 1) + lsb
	}
}
