package gpu

import (
	"gameboy-emulator/internal/interrupts"
)

const (
	scanlineCounterThreshold int = 456

	hblankMode byte = 0
	vblankMode byte = 1
	oamMode    byte = 2
	vramMode   byte = 3
)

type (
	GPU struct {
		lastUpdateCycles uint64
		scanlineCounter  int

		vram [0x2000]byte

		// Contains the 384 possible tiles which consist of 8x8 pixels. The contained values are the
		// color indices (possible values 0-3)
		tileSet [384][8][8]byte

		screen [144][160]byte

		control            byte    // LCDC (0xFF40)
		status             byte    // STAT (0xFF41)
		scrollY            byte    // SCY (0xFF42)
		scrollX            byte    // SCX (0xFF43)
		currentLine        byte    // LY (0xFF44)
		currentLineCompare byte    // LYC (0xFF45)
		bgPalette          [4]byte // BGP (0xFF47) as array for easier access
		objPalette0        [4]byte // OBP0 (0xFF48) as array for easier access
		objPalette1        [4]byte // OBP1 (0xFF49) as array for easier access
		windowY            byte    // WY (0xFF4A)
		windowX            byte    // WX (0xFF4B)

		interrupts *interrupts.Interrupts
	}
)

func New(inter *interrupts.Interrupts) *GPU {
	return &GPU{
		interrupts:      inter,
		scanlineCounter: scanlineCounterThreshold,
	}
}

func (g *GPU) UpdateDisplay(currentCycles uint64) {
	cyclesSinceLastUpdate := int(currentCycles - g.lastUpdateCycles)
	g.lastUpdateCycles = currentCycles

	g.updateLCDStatus()

	if !g.isLCDEnabled() {
		return
	}

	g.scanlineCounter += cyclesSinceLastUpdate

	// Advance scanline by one and react to change
	if g.scanlineCounter >= scanlineCounterThreshold {
		g.scanlineCounter %= scanlineCounterThreshold

		// Draw the finished scanline
		if g.currentLine < 144 {
			g.drawLine()
		}

		g.currentLine++

		switch {
		case g.currentLine == 144: // Entering Vblank period
			g.interrupts.RequestInterrupt(interrupts.Vblank)
		case g.currentLine > 153:
			g.ResetCurrentLine()
		}
	}
}

// WriteVRam writes the given data to VRAM and at the same time updates the tile set.
//
// VRAM contains 0x2000 addressable bytes and contains the tile data (0x0000 - 0x17FF)
// and the tile maps (map 1: 0x1800 - 0x1BFF, map 2: 0x1C00 - 0x1FFF)
func (g *GPU) WriteVRam(address uint16, data byte) {
	g.vram[address] = data

	if address < 0x1800 { // when we have written tile data, we have to update the tile set
		g.updateTileSet(address)
	}
}

func (g *GPU) ReadVRam(address uint16) byte {
	return g.vram[address]
}

func (g *GPU) GetControl() byte {
	return g.control
}

func (g *GPU) SetControl(data byte) {
	g.control = data
}

func (g *GPU) GetStatus() byte {
	return g.status
}

// SetStatus sets the value of the status register. Only bits 3-6 (starting from 0)
// are writable. Other bits will be ignored.
func (g *GPU) SetStatus(data byte) {
	g.status &= 0x07        // turn bits 3-7 off (0x07 = 0b00000111)
	g.status |= data & 0x78 // only take bits 3-6 from data (0x78 = 0b01111000)
}

func (g *GPU) GetScrollY() byte {
	return g.scrollY
}

func (g *GPU) SetScrollY(data byte) {
	g.scrollY = data
}

func (g *GPU) GetScrollX() byte {
	return g.scrollX
}

func (g *GPU) SetScrollX(data byte) {
	g.scrollX = data
}

func (g *GPU) GetCurrentLine() byte {
	return g.currentLine
}

func (g *GPU) ResetCurrentLine() {
	g.currentLine = 0
}

func (g *GPU) GetCurrentLineCompare() byte {
	return g.currentLineCompare
}

func (g *GPU) SetCurrentLineCompare(data byte) {
	g.currentLineCompare = data
}

// GetBackgroundPalette returns the value as byte by constructing it from the 4 indices of the palette array
func (g *GPU) GetBackgroundPalette() byte {
	return g.bgPalette[3]<<6 | g.bgPalette[2]<<4 | g.bgPalette[1]<<2 | g.bgPalette[0]
}

// SetBackgroundPalette splits the given value directly into the palette array to allow for easier access while rendering
func (g *GPU) SetBackgroundPalette(data byte) {
	g.bgPalette[3] = data & 0xC0 >> 6
	g.bgPalette[2] = data & 0x30 >> 4
	g.bgPalette[1] = data & 0x0C >> 2
	g.bgPalette[0] = data & 0x03
}

// GetObjectPalette0 returns the value as byte by constructing it from the 4 indices of the palette array
func (g *GPU) GetObjectPalette0() byte {
	return g.objPalette0[3]<<6 | g.objPalette0[2]<<4 | g.objPalette0[1]<<2 | g.objPalette0[0]
}

// SetObjectPalette0 splits the given value directly into the palette array to allow for easier access while rendering
func (g *GPU) SetObjectPalette0(data byte) {
	g.objPalette0[3] = data & 0xC0 >> 6
	g.objPalette0[2] = data & 0x30 >> 4
	g.objPalette0[1] = data & 0x0C >> 2
	g.objPalette0[0] = data & 0x03
}

// GetObjectPalette1 returns the value as byte by constructing it from the 4 indices of the palette array
func (g *GPU) GetObjectPalette1() byte {
	return g.objPalette1[3]<<6 | g.objPalette1[2]<<4 | g.objPalette1[1]<<2 | g.objPalette1[0]
}

// SetObjectPalette1 splits the given value directly into the palette array to allow for easier access while rendering
func (g *GPU) SetObjectPalette1(data byte) {
	g.objPalette1[3] = data & 0xC0 >> 6
	g.objPalette1[2] = data & 0x30 >> 4
	g.objPalette1[1] = data & 0x0C >> 2
	g.objPalette1[0] = data & 0x03
}

func (g *GPU) GetWindowY() byte {
	return g.windowY
}

func (g *GPU) SetWindowY(data byte) {
	g.windowY = data
}

// GetWindowX gets the X Position of the window plus 7.
//
// Source: https://gbdev.io/pandocs/Scrolling.html#ff4aff4b--wy-wx-window-y-position-x-position-plus-7
func (g *GPU) GetWindowX() byte {
	return g.windowX + 7
}

// SetWindowX sets the X Position of the window.
// Somehow the value is always 7 bigger than the actual position on screen (windowX = 7 equals left most column on screen)
//
// Source: https://gbdev.io/pandocs/Scrolling.html#ff4aff4b--wy-wx-window-y-position-x-position-plus-7
func (g *GPU) SetWindowX(data byte) {
	g.windowX = data - 7
}

func (g *GPU) GetScreen() [144][160]byte {
	return g.screen
}

func (g *GPU) drawLine() {

	// If first bit of control is set render tiles
	if g.control&0x01 == 0x01 {
		g.renderTiles()
	}

	// if second bit of control is set render sprites
	if g.control&0x02 == 0x02 {

	}
}

func (g *GPU) isLCDEnabled() bool {
	return g.control&0x80 == 0x80
}

func (g *GPU) updateLCDStatus() {

	// If LCD is disabled set mode to 1 and reset scanline counter
	if !g.isLCDEnabled() {
		g.scanlineCounter = 0
		g.setPPUMode(hblankMode) // codeslinger.co.uk says vblankMode

		return
	}

	g.doUpdatePPUMode()

	g.doCompareLYCAndLC()
}

func (g *GPU) doUpdatePPUMode() {
	oldMode := g.getPPUMode()
	var requestInterrupt = false

	if g.currentLine >= 144 {

		// During Vblank the status stays always the same during line rendering
		g.setPPUMode(vblankMode)
		requestInterrupt = g.status&0x10 == 0x10

	} else {

		// When we are not in Vblank the status changes depending on how many cycles we are already rendering
		// the line. (first 80 cycles -> oamMode, following 172 cycles -> vramMode, until end of line (204 cycles) -> hblankMode)
		switch {
		case g.scanlineCounter <= 80:
			g.setPPUMode(oamMode)
			requestInterrupt = g.status&0x20 == 0x20

		case g.scanlineCounter <= 80+172:
			g.setPPUMode(vramMode)
		// There is never an interrupt in vram mode

		default: // rest of the scan line
			g.setPPUMode(hblankMode)
			requestInterrupt = g.status&0x08 == 0x08
		}
	}

	if requestInterrupt && oldMode != g.getPPUMode() {
		g.interrupts.RequestInterrupt(interrupts.LcdStat)
	}
}

// doCompareLYCAndLC compares the LYC and LC register.
//
// The Game Boy constantly compares the value of the LYC and LY registers.
// When both values are identical, the “LYC=LY” flag in the STAT register is set,
// and (if enabled) a STAT interrupt is requested.
//
// Source: https://gbdev.io/pandocs/STAT.html#ff45--lyc-ly-compare
func (g *GPU) doCompareLYCAndLC() {
	g.status &= 0xFB // set LYC == LY flag to 0
	if g.currentLine == g.currentLineCompare {
		g.status |= 0x04

		if g.status&0x40 == 0x40 {
			g.interrupts.RequestInterrupt(interrupts.LcdStat)
		}
	}
}

func (g *GPU) getPPUMode() byte {
	return g.status & 0x03 // PPU mode is encoded in the last two bits of the status
}

func (g *GPU) setPPUMode(mode byte) {
	g.status &= 0xFC // Set last two bits to zero
	g.status |= mode // insert mode
}

// updateTileSet updates the tile set from the current state of VRAM
//
// Source: https://rylev.github.io/DMG-01/public/book/graphics/tile_ram.html
func (g *GPU) updateTileSet(address uint16) {

	// Tiles rows are encoded in two bytes with the first byte always
	// on an even address. Bitwise ANDing the address with 0xffe
	// gives us the address of the first byte.
	// For example: `12 & 0xFFFE == 12` and `13 & 0xFFFE == 12`
	normalizedAddress := address & 0xFFFE

	// First we need to get the two bytes that encode the affected tile row.
	byte1 := g.vram[normalizedAddress]   // least significant bits
	byte2 := g.vram[normalizedAddress+1] // most significant bits

	// A tile is 8 rows tall. Each row is encoded in two bytes. Therefore the index of the tile within the set is
	// its address divided by 16 (whole number division - rest is dropped)
	tileIndex := address / 16

	// Every two bytes is a new row within a tile
	rowIndex := (address % 16) / 2

	// looping over each pixel (column) in the line
	for colIndex := range g.tileSet[tileIndex][rowIndex] {

		mask := byte(1) << (7 - colIndex)
		lsb := byte1 & mask >> (7 - colIndex)
		msb := byte2 & mask >> (7 - colIndex)

		// Setting the color index at the specific pixel
		g.tileSet[tileIndex][rowIndex][colIndex] = (msb << 1) + lsb
	}
}

func (g *GPU) renderTiles() {

	// is window enabled (bit 5 (counting from 0) of control set to 1) ad are we currently drawing it?
	drawingWindow := g.control&0x20 == 0x20 && g.windowY <= g.currentLine

	// Which tile map are we using for the current line
	var tileMapStartAddress uint16
	if (g.control&0x08 == 0x08) || // for the screen check bit 3
		(drawingWindow && g.control&0x40 == 0x40) { // but while drawing the window, check bit 6
		tileMapStartAddress = 0x1C00
	} else {
		tileMapStartAddress = 0x1800
	}

	// Calculate which line of the background map or the window we are currently drawing
	var yPos byte
	if !drawingWindow {
		yPos = g.scrollY + g.currentLine
	} else {
		yPos = g.currentLine - g.windowY
	}
	// Translate this value to the tile row (one tile is 8 rows high; full number division)
	tileRow := yPos / 8

	for pixel := range g.screen[g.currentLine] {
		xPos := byte(pixel) + g.scrollX

		// Translate current X position to window coordinates (if we are currently drawing the window)
		if drawingWindow && byte(pixel) >= g.windowX {
			xPos = byte(pixel) - g.windowX
		}

		// Translate this value to the tile column (one tile is 8 pixels wide; full number division)
		tileCol := xPos / 8

		tileAddress := tileMapStartAddress + uint16(tileRow)*32 + uint16(tileCol)
		tileIdentifier := g.vram[tileAddress]

		tile := g.getTile(tileIdentifier)
		g.screen[g.currentLine][pixel] = tile[yPos%8][xPos%8]
	}

}

func (g *GPU) getTile(identifier byte) [8][8]byte {
	if g.control&0x10 == 0x10 {
		return g.tileSet[identifier]
	} else {
		signedIdentifier := int(int8(identifier))
		return g.tileSet[256+signedIdentifier]
	}
}
