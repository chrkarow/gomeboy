package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"gameboy-emulator/internal/cycle"
	"image"
	"image/color"
)

var colors = [4]color.NRGBA{
	{255, 255, 255, 255},
	{192, 192, 192, 255},
	{96, 96, 96, 255},
	{0, 0, 0, 255},
}

var keyMap = map[fyne.KeyName]byte{
	fyne.KeyRight:  0,
	fyne.KeyLeft:   1,
	fyne.KeyUp:     2,
	fyne.KeyDown:   3,
	fyne.KeyA:      4,
	fyne.KeyB:      5,
	fyne.KeyEscape: 6,
	fyne.KeyReturn: 7,
}

type UserInterface struct {
	app     fyne.App
	window  fyne.Window
	display *canvas.Image

	screenContents *image.NRGBA

	openAction  *widget.ToolbarAction
	stopAction  *widget.ToolbarAction
	pauseAction *widget.ToolbarAction
	playAction  *widget.ToolbarAction

	emulator *cycle.Emulator
}

func NewUserInterface(emu *cycle.Emulator) *UserInterface {
	ui := &UserInterface{
		emulator: emu,
	}
	ui.initialize()

	emu.SetScreenHandler(ui.UpdateFrame)

	return ui
}

func (ui *UserInterface) ShowAndRun() {

	ui.window.ShowAndRun()
}

func (ui *UserInterface) initialize() {
	ui.app = app.NewWithID("de.cka.gomeboy")

	ui.window = ui.app.NewWindow("GOmeBoy")
	ui.window.Resize(fyne.NewSize(432, 500))

	ui.openAction = widget.NewToolbarAction(theme.FolderOpenIcon(), ui.handleOpen)

	ui.stopAction = widget.NewToolbarAction(theme.MediaStopIcon(), ui.handleStop)
	ui.stopAction.Disable()

	ui.playAction = widget.NewToolbarAction(theme.MediaPlayIcon(), ui.handlePlay)
	ui.playAction.Disable()

	ui.pauseAction = widget.NewToolbarAction(theme.MediaPauseIcon(), ui.handlePause)
	ui.pauseAction.Disable()

	toolBar := widget.NewToolbar(
		ui.openAction,
		widget.NewToolbarSpacer(),
		ui.playAction,
		ui.pauseAction,
		ui.stopAction,
	)

	ui.screenContents = image.NewNRGBA(image.Rect(0, 0, 160, 144))
	ui.display = canvas.NewImageFromImage(ui.screenContents)

	content := container.NewBorder(toolBar, nil, nil, nil, ui.display)
	ui.window.SetContent(content)

	if deskCanvas, ok := ui.window.Canvas().(desktop.Canvas); ok {
		deskCanvas.SetOnKeyDown(func(e *fyne.KeyEvent) {
			if e.Name == fyne.KeyT {
				ui.emulator.ToggleTurbo()
				return
			}

			if keyIndex, exists := keyMap[e.Name]; exists {
				ui.emulator.KeyPressed(keyIndex)
			}
		})

		deskCanvas.SetOnKeyUp(func(e *fyne.KeyEvent) {
			if keyIndex, exists := keyMap[e.Name]; exists {
				ui.emulator.KeyReleased(keyIndex)
			}
		})
	}

	ui.window.CenterOnScreen()
}

func (ui *UserInterface) handleOpen() {
	w := ui.app.NewWindow("Open ROM Image")
	size := fyne.NewSize(1000, 600)
	w.Resize(size)

	fo := dialog.NewFileOpen(func(f fyne.URIReadCloser, err error) {

		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		if f == nil {
			w.Close()
			return
		}
		selectedFile := f.URI()

		ui.window.SetTitle(selectedFile.Name())

		ui.pauseAction.Enable()
		ui.stopAction.Enable()
		ui.openAction.Disable()

		ui.emulator.InsertCartridge(f.URI().Path())
		ui.emulator.Run()
		w.Close()
	}, w)
	fo.SetFilter(storage.NewExtensionFileFilter([]string{".gb"}))

	fo.Resize(size)
	fo.Show()
	w.Show()
}

func (ui *UserInterface) handleStop() {
	ui.stopAction.Disable()
	ui.pauseAction.Disable()
	ui.playAction.Enable()
	ui.openAction.Enable()

	ui.emulator.Stop()
}

func (ui *UserInterface) handlePause() {
	ui.pauseAction.Disable()
	ui.playAction.Enable()

	ui.emulator.TogglePause()
}

func (ui *UserInterface) handlePlay() {
	ui.playAction.Disable()
	ui.stopAction.Enable()
	ui.pauseAction.Enable()

	if ui.emulator.IsPaused() {
		ui.emulator.TogglePause()
	} else {
		ui.emulator.Run()
	}
}

func (ui *UserInterface) UpdateFrame(screen [144][160]byte) {
	go fyne.DoAndWait(func() {
		for y, line := range screen {
			for x, pixel := range line {
				ui.screenContents.Set(x, y, colors[pixel])
			}
		}

		ui.display.Refresh()
	})
}
