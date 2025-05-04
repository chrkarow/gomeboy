package gpu

import (
	"gameboy-emulator/internal/util"
)

const (
	spFetchTileNo spriteFetcherState = iota
	spGetTileDataLow
	spGetTileDataHigh
	spPush
)

// Due to technical limitation, the sprite buffer can only contain a maximum of ten sprites, which are drawn on the
// current line
const spriteBufferMaxSize = 10

type (
	spriteFetcherState byte

	sprite struct {
		xPos         int
		yPos         int
		yFlip        bool
		xFlip        bool
		tileIndex    byte
		paletteIndex byte
		bgPriority   bool
	}

	SpritePixel struct {
		colorId      byte
		paletteIndex byte // Only used for sprite pixels
		bgPriority   bool
	}

	SpriteFetcher struct {
		pixelQueue util.Queue[SpritePixel]

		currentSprite  sprite
		currentTileRow [8]byte
		currentTileNo  byte
		tileY          int

		oam [OAMSize]byte // Object Attribute Memory

		spriteSet    [40]*sprite
		spriteBuffer []sprite // Filled during OAM Scan (may contain max 10 sprites)

		ticks         byte
		state         spriteFetcherState
		idle          bool
		lastFetchXPos int

		// Pointers to PPU managed data
		currentLine *byte      // Pointer to current line register
		control     *byte      // Pointer to control register
		tileSet     *[384]tile // Pointer to tile set
		xPos        *byte      // current X position on screen

		bgFetcher *BackgroundFetcher
	}
)

func NewSpriteFetcher(control *byte, currentLine *byte, xPos *byte, tileSet *[384]tile, bgFetcher *BackgroundFetcher) *SpriteFetcher {
	return &SpriteFetcher{
		currentLine:  currentLine,
		control:      control,
		tileSet:      tileSet,
		xPos:         xPos,
		bgFetcher:    bgFetcher,
		oam:          [OAMSize]byte{},
		spriteSet:    [40]*sprite{},
		spriteBuffer: make([]sprite, 0, spriteBufferMaxSize),
	}
}

func (f *SpriteFetcher) OAMScan(spriteIndex byte) {

	// If spriteBuffer is already filled, return
	if len(f.spriteBuffer) == cap(f.spriteBuffer) {
		return
	}

	candidateSprite := f.spriteSet[spriteIndex]
	if candidateSprite == nil {
		return
	}

	if int(*f.currentLine) >= candidateSprite.yPos && int(*f.currentLine) < candidateSprite.yPos+f.spriteYSize() {
		f.spriteBuffer = append(f.spriteBuffer, *candidateSprite)
	}
}

func (f *SpriteFetcher) Start() {
	f.state = spFetchTileNo
	f.idle = true
	f.currentTileNo = 0
	f.currentTileRow = [8]byte{}
	f.lastFetchXPos = -1
	f.pixelQueue.Clear()
}

func (f *SpriteFetcher) Tick() {

	if f.idle {

		// if there was already a fetch for this x position, we don't have to do it again!
		if f.lastFetchXPos == int(*f.xPos) {
			return
		}

		s, found := f.fetchSprite()
		if !found {
			return
		}

		f.currentSprite = s
		f.bgFetcher.SetSuspended(true)
		f.idle = false
	}

	f.ticks++
	if f.state != spPush && f.ticks < 2 { // All states last two ticks, only bgPush will be done on the first tick
		return
	}
	f.ticks = 0

	switch f.state {
	case spFetchTileNo:

		f.currentTileNo, f.tileY = f.getTileNumber(f.currentSprite)

		f.state = spGetTileDataLow

	case spGetTileDataLow:
		f.state = spGetTileDataHigh

	case spGetTileDataHigh:
		f.currentTileRow = f.tileSet[f.currentTileNo][f.tileY]

		f.state = spPush

	case spPush:

		for i := 0; i < 8; i++ {
			xPosOfPixel := f.currentSprite.xPos + i

			// Don't fetch off-screen pixel
			if xPosOfPixel < 0 || xPosOfPixel >= int(ScreenXResolution) {
				continue
			}

			columnWithinTile := i

			// account for xFlip
			if f.currentSprite.xFlip {
				columnWithinTile -= 8 - 1
				columnWithinTile *= -1
			}

			newPixel := SpritePixel{
				colorId:      f.currentTileRow[columnWithinTile],
				paletteIndex: f.currentSprite.paletteIndex,
				bgPriority:   f.currentSprite.bgPriority,
			}

			queueIndex := xPosOfPixel - int(*f.xPos)
			p, err := f.pixelQueue.Peek(queueIndex)

			if err != nil {
				f.pixelQueue.Push(newPixel)
			} else if p.IsTransparent() {
				_ = f.pixelQueue.Set(queueIndex, newPixel)
			}
		}

		f.bgFetcher.SetSuspended(false)
		f.lastFetchXPos = int(*f.xPos)
		f.idle = true
		f.state = spFetchTileNo
	}

}

func (f *SpriteFetcher) HBlank() {
	f.spriteBuffer = make([]sprite, 0, spriteBufferMaxSize)
	f.lastFetchXPos = -1
}

func (f *SpriteFetcher) OutputPixel() *SpritePixel {
	if f.pixelQueue.Size() == 0 {
		return nil
	}

	p, _ := f.pixelQueue.Pop()
	return &p
}

func (f *SpriteFetcher) WriteOAM(address uint16, data byte) {
	f.oam[address] = data
	f.updateSpriteSet(address)
}

func (f *SpriteFetcher) ReadOAM(address uint16) byte {
	return f.oam[address]
}

func (f *SpriteFetcher) updateSpriteSet(address uint16) {
	data := f.oam[address]
	s := f.spriteSet[address/4]

	if s == nil {
		s = &sprite{}
		f.spriteSet[address/4] = s
	}

	switch address % 4 {
	case 0: // Y Position
		s.yPos = int(data) - 16
	case 1: // X Position
		s.xPos = int(data) - 8
	case 2: // Tile index
		s.tileIndex = data
	case 3:
		s.yFlip = util.BitIsSet8(data, 6)
		s.xFlip = util.BitIsSet8(data, 5)
		s.bgPriority = util.BitIsSet8(data, 7)
		s.paletteIndex = data & 0x10 >> 4
	}
}

func (f *SpriteFetcher) spriteYSize() int {
	// Are we using 8x16 pixel sprites instead of 8x8
	use8x16 := util.BitIsSet8(*f.control, 2)
	if use8x16 {
		return 16
	}
	return 8
}

func (f *SpriteFetcher) fetchSprite() (sp sprite, found bool) {
	for _, s := range f.spriteBuffer {
		if s.xPos == int(*f.xPos) || (*f.xPos == 0 && s.xPos <= int(*f.xPos)) {
			return s, true
		}
	}
	return sprite{}, false
}

func (f *SpriteFetcher) getTileNumber(s sprite) (tileNumber byte, tileY int) {
	ySize := f.spriteYSize()

	tileId := s.tileIndex
	// In 8x16 mode hardware ensures, that the tileIndex always points to an even address (top tile).
	// The following odd address contains the bottom tile)
	if ySize == 16 {
		tileId &= 0xFE
	}

	rowOfSprite := int(*f.currentLine) - s.yPos

	// If yFlip is enabled, read the sprite in backwards
	if s.yFlip {
		rowOfSprite -= ySize - 1
		rowOfSprite *= -1
	}

	tileOffset := rowOfSprite / 8

	tileNumber = tileId + byte(tileOffset)
	tileY = rowOfSprite - (tileOffset * 8)

	return
}

func (p *SpritePixel) IsTransparent() bool {
	return p.colorId == 0x0
}
