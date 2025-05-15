package main

import "fyne.io/fyne/v2"

const (
	paletteKey = "de.cka.gomeboy.settings.palette"
	keyMapKey  = "de.cka.gomeboy.settings.keymap"
)

var defaultKeyMap = []string{
	string(fyne.KeyRight),
	string(fyne.KeyLeft),
	string(fyne.KeyUp),
	string(fyne.KeyDown),
	string(fyne.KeyA),
	string(fyne.KeyB),
	string(fyne.KeyEscape),
	string(fyne.KeyReturn),
}

type Settings struct {
	paletteName string
	keys        []string
	keyMap      map[fyne.KeyName]byte

	preferences fyne.Preferences
}

func NewSettings(pref fyne.Preferences) *Settings {
	s := &Settings{
		preferences: pref,
	}
	s.initialize()
	return s
}

func (s *Settings) initialize() {
	s.keyMap = make(map[fyne.KeyName]byte)
	s.Load()
}

func (s *Settings) GetKeyMap() map[fyne.KeyName]byte {
	return s.keyMap
}

func (s *Settings) GetPaletteName() string {
	return s.paletteName
}

func (s *Settings) Save() {
	s.preferences.SetString(paletteKey, s.paletteName)
	s.preferences.SetStringList(keyMapKey, s.keys)
	s.refreshKeyMap()
}

func (s *Settings) Load() {
	s.paletteName = s.preferences.StringWithFallback(paletteKey, "plainGrayscale")
	s.keys = s.preferences.StringListWithFallback(keyMapKey, defaultKeyMap)
	s.refreshKeyMap()
}

func (s *Settings) refreshKeyMap() {
	for i, str := range s.keys {
		s.keyMap[fyne.KeyName(str)] = byte(i)
	}
}
