package joypad

import (
	"gameboy-emulator/internal/step/interrupts"
	"gameboy-emulator/internal/util"
)

type Joypad struct {
	state byte

	control byte

	interrupts *interrupts.Interrupts
}

func New(inter *interrupts.Interrupts) *Joypad {
	j := &Joypad{
		interrupts: inter,
	}
	j.Reset()
	return j
}

// Reset the joypad to initial state.
//
// Values taken from https://github.com/Gekkio/mooneye-test-suite/blob/main/acceptance/boot_hwio-dmgABCmgb.s
func (j *Joypad) Reset() {
	j.state = 0xFF
	j.control = 0xC
}

func (j *Joypad) WriteRegister(data byte) {
	j.control = data & 0x30
}

func (j *Joypad) ReadRegister() byte {

	switch j.control {
	case 0x10:
		return j.control | j.state>>4
	case 0x20:
		return j.control | j.state&0xF
	}
	return 0x3F
}

// KeyPressed records the press of a key. Indexes are set up as follows:
//
//	0 = Right
//	1 = Left
//	2 = Up
//	3 = Down
//	4 = A
//	5 = B
//	6 = Select
//	7 = Start
func (j *Joypad) KeyPressed(index byte) {

	if !util.BitIsSet8(j.state, index) {
		return
	}

	util.UnsetBit(&j.state, index)

	if (index >= 4 && j.control == 0x10) || (index < 4 && j.control == 0x20) {
		j.interrupts.RequestInterrupt(interrupts.Joypad)
	}
}

func (j *Joypad) KeyReleased(index byte) {
	util.SetBit(&j.state, index)
}
