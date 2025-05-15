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
	"gameboy-emulator/internal/cycle/emulation"
	"image"
	"image/color"
)

var colorPalettes = map[string][4]color.NRGBA{
	"plainGrayscale": {
		{255, 255, 255, 255},
		{192, 192, 192, 255},
		{96, 96, 96, 255},
		{0, 0, 0, 255},
	},
	"gbOriginal": {
		{155, 188, 55, 255},
		{139, 172, 15, 255},
		{48, 98, 48, 255},
		{15, 56, 15, 255},
	},
}

type UserInterface struct {
	app     fyne.App
	window  fyne.Window
	display *canvas.Image

	screenContents *image.NRGBA

	openAction     *widget.ToolbarAction
	stopAction     *widget.ToolbarAction
	pauseAction    *widget.ToolbarAction
	playAction     *widget.ToolbarAction
	muteAction     *widget.ToolbarAction
	settingsAction *widget.ToolbarAction

	driver   emulation.Driver
	settings *Settings
}

func NewUserInterface(driver emulation.Driver) *UserInterface {
	ui := &UserInterface{
		driver: driver,
	}
	ui.initialize()

	driver.GetCore().SetScreenHandler(ui.UpdateFrame)

	return ui
}

func (ui *UserInterface) ShowAndRun() {
	ui.window.ShowAndRun()
}

func (ui *UserInterface) initialize() {
	ui.app = app.NewWithID("de.cka.gomeboy")

	ui.settings = NewSettings(ui.app.Preferences())

	ui.window = ui.app.NewWindow("GOmeBoy")
	ui.window.Resize(fyne.NewSize(432, 500))
	ui.window.SetMaster()
	ui.window.SetPadded(false)

	ui.openAction = widget.NewToolbarAction(theme.FolderOpenIcon(), ui.onOpen)

	ui.muteAction = widget.NewToolbarAction(theme.VolumeMuteIcon(), ui.onMute)
	ui.muteAction.Disable()

	ui.stopAction = widget.NewToolbarAction(theme.MediaStopIcon(), ui.onStop)
	ui.stopAction.Disable()

	ui.playAction = widget.NewToolbarAction(theme.MediaPlayIcon(), ui.onPlay)
	ui.playAction.Disable()

	ui.pauseAction = widget.NewToolbarAction(theme.MediaPauseIcon(), ui.onPause)
	ui.pauseAction.Disable()

	ui.settingsAction = widget.NewToolbarAction(theme.SettingsIcon(), ui.onSettings)

	toolBar := widget.NewToolbar(
		ui.openAction,
		widget.NewToolbarSeparator(),
		ui.playAction,
		ui.pauseAction,
		ui.stopAction,
		widget.NewToolbarSpacer(),
		ui.muteAction,
		ui.settingsAction,
	)

	ui.screenContents = image.NewNRGBA(image.Rect(0, 0, 160, 144))
	ui.display = canvas.NewImageFromImage(ui.screenContents)

	content := container.NewBorder(toolBar, nil, nil, nil, ui.display)
	ui.window.SetContent(content)

	if deskCanvas, ok := ui.window.Canvas().(desktop.Canvas); ok {
		deskCanvas.SetOnKeyDown(func(e *fyne.KeyEvent) {

			if keyIndex, exists := ui.settings.GetKeyMap()[e.Name]; exists {
				ui.driver.GetCore().KeyPressed(keyIndex)
			}
		})

		deskCanvas.SetOnKeyUp(func(e *fyne.KeyEvent) {
			if keyIndex, exists := ui.settings.GetKeyMap()[e.Name]; exists {
				ui.driver.GetCore().KeyReleased(keyIndex)
			}
		})
	}

	ui.window.CenterOnScreen()
}

func (ui *UserInterface) onOpen() {
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
		ui.muteAction.Enable()
		ui.openAction.Disable()
		ui.settingsAction.Disable()

		ui.driver.GetCore().InsertCartridge(f.URI().Path())
		ui.driver.Run()
		w.Close()
	}, w)
	fo.SetFilter(storage.NewExtensionFileFilter([]string{".gb"}))

	fo.Resize(size)
	fo.Show()
	w.Show()
}

func (ui *UserInterface) onStop() {
	ui.stopAction.Disable()
	ui.pauseAction.Disable()
	ui.muteAction.Disable()
	ui.playAction.Enable()
	ui.openAction.Enable()
	ui.settingsAction.Enable()

	ui.driver.Stop()
}

func (ui *UserInterface) onPause() {
	ui.pauseAction.Disable()
	ui.playAction.Enable()
	ui.settingsAction.Enable()

	ui.driver.TogglePause()
}

func (ui *UserInterface) onPlay() {
	ui.playAction.Disable()
	ui.stopAction.Enable()
	ui.pauseAction.Enable()
	ui.muteAction.Enable()
	ui.settingsAction.Disable()

	if ui.driver.IsPaused() {
		ui.driver.TogglePause()
	} else {
		ui.driver.Run()
	}
}

func (ui *UserInterface) onMute() {
	if ui.muteAction.Icon == theme.VolumeMuteIcon() {
		ui.muteAction.SetIcon(theme.VolumeUpIcon())
	} else {
		ui.muteAction.SetIcon(theme.VolumeMuteIcon())
	}
	if d, ok := ui.driver.(*SoundDriver); ok {
		d.ToggleMute()
	}
}

func (ui *UserInterface) UpdateFrame(screen [144][160]byte) {
	go fyne.DoAndWait(func() {
		for y, line := range screen {
			for x, pixel := range line {
				ui.screenContents.Set(x, y, colorPalettes[ui.settings.GetPaletteName()][pixel])
			}
		}

		ui.display.Refresh()
	})
}

func (ui *UserInterface) onSettings() {
	NewSettingsDialog(ui.window.Canvas(), ui.settings).Open()
}
