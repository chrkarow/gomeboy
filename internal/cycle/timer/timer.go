package timer

import (
	"gameboy-emulator/internal/cycle/interrupts"
	"gameboy-emulator/internal/util"
)

// Timer represents the timer and divider registers which are updated in certain frequencies.
//
// Source: https://gbdev.io/pandocs/Timer_and_Divider_Registers.html
type Timer struct {
	tima          byte   // Timer counter
	tma           byte   // Timer modulo
	tac           byte   // Timer control
	systemCounter uint16 // upper 8 bits of this is div

	high            bool // true if selected DIV bit for TIMA was high in previous cycle
	selectedDIVBit  byte // which bit of div does currently increment TIMA
	timerEnabled    bool
	timaReloadDelay byte
	timaReloading   byte

	interrupts *interrupts.Interrupts
}

func New(inter *interrupts.Interrupts) *Timer {
	t := &Timer{
		interrupts: inter,
	}
	t.Reset()
	return t
}

// Reset the timer to initial state.
//
// Values taken from https://github.com/Gekkio/mooneye-test-suite/blob/main/acceptance/boot_hwio-dmgABCmgb.s
func (t *Timer) Reset() {
	t.tima = 0x0
	t.tma = 0x0
	t.tac = 0xF8
	t.systemCounter = 0x0000
	t.timerEnabled = false
}

func (t *Timer) Tick() {

	t.systemCounter++

	if t.timaReloading > 0 {
		t.timaReloading--
	}

	if t.timaReloadDelay > 0 {
		t.timaReloadDelay--
		if t.timaReloadDelay == 0 {
			t.timaReloading = 4
			t.tima = t.tma
			t.interrupts.RequestInterrupt(interrupts.Timer)
		}
	}

	t.doUpdateTimer()

}

func (t *Timer) GetDiv() byte {
	return byte(t.systemCounter >> 8)
}

func (t *Timer) ResetDiv() {
	t.systemCounter = 0x0000
	if t.high {
		t.timerTick()
	}
	t.high = false
}

func (t *Timer) GetTima() byte {
	return t.tima
}

func (t *Timer) SetTima(value byte) {

	// while TIMA is reloading (1 M-Cycle) writes are ignored
	if t.timaReloading > 0 {
		return
	}

	// if a write is done during the reload delay, the reload is canceled
	if t.timaReloadDelay > 0 {
		t.timaReloadDelay = 0
	}

	t.tima = value
}

func (t *Timer) GetTma() byte {
	return t.tma
}

func (t *Timer) SetTma(value byte) {
	if t.timaReloading > 0 {
		t.tima = value
	}
	t.tma = value
}

func (t *Timer) GetTac() byte {
	return t.tac
}

func (t *Timer) SetTac(value byte) {
	t.tac = 0xF8 | value // set bits 3-7 to 1 because they are not implemented

	// Disabling timer while currently selected bit is high, triggers increment
	if t.timerEnabled && !util.BitIsSet8(value, 2) && t.high {
		t.timerTick()
	}

	t.timerEnabled = util.BitIsSet8(value, 2)

	previousSelectedDIVBit := t.selectedDIVBit
	switch t.tac & 0x3 {
	case 1:
		t.selectedDIVBit = 3
	case 2:
		t.selectedDIVBit = 5
	case 3:
		t.selectedDIVBit = 7
	case 0:
		t.selectedDIVBit = 9
	}

	if previousSelectedDIVBit != t.selectedDIVBit && t.timerEnabled {
		t.doUpdateTimer()
	}

}

func (t *Timer) GetSystemCounter() uint16 {
	return t.systemCounter
}

func (t *Timer) doUpdateTimer() {
	previousHigh := t.high
	t.high = util.BitIsSet16(t.systemCounter, t.selectedDIVBit)

	// only reacting to falling edges
	if previousHigh && !t.high && t.timerEnabled {
		t.timerTick()
	}
}

func (t *Timer) timerTick() {
	if t.tima == 0xFF {
		t.tima = 0
		t.timaReloadDelay = 4
	} else {
		t.tima++
	}
}
