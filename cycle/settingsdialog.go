package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type SettingsDialog struct {
	form      *widget.Form
	container *fyne.Container
	popUp     *widget.PopUp
	canvas    fyne.Canvas

	rightButton  *widget.Button
	leftButton   *widget.Button
	upButton     *widget.Button
	downButton   *widget.Button
	aButton      *widget.Button
	bButton      *widget.Button
	selectButton *widget.Button
	startButton  *widget.Button

	keyChangeButton *widget.Button
	keyChangeIndex  int

	settings *Settings
}

func NewSettingsDialog(canvas fyne.Canvas, settings *Settings) *SettingsDialog {
	sd := &SettingsDialog{
		settings: settings,
		canvas:   canvas,
	}
	sd.initialize()
	return sd
}

func (sd *SettingsDialog) initialize() {

	paletteSelect := widget.NewSelectWithData([]string{"plainGrayscale", "gbOriginal"}, binding.BindString(&sd.settings.paletteName))
	paletteSelect.Selected = sd.settings.GetPaletteName()

	paletteGrid := container.New(layout.NewGridLayout(2),
		widget.NewLabel("Color Palette"), paletteSelect,
	)

	sd.rightButton = widget.NewButton(sd.settings.keys[0], sd.activateKeyChange(&sd.rightButton, 0))
	sd.leftButton = widget.NewButton(sd.settings.keys[1], sd.activateKeyChange(&sd.leftButton, 1))
	sd.upButton = widget.NewButton(sd.settings.keys[2], sd.activateKeyChange(&sd.upButton, 2))
	sd.downButton = widget.NewButton(sd.settings.keys[3], sd.activateKeyChange(&sd.downButton, 3))
	sd.aButton = widget.NewButton(sd.settings.keys[4], sd.activateKeyChange(&sd.aButton, 4))
	sd.bButton = widget.NewButton(sd.settings.keys[5], sd.activateKeyChange(&sd.bButton, 5))
	sd.selectButton = widget.NewButton(sd.settings.keys[6], sd.activateKeyChange(&sd.selectButton, 6))
	sd.startButton = widget.NewButton(sd.settings.keys[7], sd.activateKeyChange(&sd.startButton, 7))

	keyGrid := container.New(layout.NewGridLayout(2),
		widget.NewLabel("Right"), sd.rightButton,
		widget.NewLabel("Left"), sd.leftButton,
		widget.NewLabel("Up"), sd.upButton,
		widget.NewLabel("Down"), sd.downButton,
		widget.NewLabel("A"), sd.aButton,
		widget.NewLabel("B"), sd.bButton,
		widget.NewLabel("Select"), sd.selectButton,
		widget.NewLabel("Start"), sd.startButton,
	)

	saveButton := widget.NewButton("Save", sd.onSave)
	saveButton.Importance = widget.HighImportance

	cancelButton := widget.NewButton("Cancel", sd.onCancel)

	buttonGrid := container.New(layout.NewGridLayout(2), cancelButton, saveButton)

	sd.container = container.New(layout.NewVBoxLayout(), paletteGrid, widget.NewSeparator(), keyGrid, widget.NewSeparator(), buttonGrid)

	sd.canvas.SetOnTypedKey(sd.onKeyChange)
}

func (sd *SettingsDialog) Open() {
	sd.popUp = widget.NewModalPopUp(sd.container, sd.canvas)
	sd.popUp.Resize(fyne.NewSize(300, sd.popUp.Size().Width))
	sd.popUp.Show()
}

func (sd *SettingsDialog) onSave() {
	sd.settings.Save()
	sd.popUp.Hide()
}

func (sd *SettingsDialog) onCancel() {
	sd.settings.Load()
	sd.popUp.Hide()
}

func (sd *SettingsDialog) activateKeyChange(b **widget.Button, i int) func() {
	return func() {
		if sd.keyChangeButton != nil {
			return
		}

		sd.keyChangeButton = *b
		sd.keyChangeIndex = i

		sd.keyChangeButton.Text = "Press key..."
		sd.keyChangeButton.Refresh()
	}
}

func (sd *SettingsDialog) onKeyChange(ke *fyne.KeyEvent) {
	if sd.keyChangeButton == nil {
		return
	}

	sd.keyChangeButton.Text = string(ke.Name)
	sd.settings.keys[sd.keyChangeIndex] = string(ke.Name)
	sd.keyChangeButton.Refresh()

	sd.keyChangeButton = nil
	sd.keyChangeIndex = -1
}
