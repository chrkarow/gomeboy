package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"gameboy-emulator/internal/gpu"
	"image"
	"image/color"
)

var colors = [4]color.NRGBA{
	{255, 255, 255, 255},
	{192, 192, 192, 255},
	{96, 96, 96, 255},
	{0, 0, 0, 255},
}

type UserInterface struct {
	app     fyne.App
	window  fyne.Window
	display *canvas.Image

	screenContents *image.NRGBA

	gpu *gpu.GPU
}

func NewUserInterface(gpu *gpu.GPU) *UserInterface {
	ui := &UserInterface{
		gpu: gpu,
	}
	ui.initialize()
	return ui
}

func (ui *UserInterface) ShowAndRun() {
	ui.window.ShowAndRun()
}

func (ui *UserInterface) initialize() {
	ui.app = app.New()

	ui.window = ui.app.NewWindow("GameBoy Emulator")
	ui.window.Resize(fyne.NewSize(432, 480))

	ui.screenContents = image.NewNRGBA(image.Rect(0, 0, 160, 144))

	ui.display = canvas.NewImageFromImage(ui.screenContents)
	ui.window.SetContent(ui.display)

}

func (ui *UserInterface) UpdateFrame() {
	fyne.DoAndWait(func() {
		for y, line := range ui.gpu.GetScreen() {
			for x, pixel := range line {
				ui.screenContents.Set(x, y, colors[pixel])
			}
		}

		ui.display.Refresh()
	})
}
