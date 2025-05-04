package gpu

import "fmt"

const (
	ScreenXResolution byte = 160
	ScreenYResolution byte = 144
)

// Display is actually a screen-sized array and a callback for asynchronously connecting
// the array to some display framework
type Display struct {
	screen      [ScreenYResolution][ScreenXResolution]byte
	enabled     bool
	yPos        byte
	xPos        byte
	frameOutput func([ScreenYResolution][ScreenXResolution]byte)
}

func NewDisplay() *Display {
	d := &Display{
		frameOutput: func(i [144][160]byte) {
			// Do nothing but prevent nil pointers
		},
	}
	d.Reset()
	return d
}

func (d *Display) Reset() {
	d.screen = splashScreen
	d.enabled = false
	d.yPos = 0
	d.xPos = 0
}

func (d *Display) Enable() {
	d.enabled = true
	d.screen = [ScreenYResolution][ScreenXResolution]byte{}
	go d.frameOutput(d.screen)
}

func (d *Display) Disable() {
	d.enabled = false
	d.xPos = 0
	d.yPos = 0

	d.screen = [ScreenYResolution][ScreenXResolution]byte{}
	go d.frameOutput(d.screen)
}

func (d *Display) IsEnabled() bool {
	return d.enabled
}

func (d *Display) Write(color byte) {
	d.screen[d.yPos][d.xPos] = color
	d.xPos++
}

func (d *Display) HBlank() {
	d.yPos++
	d.xPos = 0
}

func (d *Display) VBlank() {
	d.yPos = 0
	go d.frameOutput(d.screen)
}

func (d *Display) RegisterFrameOutputHandler(handler func([ScreenYResolution][ScreenXResolution]byte)) {
	d.frameOutput = handler
	go d.frameOutput(d.screen)
}

func (d *Display) PrintFrame() {
	for _, line := range d.screen {
		for _, p := range line {
			fmt.Print(p)
		}
		fmt.Println()
	}
	fmt.Println()
	fmt.Println()
	fmt.Println()
}
