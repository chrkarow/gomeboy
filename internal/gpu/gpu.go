package gpu

import (
	"gameboy-emulator/internal/bit"
	"gameboy-emulator/internal/interrupts"
	"sort"
)

const (
	OAMSize int = 0xA0

	screenXResolution        int = 160
	scanlineCounterThreshold int = 456

	hblankMode byte = 0
	vblankMode byte = 1
	oamMode    byte = 2
	vramMode   byte = 3
)

type (
	GPU struct {
		scanlineCounter int

		vram [0x2000]byte  // Video RAM
		oam  [OAMSize]byte // Object Attribute Memory

		// Contains the 384 possible tiles which consist of 8x8 pixels. The contained values are the
		// color indices (possible values 0-3)
		tileSet   [384][8][8]byte
		spriteSet [40]sprite
		screen    [144][screenXResolution]byte

		control            byte       // LCDC (0xFF40)
		status             byte       // STAT (0xFF41)
		scrollY            byte       // SCY (0xFF42)
		scrollX            byte       // SCX (0xFF43)
		currentLine        byte       // LY (0xFF44)
		currentLineCompare byte       // LYC (0xFF45)
		bgPalette          [4]byte    // BGP (0xFF47) as array for easier access
		objPalettes        [2][4]byte // 0 = OBP0 (0xFF48), 1 = OBP1 (0xFF48) in nested arrays for easier access
		windowY            byte       // WY (0xFF4A)
		windowX            byte       // WX (0xFF4B)

		interrupts *interrupts.Interrupts
		drawScreen func([144][160]byte)
	}

	sprite struct {
		xPos         int
		yPos         int
		yFlip        bool
		xFlip        bool
		tileIndex    byte
		paletteIndex byte
	}
)

func New(inter *interrupts.Interrupts) *GPU {
	g := &GPU{
		interrupts: inter,
		drawScreen: func(screen [144][160]byte) {
			// do nothing
		},
	}
	g.Reset()
	return g
}

// Reset the gpu to initial state.
//
// Values taken from https://github.com/Gekkio/mooneye-test-suite/blob/main/acceptance/boot_hwio-dmgABCmgb.s
func (g *GPU) Reset() {
	g.scanlineCounter = scanlineCounterThreshold

	g.vram = [0x2000]byte{}
	g.oam = [OAMSize]byte{}

	g.tileSet = [384][8][8]byte{}
	g.spriteSet = [40]sprite{}
	g.screen = splashScreen

	g.control = 0x91
	g.status = 0x80
	g.scrollY = 0x0
	g.scrollX = 0x0
	g.currentLine = 0x0A
	g.currentLineCompare = 0x0
	g.bgPalette = [4]byte{0x3, 0x3, 0x3, 0x0}
	g.objPalettes = [2][4]byte{}
	g.windowX = 0x0
	g.windowY = 0x0

	g.drawScreen(g.screen)
}

func (g *GPU) SetScreenHandler(handler func([144][160]byte)) {
	g.drawScreen = handler
	g.drawScreen(g.screen)
}

func (g *GPU) UpdateDisplay(stepCycles int) {
	g.updateLCDStatus()

	if !g.isLCDEnabled() {
		return
	}

	g.scanlineCounter += stepCycles

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
			g.drawScreen(g.screen)
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

func (g *GPU) WriteOAM(address uint16, data byte) {
	g.oam[address] = data
	g.updateSpriteSet(address)
}

func (g *GPU) ReadOAM(address uint16) byte {
	return g.oam[address]
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
	return g.objPalettes[0][3]<<6 | g.objPalettes[0][2]<<4 | g.objPalettes[0][1]<<2 | g.objPalettes[0][0]
}

// SetObjectPalette0 splits the given value directly into the palette array to allow for easier access while rendering
func (g *GPU) SetObjectPalette0(data byte) {
	g.objPalettes[0][3] = data & 0xC0 >> 6
	g.objPalettes[0][2] = data & 0x30 >> 4
	g.objPalettes[0][1] = data & 0x0C >> 2
	g.objPalettes[0][0] = data & 0x03
}

// GetObjectPalette1 returns the value as byte by constructing it from the 4 indices of the palette array
func (g *GPU) GetObjectPalette1() byte {
	return g.objPalettes[1][3]<<6 | g.objPalettes[1][2]<<4 | g.objPalettes[1][1]<<2 | g.objPalettes[1][0]
}

// SetObjectPalette1 splits the given value directly into the palette array to allow for easier access while rendering
func (g *GPU) SetObjectPalette1(data byte) {
	g.objPalettes[1][3] = data & 0xC0 >> 6
	g.objPalettes[1][2] = data & 0x30 >> 4
	g.objPalettes[1][1] = data & 0x0C >> 2
	g.objPalettes[1][0] = data & 0x03
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

func (g *GPU) drawLine() {

	// If first bit of control is set render tiles
	if bit.IsSet8(g.control, 0) {
		g.renderTiles()
	}

	// if second bit of control is set render sprites
	if bit.IsSet8(g.control, 1) {
		g.renderSprites()
	}
}

func (g *GPU) isLCDEnabled() bool {
	return bit.IsSet8(g.control, 7)
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
		requestInterrupt = bit.IsSet8(g.status, 4)

	} else {

		// When we are not in Vblank the status changes depending on how many cycles we are already rendering
		// the line. (first 80 cycles -> oamMode, following 172 cycles -> vramMode, until end of line (204 cycles) -> hblankMode)
		switch {
		case g.scanlineCounter <= 80:
			g.setPPUMode(oamMode)
			requestInterrupt = bit.IsSet8(g.status, 5)

		case g.scanlineCounter <= 80+172:
			g.setPPUMode(vramMode)
		// There is never an interrupt in vram mode

		default: // rest of the scan line
			g.setPPUMode(hblankMode)
			requestInterrupt = bit.IsSet8(g.status, 3)
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

		if bit.IsSet8(g.status, 6) {
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

func (g *GPU) updateSpriteSet(address uint16) {
	data := g.oam[address]
	s := &g.spriteSet[address/4]

	switch address % 4 {
	case 0: // Y Position
		s.yPos = int(data) - 16
	case 1: // X Position
		s.xPos = int(data) - 8
	case 2: // Tile index
		s.tileIndex = data
	case 3:
		s.yFlip = bit.IsSet8(data, 6)
		s.xFlip = bit.IsSet8(data, 5)
		s.paletteIndex = data & 0x10 >> 4
	}
}

func (g *GPU) renderTiles() {

	// is window enabled (bit 5 (counting from 0) of control set to 1) ad are we currently drawing it?
	drawingWindow := bit.IsSet8(g.control, 5) && g.windowY <= g.currentLine

	// Which tile map are we using for the current line
	var tileMapStartAddress uint16
	if bit.IsSet8(g.control, 3) || // for the screen check bit 3
		(drawingWindow && bit.IsSet8(g.control, 6)) { // but while drawing the window, check bit 6
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

		tile := g.getTileForBackgroundOrWindow(tileIdentifier)
		g.screen[g.currentLine][pixel] = g.bgPalette[tile[yPos%8][xPos%8]]
	}

}

func (g *GPU) getTileForBackgroundOrWindow(identifier byte) [8][8]byte {
	if bit.IsSet8(g.control, 4) {
		return g.tileSet[identifier]
	} else {
		signedIdentifier := int(int8(identifier))
		return g.tileSet[256+signedIdentifier]
	}
}

// renderSprites draws the sprite data on screen.
func (g *GPU) renderSprites() {

	// Are we using 8x16 pixel sprites instead of 8x8
	use8x16 := bit.IsSet8(g.control, 2)
	var ySize int
	if use8x16 {
		ySize = 16
	} else {
		ySize = 8
	}

	// Select the sprites to draw based on their y-Positions
	spritesToDraw := make([]sprite, 0, 10)
	for _, s := range g.spriteSet {
		if int(g.currentLine) >= s.yPos && int(g.currentLine) < s.yPos+ySize {
			spritesToDraw = append(spritesToDraw, s)
		}

		// Hardware restrictions allow maximum ten sprites drawn on one line
		if len(spritesToDraw) == 10 {
			break
		}
	}

	// Sort sprites descending based on their x Position (we want to draw higher x Positions first)
	sort.Slice(spritesToDraw, func(i, j int) bool {
		return spritesToDraw[i].xPos > spritesToDraw[j].xPos
	})

	for _, s := range spritesToDraw {

		lineWithinSprite := int(g.currentLine) - s.yPos

		tileIdentifier := s.tileIndex
		// In 8x16 mode hardware ensures, that the tileIndex always points to an even address (top tile).
		// The following odd address contains the bottom tile)
		if use8x16 {
			tileIdentifier &= 0xFE
		}

		// If yFlip is enabled read the sprite in backwards
		if s.yFlip {
			lineWithinSprite -= ySize + 1
			lineWithinSprite *= -1
		}

		tileOffset := lineWithinSprite / 8

		tile := g.tileSet[int(tileIdentifier)+tileOffset]

		lineWithinTile := lineWithinSprite - (tileOffset * 8)

		for i := 0; i < 8; i++ {
			columnWithinTile := i

			// account for xFlip
			if s.xFlip {
				columnWithinTile -= 8 - 1
				columnWithinTile *= -1
			}

			colorIndexOfPixel := tile[lineWithinTile][columnWithinTile]

			pixel := s.xPos + i
			// X bound check - don't draw the sprite when we are out of bounds
			if pixel < 0 || pixel >= screenXResolution {
				return
			}

			// colorId = 0 means transparent for sprites and thus does not affect the screen
			if colorIndexOfPixel != 0 {
				g.screen[g.currentLine][pixel] = g.objPalettes[s.paletteIndex][colorIndexOfPixel]
			}
		}
	}
}
