package gpu

import "gameboy-emulator/internal/util"

const (
	bgFetchTileNo backgroundFetcherState = iota
	bgGetTileDataLow
	bgGetTileDataHigh
	bgPush
)

type (
	backgroundFetcherState byte

	BackgroundFetcher struct {
		pixelQueue util.Queue[byte]

		currentTileNo    byte
		currentTile      tile
		currentRowOfTile int

		fetcherX byte // internal count of fetched background tiles per row

		state               backgroundFetcherState
		ticks               int
		resetOnTileDataHigh bool // reset backgroundFetcher to "bgFetchTileNo" on each start of a new scan line in state "bgGetTileDataHigh"
		dequeuedPixelCount  int  // count of dequeued pixel
		skippedPixel        int  // count of skipped pixel at the start of scanline
		bgPixelToSkip       int  // count of background pixels to skip based on SCX%8 (Sub-Tile X positioning)
		drawingWindow       bool // set to true if we are in window drawing mode
		suspended           bool

		wyEqCurrentLine bool // Set to true if at any point in frame WY <= currentLine - reset at VBlank
		windowLineCount int

		// Pointers to PPU managed data
		tileSet     *[384]tile    // Pointer to tile set
		vram        *[0x2000]byte // Pointer to VRAM
		control     *byte         // Pointer to Control Register
		currentLine *byte         // Pointer to Current Line Register

		scrollY byte // SCY (0xFF42)
		scrollX byte // SCX (0xFF43)
		windowY byte // WY (0xFF4A)
		windowX int  // WX (0xFF4B) could be negative because of +7 logic
	}
)

func NewBackgroundFetcher(control *byte, currentLine *byte, vram *[0x2000]byte, tileSet *[384]tile) *BackgroundFetcher {
	return &BackgroundFetcher{
		pixelQueue:      util.Queue[byte]{},
		windowLineCount: -1,
		control:         control,
		currentLine:     currentLine,
		vram:            vram,
		tileSet:         tileSet,
	}
}

func (f *BackgroundFetcher) Start() {
	f.reset()
	f.resetOnTileDataHigh = true
	f.drawingWindow = false
	f.windowCheck()
}

func (f *BackgroundFetcher) Tick() {
	if f.suspended {
		return
	}

	f.ticks++
	if f.state != bgPush && f.ticks < 2 { // All states last two ticks, only bgPush will be tried every tick
		return
	}
	f.ticks = 0

	switch f.state {
	case bgFetchTileNo:

		if f.drawingWindow {
			f.fetchWindowTileNo()
		} else {
			f.fetchBackgroundTileNo()
		}

		f.state = bgGetTileDataLow

	case bgGetTileDataLow: // Nothing to do other than state transition because of tileSet
		f.state = bgGetTileDataHigh

	case bgGetTileDataHigh:
		f.currentTile = f.getTileForBackgroundOrWindow(f.currentTileNo)

		if f.resetOnTileDataHigh {
			f.state = bgFetchTileNo
			f.resetOnTileDataHigh = false
		} else {
			f.state = bgPush
		}
	case bgPush:
		if f.pixelQueue.Size() == 0 {

			for _, colorId := range f.currentTile[f.currentRowOfTile] {
				f.pixelQueue.Push(colorId)
			}

			f.fetcherX++
			f.state = bgFetchTileNo
		}
	}
}

func (f *BackgroundFetcher) VBlank() {
	f.wyEqCurrentLine = false
	f.windowLineCount = -1
}

func (f *BackgroundFetcher) OutputPixel() (pixel byte, skip bool) {
	if f.pixelQueue.Size() == 0 || f.suspended {
		return 0, true
	}

	// Window check done after pushing out or skipping a pixel
	defer f.windowCheck()

	f.dequeuedPixelCount++
	if f.skippedPixel < f.bgPixelToSkip {
		_, _ = f.pixelQueue.Pop()
		f.skippedPixel++
		return 0, true
	}

	p, _ := f.pixelQueue.Pop()
	return p, false
}

func (f *BackgroundFetcher) SetSuspended(suspended bool) {

	if !f.suspended && suspended {
		f.state = bgFetchTileNo
		f.ticks = 0
	}

	f.suspended = suspended
}

func (f *BackgroundFetcher) GetScrollY() byte {
	return f.scrollY
}

func (f *BackgroundFetcher) SetScrollY(data byte) {
	f.scrollY = data
}

func (f *BackgroundFetcher) GetScrollX() byte {
	return f.scrollX
}

func (f *BackgroundFetcher) SetScrollX(data byte) {
	f.scrollX = data
}

func (f *BackgroundFetcher) GetWindowY() byte {
	return f.windowY
}

func (f *BackgroundFetcher) SetWindowY(data byte) {
	f.windowY = data
}

// GetWindowX gets the X Position of the window plus 7.
//
// Source: https://gbdev.io/pandocs/Scrolling.html#ff4aff4b--wy-wx-window-y-position-x-position-plus-7
func (f *BackgroundFetcher) GetWindowX() byte {
	return byte(f.windowX + 7)
}

// SetWindowX sets the X Position of the window.
// Somehow the value is always 7 bigger than the actual position on screen (windowX = 7 equals left most column on screen)
//
// Source: https://gbdev.io/pandocs/Scrolling.html#ff4aff4b--wy-wx-window-y-position-x-position-plus-7
func (f *BackgroundFetcher) SetWindowX(data byte) {
	f.windowX = int(data) - 7
}

func (f *BackgroundFetcher) fetchBackgroundTileNo() {
	var tileMapStartAddress uint16
	if util.BitIsSet8(*f.control, 3) {
		tileMapStartAddress = 0x1C00
	} else {
		tileMapStartAddress = 0x1800
	}

	// mod 256 because the background map has 256 lines and we need the line row to wrap around
	effectiveLine := (int(*f.currentLine) + int(f.scrollY)) % 256
	tileRow := effectiveLine / 8
	f.currentRowOfTile = effectiveLine % 8

	// Same here - mod 32 because of 32 tiles in one row of the background map
	tileCol := ((f.scrollX / 8) + f.fetcherX) % 32
	f.bgPixelToSkip = int(f.scrollX % 8)

	tileAddress := tileMapStartAddress + uint16(tileRow)*32 + uint16(tileCol)
	f.currentTileNo = f.vram[tileAddress]
}

func (f *BackgroundFetcher) fetchWindowTileNo() {
	var tileMapStartAddress uint16
	if util.BitIsSet8(*f.control, 6) {
		tileMapStartAddress = 0x1C00
	} else {
		tileMapStartAddress = 0x1800
	}

	tileRow := f.windowLineCount / 8
	f.currentRowOfTile = f.windowLineCount % 8

	tileCol := f.fetcherX

	tileAddress := tileMapStartAddress + uint16(tileRow)*32 + uint16(tileCol)
	f.currentTileNo = f.vram[tileAddress]
}

func (f *BackgroundFetcher) windowCheck() {
	if f.windowY <= *f.currentLine {
		f.wyEqCurrentLine = true
	}

	if !f.drawingWindow && // we are currently not drawing the window
		util.BitIsSet8(*f.control, 5) && // window enable bit in LCDC is set
		f.wyEqCurrentLine && // at any point in frame WY <= LY
		f.dequeuedPixelCount-f.skippedPixel >= f.windowX { // the effective x position is right of WX

		f.windowLineCount++
		f.drawingWindow = true
		f.reset()
	}
}

func (f *BackgroundFetcher) reset() {
	f.state = bgFetchTileNo
	f.currentTileNo = 0
	f.currentTile = tile{}
	f.fetcherX = 0
	f.dequeuedPixelCount = 0
	f.bgPixelToSkip = 0
	f.skippedPixel = 0
	f.pixelQueue.Clear()
}

func (f *BackgroundFetcher) getTileForBackgroundOrWindow(identifier byte) tile {
	if util.BitIsSet8(*f.control, 4) {
		return f.tileSet[identifier]
	} else {
		signedIdentifier := int(int8(identifier))
		return f.tileSet[256+signedIdentifier]
	}
}
